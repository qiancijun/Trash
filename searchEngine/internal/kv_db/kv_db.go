package kvdb

import (
	"os"
	"strings"

	"github.com/qiancijun/trash/searchEngine/util"
)

const (
	BOLT = iota
	BADGER
)

type IKeyValueDB interface {
	Open() error
	GetDbPath() string // 获取存储数据的目录
	Set(k, v []byte) error
	BatchSet(keys, values [][]byte) error
	Get(k []byte) ([]byte, error)
	BatchGet(keys [][]byte) ([][]byte, error)
	Delete(k []byte) error
	BatchDelete(keys [][]byte) error
	Has(k []byte) bool
	IterDB(fn func(k, v []byte) error) int64 // 遍历数据库，返回数据条数
	IterKey(fn func(k []byte) error) int64 // 遍历所有 key 
	Close() error
}

func GetKvDB(dbtype int, path string) (IKeyValueDB, error) {
	paths := strings.Split(path, "/")
	parentPath := strings.Join(paths[0:len(paths)-1], "/")

	info, err := os.Stat(parentPath)
	if os.IsNotExist(err) {
		util.Log.Printf("create dir %s", parentPath)
		os.MkdirAll(parentPath, os.ModePerm)
	} else {
		if info.Mode().IsRegular() {
			util.Log.Printf("%s is a regular file, will delete it", parentPath)
			os.Remove(parentPath)
		}
	}

	var db IKeyValueDB
	switch dbtype {
	case BADGER:
		db = new(Badger).WithDataPath(path)
	default:
		db = new(Bolt).WithDataPath(path).WithBucket("cheryl")
	}
	err = db.Open()
	return db, err
}