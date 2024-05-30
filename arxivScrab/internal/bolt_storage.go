package internal

import (
	"net/url"

	"github.com/gocolly/colly/v2/storage"
)

type BoltStorage struct {

}

var _ storage.Storage = (*BoltStorage)(nil)

func (bolt *BoltStorage) Init() error {
	return nil
}

func (bolt *BoltStorage) Close() error {

	return nil
}

func (bolt *BoltStorage) Visited(requestId uint64) error {
	return nil
}

func (bolt *BoltStorage) IsVisited(requestId uint64) (bool, error) {
	return false, nil
}

func (bolt *BoltStorage) Cookies(u *url.URL) string {
	return "";	
}

func (bolt *BoltStorage) SetCookies(u *url.URL, cookies string) {

}