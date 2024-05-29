package test

import (
	"testing"

	"github.com/qiancijun/trash/searchEngine/index_service"
	"github.com/stretchr/testify/assert"
)

var (
	serviceName = "test_service"
)

func TestGetServiceEndpoints(t *testing.T) {
	hub := index_service.GetServiceHub([]string{"127.0.0.1:2379"}, 3)
	assert.NotNil(t, hub)

	endpoint := "127.0.0.1:5000"
	_, err := hub.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer hub.UnRegist(serviceName, endpoint)


	endpoint = "127.0.0.2:5000"
	_, err = hub.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer hub.UnRegist(serviceName, endpoint)

	endpoint = "127.0.0.3:5000"
	_, err = hub.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer hub.UnRegist(serviceName, endpoint)

	endpoints := hub.GetServiceEndpoints(serviceName)
	assert.Equal(t, []string{
		"127.0.0.1:5000",
		"127.0.0.2:5000",
		"127.0.0.3:5000",
	}, endpoints)
}