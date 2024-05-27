package test

import (
	"testing"

	kvdb "github.com/qiancijun/trash/searchEngine/internal/kv_db"
	"github.com/qiancijun/trash/searchEngine/util"
	"github.com/stretchr/testify/assert"
)

var (
	db kvdb.IKeyValueDB 
	setup func(t *testing.T) // 初始化工作
	teardown func() // 销毁工作
)

func init() {
	setup = func(t *testing.T) {
        db, err := kvdb.GetKvDB(kvdb.BOLT, util.RootPath + "data/bolt_db")
		assert.NoError(t, err)
		assert.NotNil(t, db)
    }
	teardown = func() {
		db.Close()
	}
}

func TestBolt(t *testing.T) {
	var err error
	db, err = kvdb.GetKvDB(kvdb.BOLT, util.RootPath + "data/bolt_db")
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestBadger(t *testing.T) {
	var err error
	db, err = kvdb.GetKvDB(kvdb.BADGER, util.RootPath + "data/badger_db")
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestGetDbPath(t *testing.T) {
	db, err := kvdb.GetKvDB(kvdb.BOLT, util.RootPath + "data/bolt_db")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer func() {
		db.Close()
	}()
	p := db.GetDbPath()
	assert.Equal(t, util.RootPath + "data/bolt_db", p)
}
