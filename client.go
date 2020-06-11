package client

import (
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
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

// NewHTTPClient makes a http client for future usage
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

var ErrServerFailure = errors.New("server error")

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
	// get raw page string from server
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return -1, nil, nil, err
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
	res, err := p.client.Do(req)
	if err != nil {
		return -1, nil, nil, err
	}
	defer res.Body.Close()
	pageSource, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Clone(), nil, err
	}
	return res.StatusCode, res.Header.Clone(), pageSource, nil
}
