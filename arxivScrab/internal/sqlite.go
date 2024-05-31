package internal

import (
	"fmt"

	"github.com/qiancijun/Trash/arxivScrab/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 用于保存爬取数据

func NewSqlite3(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	// 自动创建表
	if !tableCheck(db, "arxiv_items") {
		db.AutoMigrate(&types.ArxivItem{})
	}
	return db, nil
}

func tableCheck(db *gorm.DB, schema string) bool {
	tx := db.Exec(fmt.Sprintf("SELECT COUNT(1) FROM %s", schema))
	if tx.Error == nil {
		return true
	} else {
		return false
	}
}