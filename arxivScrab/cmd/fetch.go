package cmd

import (
	"fmt"

	"github.com/qiancijun/Trash/arxivScrab/internal"
	"github.com/spf13/cobra"
)

var (
	keywords   []string
	domains    []string
	emails     []string
	searchType string
	page       int
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch daily data",
	Long:  "fetch daily arxiv data to email",
	Run: func(cmd *cobra.Command, args []string) {
		// 创建一个进度条
		scrab, err := internal.GetArxivScrab()
		if err != nil {
			fmt.Printf("GetArxivScrab error: %v\n", err)
			return
		}
		scrab.WithKeywords(keywords...).WithDomains(domains...).WithSearchType(searchType).WithEmails(emails...)
		if err != nil {
			fmt.Printf("Init error: %v\n", err)
			return
		}

		for i := 0; i < page; i++ {
			scrab.Run((i - 1) * 50)
		}
		// wait async fetch finish
		scrab.Wait()
	},
}

func init() {
	fetchCmd.Flags().StringArrayVarP(&keywords, "keywords", "k", []string{}, "指定论文关键字")
	fetchCmd.Flags().StringArrayVarP(&domains, "domains", "", []string{"arxiv.org"}, "指定爬虫域名，一般不需要额外填写")
	fetchCmd.Flags().StringVarP(&searchType, "search-type", "t", "all", "指定搜索类型")
	fetchCmd.Flags().IntVarP(&page, "page", "p", 1, "指定页数")
	fetchCmd.Flags().StringArrayVarP(&emails, "emails", "e", []string{}, "emails")
}
