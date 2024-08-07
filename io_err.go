package io

import (
	"errors"
	"io"
	"strings"
)

var (
	// DealErr 错误处理配置
	DealErr = defaultDealErr
)

var (
	ErrHandClose          = errors.New("主动关闭")
	ErrRemoteClose        = errors.New("远程端主动关闭连接")
	ErrRemoteCloseUnusual = errors.New("远程端意外关闭连接")
	ErrCloseClose         = errors.New("关闭已关闭连接")
	ErrWriteClosed        = errors.New("写入已关闭连接")
	ErrReadClosed         = errors.New("读取已关闭连接")
	ErrPortNull           = errors.New("端口不足或绑定未释放")
	ErrRemoteOff          = errors.New("远程服务可能未开启")
	ErrNetworkUnusual     = errors.New("网络异常")
	ErrWithContext        = errors.New("上下文关闭")
	ErrWithTimeout        = errors.New("超时")
	ErrWithConnectTimeout = errors.New("连接超时")
	ErrWithReadTimeout    = errors.New("读超时")
	ErrWithWriteTimeout   = errors.New("写超时")
	ErrInvalidReadFunc    = errors.New("无效数据读取函数")
	ErrMaxConnect         = errors.New("到达最大连接数")
	ErrUseReadMessage     = errors.New("不支持,请使用ReadMessage")
	ErrUseReadAck         = errors.New("不支持,请使用ReadAck")
)

// 错误处理 错误信息处理
func dealErr(err error) error {
	if DealErr == nil {
		return err
	}
	return DealErr(err)
}

// defaultDealErr 错误处理,常见整理成中文
func defaultDealErr(err error) error {
	if err != nil {
		s := err.Error()
		switch {
		case err == io.EOF:

		case strings.Contains(s, "An existing connection was forcibly closed by the remote host"):
			return ErrRemoteCloseUnusual

		case strings.Contains(s, "use of closed network connection"):
			switch {
			case strings.Contains(s, "close tcp"):
				return ErrCloseClose

			case strings.Contains(s, "write tcp"):
				return ErrWriteClosed

			case strings.Contains(s, "read tcp"):
				return ErrReadClosed

			}

		case strings.Contains(s, "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full"):
			return ErrPortNull

		case strings.Contains(s, "connectex: No connection could be made because the target machine actively refused it."):
			return ErrRemoteOff

		case strings.Contains(s, "A socket operation was attempted to an unreachable network"):
			return ErrNetworkUnusual

		}
	}
	return err
}
