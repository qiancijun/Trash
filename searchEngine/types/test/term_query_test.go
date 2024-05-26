package test

import (
	"testing"

	"github.com/qiancijun/trash/searchEngine/types"
	"github.com/stretchr/testify/assert"
)

const FIELD = ""

func TestTermQuery(t *testing.T) {
	A := types.NewTermQuery(FIELD, "")
	B := types.NewTermQuery(FIELD, "B")
	C := types.NewTermQuery(FIELD, "C")
	D := types.NewTermQuery(FIELD, "D")
	E := &types.TermQuery{}
	F := types.NewTermQuery(FIELD, "F")
	G := types.NewTermQuery(FIELD, "G")
	H := types.NewTermQuery(FIELD, "H")

	var q *types.TermQuery
	cases := []struct {
		name string
		operator func() *types.TermQuery
		expect string
	} {
		{
			"空表达式",
			func() *types.TermQuery { return A },
			"",
		}, {
			"简单表达式",
			func() *types.TermQuery { return B.Or(C) },
			"(\001B|\001C)",
		}, {
			"复杂表达式",
			func() *types.TermQuery { return A.Or(B).Or(C).And(D).Or(E).And(F.Or(G)).And(H) },
			"(((((\001B)|\001C)&\001D)&(\001F|\001G))&\001H)",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q = c.operator()
			assert.Equal(t, q.ToString(), c.expect)
		})
	}
}