package internal

import (
	"fmt"
	"log"
	"strings"

	// "github.com/bytedance/sonic"
	"github.com/gocolly/colly/v2"
	"github.com/qiancijun/Trash/arxivScrab/types"
	"github.com/qiancijun/Trash/arxivScrab/util"
	"gorm.io/gorm"
)

const baseUrl = "https://arxiv.org/search/"

type ArxivScrab struct {
	collector  *colly.Collector
	db         *BoltStorage // 用于用作本地缓存
	sqlite     *gorm.DB
	keywords   []string
	searchType string
	domains    []string
	url        string
}

func GetArxivScrab() (*ArxivScrab, error) {
	boltDb, err := GetBoltStorage(util.RootPath + "data/scrab.bolt")
	if err != nil {
		return nil, err
	}
	sqliteDb, err := NewSqlite3(util.RootPath + "data/data.sqlite")
	if err != nil {
		return nil, err
	}
	scrab := new(ArxivScrab)
	scrab.db = boltDb
	scrab.sqlite = sqliteDb
	return scrab, nil
}

func (s *ArxivScrab) WithKeywords(word ...string) *ArxivScrab {
	if len(word) > 0 {
		s.keywords = append(s.keywords, word...)
	}
	return s
}

func (s *ArxivScrab) WithSearchType(t string) *ArxivScrab {
	s.searchType = t
	return s
}

func (s *ArxivScrab) WithDomains(domains ...string) *ArxivScrab {
	if len(domains) > 0 {
		s.domains = append(s.domains, domains...)
	}
	return s
}

func (s *ArxivScrab) Init() error {
	// 为 colly 初始化本地缓存数据库

	s.collector = colly.NewCollector(
		colly.AllowedDomains(s.domains...),
	)
	if err := s.collector.SetStorage(s.db); err != nil {
		return err
	}

	s.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("colly has error: %v", err)
	})
	s.collector.OnHTML("ol li.arxiv-result", func(h *colly.HTMLElement) {
		var arxiv types.ArxivItem
		if err := h.Unmarshal(&arxiv); err != nil {
			log.Printf("unmarshal arxiv failed: %v", err)
			return
		}
		abstracts := strings.Split(strings.TrimSpace(arxiv.Abstract), "\n")
		arxiv.Abstract = abstracts[0]
		dates := strings.Split(strings.TrimSpace(arxiv.Date), "\n")
		arxiv.Date = dates[0]
		// if str, err := sonic.MarshalString(arxiv); err != nil {
		// 	log.Printf("unmarshal arxiv failed: %s", err)
		// } else {
		// 	log.Printf("arxiv item: %s", str)
		// }
		// 写入本地数据库
		if tx := s.sqlite.Create(arxiv); tx.Error != nil {
			log.Printf("write to sqlite error: %v", tx.Error)
		}
	})
	return nil
}

func (s *ArxivScrab) Run() error {
	// 构造 URL
	keywords := strings.Join(s.keywords, " ")
	s.url = fmt.Sprintf("%s?query=%s&searchtype=%s&source=header", baseUrl, keywords, s.searchType)
	log.Printf("fetch url %s", s.url)
	return s.collector.Visit(s.url)
}

func (s *ArxivScrab) Close() {
	s.db.Close()
}
