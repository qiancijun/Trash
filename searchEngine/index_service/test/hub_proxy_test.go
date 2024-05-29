package test

import (
	"testing"

	"github.com/qiancijun/trash/searchEngine/index_service"
	"github.com/stretchr/testify/assert"
)

func TestGetServiceEndpointsByProxy(t *testing.T) {
	const qps = 10
	proxy := index_service.GetServiceHubProxy([]string{"127.0.0.1:2379"}, 3, qps)

	assert.NotNil(t, proxy)
	
	endpoint := "127.0.0.1:5000"
	_, err := proxy.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer proxy.UnRegist(serviceName, endpoint)

	endpoint = "127.0.0.2:5000"
	_, err = proxy.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer proxy.UnRegist(serviceName, endpoint)

	endpoint = "127.0.0.3:5000"
	_, err = proxy.Regist(serviceName, endpoint, 0)
	assert.NoError(t, err)
	defer proxy.UnRegist(serviceName, endpoint)

	endpoints := proxy.GetServiceEndpoints(serviceName)
	assert.Equal(t, []string{
		"127.0.0.1:5000",
		"127.0.0.2:5000",
		"127.0.0.3:5000",
	}, endpoints)
}