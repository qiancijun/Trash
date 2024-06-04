package main

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/gocolly/colly/v2"
	"github.com/qiancijun/Trash/arxivScrab/types"
)

// https://arxiv.org/search/?query=HyperGraph&searchtype=all&source=header&start=50
var (
	arxivUrl = "https://arxiv.org/search/?query=HyperGraph&searchtype=all&source=header"
)

func main2() {
	c := colly.NewCollector(
		colly.AllowedDomains("arxiv.org"),
	)
	c.OnHTML("ol li.arxiv-result", func(h *colly.HTMLElement) {
		var arxiv types.ArxivItem
		if err := h.Unmarshal(&arxiv); err != nil {
			fmt.Printf("unmarshal arxiv failed: %v", err)
			return
		}
		abstracts := strings.Split(strings.TrimSpace(arxiv.Abstract), "\n")
		arxiv.Abstract = abstracts[0]
		dates := strings.Split(strings.TrimSpace(arxiv.Date), "\n")
		arxiv.Date = dates[0]

		fmt.Println(sonic.MarshalString(arxiv))
	})
	if err := c.Visit(arxivUrl); err != nil {
		fmt.Println(err)
	}
}
