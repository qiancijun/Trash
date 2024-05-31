package internal

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2/storage"
	bolt "go.etcd.io/bbolt"
)

type BoltStorage struct {
	db     *bolt.DB
	path   string
	bucket []byte
}

var (
	_ storage.Storage = (*BoltStorage)(nil)
	ErrNoData = errors.New("no data")
)

func GetBoltStorage(path string) (*BoltStorage, error) {
	paths := strings.Split(path, "/")
	parentPath := strings.Join(paths[0:len(paths)-1], "/")

	info, err := os.Stat(parentPath)
	if os.IsNotExist(err) {
		log.Printf("create dir %s", parentPath)
		os.MkdirAll(parentPath, os.ModePerm)
	} else {
		if info.Mode().IsRegular() {
			log.Printf("%s is a regular file, will delete it", parentPath)
			os.Remove(parentPath)
		}
	}
	db := new(BoltStorage).WithDataPath(path).WithBucket("arxiv")
	err = db.Init()
	return db, err
}

func (b *BoltStorage) WithDataPath(path string) *BoltStorage {
	b.path = path
	return b
}

func (b *BoltStorage) WithBucket(bucket string) *BoltStorage {
	b.bucket = []byte(bucket)
	return b
}

func (b *BoltStorage) Set(k, v []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(b.bucket).Put(k, v)
	})
}

func (b *BoltStorage) Get(k []byte) ([]byte, error) {
	var ival []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		ival = tx.Bucket(b.bucket).Get(k)
		return nil
	})
	if len(ival) == 0 {
		return nil, ErrNoData
	}
	return ival, err
}

func (b *BoltStorage) Init() error {
	dataDir := b.path
	db, err := bolt.Open(dataDir, 0o600, bolt.DefaultOptions)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(b.bucket)
		return err
	})
	if err != nil {
		db.Close()
		return err
	} else {
		b.db = db
		return nil
	}
}

func (b *BoltStorage) Close() error {
	return b.db.Close()
}

func (b *BoltStorage) Visited(requestId uint64) error {
	key := strconv.Itoa(int(requestId))
	return b.Set([]byte(key), []byte(key))
}

func (b *BoltStorage) IsVisited(requestId uint64) (bool, error) {
	key := strconv.Itoa(int(requestId))
	_, err := b.Get([]byte(key))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *BoltStorage) Cookies(u *url.URL) string {
	key := []byte(u.String())
	if val, err := b.Get(key); err == nil {
		return string(val)
	}
	return ""
}

func (b *BoltStorage) SetCookies(u *url.URL, cookies string) {
	key := []byte(u.String())
	b.Set(key, []byte(cookies))
}
