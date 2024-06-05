package internal

import (
	"github.com/qiancijun/Trash/arxivScrab/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 用于保存爬取数据

func NewSqlite3(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	// 自动创建表
	if !db.Migrator().HasTable(&types.ArxivItem{}) {
		if err := db.AutoMigrate(&types.ArxivItem{}); err != nil {
			return nil, err
		}
	}
	return db, nil
}