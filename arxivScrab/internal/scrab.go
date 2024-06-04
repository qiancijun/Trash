package internal

import (
	"fmt"
	"log"
	"strings"
	"sync"

	// "github.com/bytedance/sonic"
	"github.com/gocolly/colly/v2"
	"github.com/qiancijun/Trash/arxivScrab/types"
	"github.com/qiancijun/Trash/arxivScrab/util"
	"github.com/vbauerster/mpb/v8"
	"gorm.io/gorm"
)

const baseUrl = "https://arxiv.org/search/"

type ArxivScrab struct {
	collector     *colly.Collector
	sqlite        *gorm.DB
	keywords      []string
	searchType    string
	domains       []string
	emailsMapping map[string]struct{} // email delivery mapping
	url           string
	bar           *mpb.Bar // fetch progress bar

	Wg sync.WaitGroup
}

func GetArxivScrab() (*ArxivScrab, error) {
	sqliteDb, err := NewSqlite3(util.RootPath + "data/data.sqlite")
	if err != nil {
		return nil, err
	}
	scrab := new(ArxivScrab)
	scrab.sqlite = sqliteDb
	scrab.Wg = sync.WaitGroup{}
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

func (s *ArxivScrab) WithBar(bar *mpb.Bar) *ArxivScrab {
	s.bar = bar
	return s
}

func (s *ArxivScrab) WithEmails(emails ...string) *ArxivScrab {
	for _, email := range emails {
		s.emailsMapping[email] = struct{}{}
	}
	return s
}

func (s *ArxivScrab) Init() error {
	s.collector = colly.NewCollector(
		colly.AllowedDomains(s.domains...),
		colly.Async(true),
	)
	s.collector.Limit(&colly.LimitRule{
		Parallelism: 3,
	})
	// 不需要本地缓存，每天爬取最新的都是同一个 url
	// if err := s.collector.SetStorage(&SqliteStorage{}); err != nil {
	// 	return err
	// }

	s.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("colly has error: %v", err)
		s.Wg.Done()
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
		if s.bar != nil {
			s.bar.Increment()
		}
		s.sqlite.Create(arxiv)
	})
	s.collector.OnScraped(func(r *colly.Response) {
		s.Wg.Done()
	})
	return nil
}

func (s *ArxivScrab) Run(offset int) error {
	// 构造 URL
	keywords := strings.Join(s.keywords, " ")
	s.url = fmt.Sprintf("%s?query=%s&searchtype=%s&source=header&start=%d", baseUrl, keywords, s.searchType, offset)
	log.Printf("fetch url %s", s.url)
	return s.collector.Visit(s.url)
}

func (s *ArxivScrab) Wait() {
	s.collector.Wait()
	if s.bar != nil {
		s.bar.Wait()
	}
	// wait all data fetch finished
	// send emails
}

func (s *ArxivScrab) Clean() {
	// TODO delete db file
}
