package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"sync"
	"time"

	// "github.com/bytedance/sonic"
	"github.com/gocolly/colly/v2"
	"github.com/qiancijun/Trash/arxivScrab/types"
	"github.com/qiancijun/Trash/arxivScrab/util"
	"github.com/vbauerster/mpb/v8"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

const (
	baseUrl     = "https://arxiv.org/search/"
	htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Daily Arxiv</title>
</head>
<body>
	<table>
		{{ range . }}
			<tr class="item">
				<!-- https://arxiv.org/abs/2406.02514 -->
				<td><a href="https://arxiv.org/abs/{{.ArxivLink | GetTag}}">{{ .ArxivLink }}</a></td>
				<td>{{ .Title }}</td>
				<td>{{ .Authors | Concat }}</td>
				<td>{{ .Abstract }}</td>
				<td>{{ .Date }}</td>
				<td>{{ .Comments }}</td>
			</tr>
		{{ end }}
	</table>
</body>
</html>

<style>
	table {
		width: 100%;
		border-collapse: collapse;
	}

	table, th, td {
		border: 1px solid black;
	}

	th, td {
		padding: 8px;
		text-align: left;
	}

	/* 奇数行背景颜色 */
	tr:nth-child(odd) {
		background-color: #f2f2f2;
	}

	/* 偶数行背景颜色 */
	tr:nth-child(even) {
		background-color: #ffffff;
	}

	th {
		background-color: #4CAF50;
		color: white;
	}
</style>
	`
)

type ArxivScrab struct {
	collector  *colly.Collector
	sqlite     *gorm.DB
	keywords   []string
	searchType string
	domains    []string
	emails     []string // email delivery mapping
	url        string
	bar        *mpb.Bar // fetch progress bar
	cache      []types.ArxivItem

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
	scrab.cache = make([]types.ArxivItem, 0)
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
	if len(emails) > 0 {
		s.emails = append(s.emails, emails...)
	}
	return s
}

func (s *ArxivScrab) Init() error {
	s.collector = colly.NewCollector(
		colly.AllowedDomains(s.domains...),
		colly.Async(true),
	)
	s.collector.Limit(&colly.LimitRule{
		Parallelism: 1,
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
		tx := s.sqlite.Create(arxiv)
		if tx.RowsAffected > 0 {
			s.cache = append(s.cache, arxiv)
		}
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
	s.Wg.Add(1)
	return s.collector.Visit(s.url)
}

func (s *ArxivScrab) Wait() {
	s.Wg.Wait()
	if s.bar != nil {
		s.bar.Wait()
	}
	// wait all data fetch finished
	// send emails
	emailWg := sync.WaitGroup{}
	for _, k := range s.emails {
		emailWg.Add(1)
		go func(email string) {
			if err := s.sendEmail(email, &emailWg); err != nil {
				log.Printf("send email fail: %s", err)
			}
		}(k)
	}
	emailWg.Wait()
}

func (s *ArxivScrab) Clean() {
	// TODO delete db file
}

func (s *ArxivScrab) renderHTML() (*bytes.Buffer, error) {
	funcMap := template.FuncMap{
		"Concat": func(strs []string) string {
			return strings.Join(strs, ", ")
		},
		"GetTag": func(arixvLink string) string {
			return strings.Split(arixvLink, ":")[1]
		},
	}
	data := s.cache
	if len(data) == 0 {
		return nil, fmt.Errorf("no data")
	}
	t, err := template.New("template.html").Funcs(funcMap).Parse(htmlContent)
	if err != nil {
		return nil, err
	}
	var htmlBody bytes.Buffer
	err = t.Execute(&htmlBody, data)
	if err != nil {
		return nil, err
	}
	return &htmlBody, nil
}

func (s *ArxivScrab) sendEmail(dest string, wg *sync.WaitGroup) error {
	defer wg.Done()
	buffer, err := s.renderHTML()
	if err != nil {
		log.Printf("render html has error: %s", err)
		return err
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "769303522@qq.com")
	m.SetHeader("To", dest)

	// 获取当前时间
	currentTime := time.Now()
	currentDate := currentTime.Format("2006-01-02")

	m.SetHeader("Subject", fmt.Sprintf("[%s]Daily Arxiv: Keywords %v", currentDate, s.keywords))
	m.SetBody("text/html", buffer.String())

	d := gomail.NewDialer("smtp.qq.com", 587, "769303522@qq.com", "umlzqqguzjbnbcig")

	if err := d.DialAndSend(m); err != nil {
		log.Printf("error sending email: %v", err)
		return err
	}
	return nil
}
