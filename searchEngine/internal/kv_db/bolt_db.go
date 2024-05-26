package kvdb

import (
	"errors"

	"github.com/boltdb/bolt"
)

var ErrNoData = errors.New("no")

type Bolt struct {
	db *bolt.DB
	path string
	bucket []byte
}

func (s *Bolt) WithDataPath(path string) *Bolt {
	s.path = path
	return s
}
