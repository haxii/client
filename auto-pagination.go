package client

type SinglePagination func(currPage int) (nextPage int, err error)

// AutoPagination auto pagination util next page is non-positive or error happens
func AutoPagination(startPage int, p SinglePagination) error {
	page := startPage
	for {
		nextPage, err := p(page)
		if err != nil {
			return err
		}
		page = nextPage
		if page < 0 {
			return nil
		}
	}
}
