package io

import (
	"errors"
	"io"
	"strings"
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
	ErrWithReadTimeout    = errors.New("读超时")
	ErrWithWriteTimeout   = errors.New("写超时")
	ErrInvalidReadFunc    = errors.New("无效数据读取函数")
)

// dealErr 错误处理,常见整理成中文
func dealErr(err error) error {
	if err != nil {
		s := err.Error()
		if err == io.EOF {
			return ErrRemoteClose
		} else if strings.Contains(s, "An existing connection was forcibly closed by the remote host") {
			return ErrRemoteCloseUnusual
		} else if strings.Contains(s, "use of closed network connection") {
			if strings.Contains(s, "close tcp") {
				return ErrCloseClose
			} else if strings.Contains(s, "write tcp") {
				return ErrWriteClosed
			} else if strings.Contains(s, "read tcp") {
				return ErrReadClosed
			}
		} else if strings.Contains(s, "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full") {
			return ErrPortNull
		} else if strings.Contains(s, "connectex: No connection could be made because the target machine actively refused it.") {
			return ErrRemoteOff
		} else if strings.Contains(s, "A socket operation was attempted to an unreachable network") {
			return ErrNetworkUnusual
		}
	}
	return err
}
