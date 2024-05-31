package test

import (
	"testing"

	"github.com/qiancijun/Trash/arxivScrab/internal"
	"github.com/qiancijun/Trash/arxivScrab/util"
	"github.com/stretchr/testify/assert"
)

func TestBoltStorage(t *testing.T) {
	db, err := internal.GetBoltStorage(util.RootPath + "data/scrab.bolt")	
	assert.NoError(t, err)
	assert.NotNil(t, db)
}