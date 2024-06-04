package test

import (
	"testing"

	"github.com/qiancijun/Trash/arxivScrab/internal"
	"github.com/stretchr/testify/assert"
)

func initArxivScrab(t *testing.T) *internal.ArxivScrab {
	arxivScrab, err := internal.GetArxivScrab()
	assert.NoError(t, err)
	assert.NotNil(t, arxivScrab)
	arxivScrab.WithDomains("arxiv.org").WithSearchType("all").WithKeywords("Graph")
	err = arxivScrab.Init()
	assert.NoError(t, err)
	return arxivScrab
}

func TestArxivInit(t *testing.T) {
	arxivScrab, err := internal.GetArxivScrab()
	assert.NoError(t, err)
	assert.NotNil(t, arxivScrab)
	arxivScrab.WithDomains("arxiv.org").WithSearchType("all").WithKeywords("Graph")
	err = arxivScrab.Init()
	assert.NoError(t, err)
}

func TestArxivRun(t *testing.T) {
	scrab := initArxivScrab(t)
	err := scrab.Run(0)
	assert.NoError(t, err)
}

// 测试多次爬取一个 URL，结果不重复
func TestMultiRun(t *testing.T) {
	scrab := initArxivScrab(t)
	err := scrab.Run(0)
	assert.NoError(t, err)

	err = scrab.Run(0)
	assert.Error(t, err, "URL already visited")
}