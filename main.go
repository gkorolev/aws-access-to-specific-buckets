package main

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)


// Below we need to embed imported type to prevent reimplementing all methods.

type vmContext struct {
	types.DefaultVMContext
}

type pluginContext struct {
	types.DefaultPluginContext
	configuration pluginConfiguration
}

type bucket struct {
	endpoint string
	name     string
}

type pluginConfiguration struct {
	buckets []bucket
}

type httpContext struct {
	types.DefaultHttpContext
	totalRequestBodySize int
	buckets         []bucket
}


func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	config, err := parsePluginConfiguration(data)
	if err != nil {
		proxywasm.LogCriticalf("error parsing plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	ctx.configuration = config
	return types.OnPluginStartStatusOK
}

// TinyGo doesn't support encoding/json, so using gjson.
func parsePluginConfiguration(data []byte) (pluginConfiguration, error) {
	if len(data) == 0 {
		return pluginConfiguration{}, nil
	}

	config := &pluginConfiguration{}
	if !gjson.ValidBytes(data) {
		return pluginConfiguration{}, fmt.Errorf("the plugin configuration is not a valid json: %q", string(data))
	}

	jsonData := gjson.ParseBytes(data)
	result := jsonData.Get("bucket")
	if result.IsArray() {
		for _, requiredKey := range result.Array() {
			config.buckets = append(config.buckets, bucket{
				endpoint: requiredKey.Get("endpoint").String(),
				name:     requiredKey.Get("name").String(),
			})
		}
	}

	return *config, nil
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{buckets: ctx.configuration.buckets}
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	path, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		proxywasm.LogCriticalf("failed to get path: %v", err)
		return types.ActionContinue
	}

	host, err := proxywasm.GetHttpRequestHeader("host")
	if err != nil {
		host, err = proxywasm.GetHttpRequestHeader(":authority")
		if err != nil {
			proxywasm.LogCriticalf("failed to get host: %v", err)
			return types.ActionContinue
		}
	}

	for _, bucket := range ctx.buckets {
		if host == fmt.Sprintf("%s.%s", bucket.name, bucket.endpoint) || (host == bucket.endpoint && strings.HasPrefix(path, fmt.Sprintf("/%s",bucket.name))) {
			proxywasm.LogInfof("Allowing request to %s", path)
			return types.ActionContinue
		}
	}

	proxywasm.LogWarnf("Blocking request to %s%s", host, path)
	proxywasm.SendHttpResponse(403, [][2]string{
		{":status", "403"},
		{"content-type", "text/plain"},
	}, []byte("Forbidden\n"), -1)
	return types.ActionPause

}


func main() {
	proxywasm.SetVMContext(&vmContext{})
}
