package client

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPClient makes an http client for future usage
func NewHTTPClient(debugProxy *url.URL, headers map[string]string,
	dialTimeout, fullTimeout time.Duration, maxConn int) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 2 * maxConn,
		MaxConnsPerHost:     maxConn,
		DialContext:         (&net.Dialer{Timeout: dialTimeout}).DialContext,
		Proxy:               func(request *http.Request) (url *url.URL, e error) { return debugProxy, nil },
	}
	if debugProxy != nil {
		// for proxy debug only
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &HTTPClient{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: transport,
			Timeout:   fullTimeout,
		},
		headers: headers,
	}
}

func (p *HTTPClient) Do(method, url, body string, extraHeaders map[string]string,
	retryTimes int) (statusCode int, header http.Header, respBody []byte, err error) {
	if retryTimes < 0 {
		retryTimes = 0
	}
	for i := 0; i < retryTimes+1; i++ {
		statusCode, header, respBody, err = p.do(method, url, body, extraHeaders)
		if err != nil {
			if strings.Contains(err.Error(), "broken pipe") {
				// do nothing just let it iterate
			} else if strings.Contains(err.Error(), "connection reset by peer") {
				// do nothing just let it iterate
			} else if strings.Contains(err.Error(), "Client.Timeout") {
				// do nothing just let it iterate
			} else {
				return
			}
		} else if statusCode/100 == 5 {
			err = ErrServerFailure
		} else {
			return
		}
	}
	return
}

func (p *HTTPClient) do(method, url, body string, extraHeaders map[string]string) (int, http.Header, []byte, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = strings.NewReader(body)
	}
	res, err := p._do(method, url, bodyReader, extraHeaders)
	if err != nil {
		return -1, nil, nil, err
	}
	defer res.Body.Close()
	pageSource, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Clone(), nil, err
	}
	return res.StatusCode, res.Header.Clone(), pageSource, nil
}

func (p *HTTPClient) _do(method, url string, bodyReader io.Reader, extraHeaders map[string]string) (*http.Response, error) {
	// get raw page string from server
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	// set default headers
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	if extraHeaders != nil {
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}
	}
	return p.client.Do(req)
}

func (p *HTTPClient) DoStream(method, url string, body io.Reader, extraHeaders map[string]string) (statusCode int, header http.Header, respBody io.ReadCloser, err error) {
	res, err := p._do(method, url, body, extraHeaders)
	if err != nil {
		return -1, nil, nil, err
	}
	return res.StatusCode, res.Header.Clone(), res.Body, err
}
