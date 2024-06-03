package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type StringSlice []string

// 实现 driver.Valuer 接口，定义如何将 StringSlice 转换为数据库值
func (s StringSlice) Value() (driver.Value, error) {
    return strings.Join(s, ","), nil
}

// 实现 sql.Scanner 接口，定义如何将数据库值转换为 StringSlice
func (s *StringSlice) Scan(value interface{}) error {
    if value == nil {
        *s = nil
        return nil
    }
    str, ok := value.(string)
    if !ok {
        return fmt.Errorf("failed to scan StringSlice: %v", value)
    }
    *s = strings.Split(str, ",")
    return nil
}

type ArxivItem struct {
	ArxivLink string `json:"arxiv_link" selector:"div.is-marginless > p.list-title > a" gorm:"primaryKey"`
	Title string `json:"title" selector:"p.title"`
	Authors StringSlice `json:"authors" selector:"p.authors a" gorm:"type:text"`
	Abstract string `json:"abstract" selector:"p.abstract > span.abstract-full"`
	Date string `json:"date" selector:"p:nth-child(5)"`
	Comments string `json:"comments" selector:"p.comments span:nth-child(2)"`
}