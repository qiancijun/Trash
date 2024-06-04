package cmd

import (
	"fmt"

	"github.com/qiancijun/Trash/arxivScrab/internal"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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
		pb := mpb.New(mpb.WithWidth(64))
		bar := pb.New(
			int64(page*50),
			mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
			mpb.PrependDecorators(
				// display our name with one space on the right
				decor.Name("Fetch Progress", decor.WC{C: decor.DindentRight | decor.DextraSpace}),
				// replace ETA decorator with "done" message, OnComplete event
				decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done"),
			),
			mpb.AppendDecorators(decor.Percentage()),
		)

		scrab, err := internal.GetArxivScrab()
		if err != nil {
			fmt.Printf("GetArxivScrab error: %v\n", err)
			return
		}
		scrab.WithKeywords(keywords...).WithDomains(domains...).WithSearchType(searchType).WithBar(bar)
		err = scrab.Init()
		if err != nil {
			fmt.Printf("Init error: %v\n", err)
			return
		}

		for i := 0; i < page; i++ {
			scrab.Wg.Add(1)
			scrab.Run(i * 50)
		}
		// wait async fetch finish
		scrab.Wg.Wait()
		scrab.Wait()
	},
}

func init() {
	fetchCmd.Flags().StringArrayVarP(&keywords, "keywords", "k", []string{}, "指定论文关键字")
	fetchCmd.Flags().StringArrayVarP(&domains, "domains", "", []string{"arxiv.org"}, "指定爬虫域名，一般不需要额外填写")
	fetchCmd.Flags().StringVarP(&searchType, "search-type", "t", "all", "指定搜索类型")
	fetchCmd.Flags().IntVarP(&page, "page", "p", 0, "指定页数")
	fetchCmd.Flags().StringArrayVarP(&emails, "emails", "e", []string{}, "emails")
}
