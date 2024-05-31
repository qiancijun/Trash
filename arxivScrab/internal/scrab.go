package internal

import "github.com/gocolly/colly/v2"

const baseUrl = "https://arxiv.org/search/"

type ArxivScrab struct {
	collector colly.Collector
	keywords []string
	searchType string
	domains string
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

func (s *ArxivScrab) WithDomains() *ArxivScrab {
	
}

func (s *ArxivScrab) Init() error {
	s.collector.
}