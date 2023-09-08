package io

var (
	// DealErr 错误处理配置
	DealErr = defaultDealErr

	// Log 日志配置
	Log ILog = &_log{}
)

func Default() {
	DealErr = defaultDealErr
	Log = &_log{}
}
