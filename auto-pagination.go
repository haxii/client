package client

import "time"

type SinglePagination func(currPage int) (nextPage int, err error)

func AutoPagination(startPage int, p SinglePagination) error {
	return AutoPaginationWithTPS(startPage, p, 0)
}

// AutoPaginationWithTPS auto pagination util next page is non-positive or error happens
// tps 限流使用, 每秒最多的请求数量, 小于等于 0 时不限流
func AutoPaginationWithTPS(startPage int, p SinglePagination, tps int) error {
	minDuration := time.Duration(0)
	if tps > 0 {
		minDuration = time.Second / time.Duration(tps)
	}
	page := startPage
	for {
		startTime := time.Now()
		nextPage, err := p(page)
		if err != nil {
			return err
		} else if nextPage <= 0 {
			return nil
		}
		if tps > 0 && page != startPage { // 第一次跳过
			duration := time.Now().Sub(startTime)
			time.Sleep(minDuration - duration)
		}
		page = nextPage
	}
}
