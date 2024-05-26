package kvdb

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

func GetKvDB() (IKeyValueDB, error) {
	return nil, nil
}