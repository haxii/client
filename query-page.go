package client

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

type QueryAllPageConfig struct {
	StartPage         int                  // 起始页码
	PageSize          int                  // 每页拿多少
	PageParamName     string               // 页码字段在 url query / json body 中的名字
	PageSizeParamName string               // 每页返回的数量在 url query / json body 中的名字
	ListPath          string               // list 数据在返回的 response 中的 path
	DetailExtractPath string               // list 中单条数据内容转换，为空则不转换
	TotalPath         string               // total 数据在返回的 response 中的 path
	MinDurationFunc   func() time.Duration // 每次请求的间隔时间
	MaxCount          int                  // 最大获取数量
}

func DefaultQueryAllPageConfig() *QueryAllPageConfig {
	return &QueryAllPageConfig{ // 默认 config 请勿修改
		StartPage:         1,
		PageSize:          10,
		PageParamName:     "page",
		PageSizeParamName: "pageSize",
		ListPath:          "data",
		TotalPath:         "total",
		MinDurationFunc: func() time.Duration { // 1~2s 一次
			return time.Second + time.Duration(rand.Intn(1000))*time.Millisecond
		},
	}
}

func (c *QueryAllPageConfig) CheckErr() error {
	if c.StartPage < 0 {
		return errors.New("invalid start page")
	} else if len(c.PageParamName) == 0 {
		return errors.New("invalid page param name")
	} else if len(c.PageSizeParamName) == 0 {
		return errors.New("invalid page size param name")
	} else if len(c.ListPath) == 0 {
		return errors.New("invalid list path")
	} else if len(c.TotalPath) == 0 {
		return errors.New("invalid total path")
	}
	return nil
}

func (c *QueryAllPageConfig) WithPageSize(v int) *QueryAllPageConfig {
	c.PageSize = v
	return c
}

func (c *QueryAllPageConfig) WithStartPage(v int) *QueryAllPageConfig {
	c.StartPage = v
	return c
}

func (c *QueryAllPageConfig) WithPageSizeParamName(v string) *QueryAllPageConfig {
	c.PageSizeParamName = v
	return c
}

func (c *QueryAllPageConfig) WithPageParamName(v string) *QueryAllPageConfig {
	c.PageParamName = v
	return c
}

func (c *QueryAllPageConfig) WithListPath(v string) *QueryAllPageConfig {
	c.ListPath = v
	return c
}

func (c *QueryAllPageConfig) WithTotalPath(v string) *QueryAllPageConfig {
	c.TotalPath = v
	return c
}

func (c *QueryAllPageConfig) WithMinDurationFunc(v func() time.Duration) *QueryAllPageConfig {
	c.MinDurationFunc = v
	return c
}

func (c *QueryAllPageConfig) WithMaxCount(v int) *QueryAllPageConfig {
	c.MaxCount = v
	return c
}

func (c *QueryAllPageConfig) MakeURL(sourceURL string, page int) (string, error) {
	urlObj, err := url.Parse(sourceURL)
	if err != nil {
		return "", err
	}
	query := urlObj.Query()
	query.Set(c.PageParamName, strconv.Itoa(page))
	query.Set(c.PageSizeParamName, strconv.Itoa(c.PageSize))
	urlObj.RawQuery = query.Encode()
	return urlObj.String(), nil
}

func (c *QueryAllPageConfig) MakeJSONBody(jsonBody []byte, page int) ([]byte, error) {
	req := make(map[string]json.RawMessage, 0)
	err := json.Unmarshal(jsonBody, &req)
	if err != nil {
		return nil, err
	}
	req[c.PageParamName] = []byte(strconv.Itoa(page))
	req[c.PageSizeParamName] = []byte(strconv.Itoa(c.PageSize))
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
