package io

import "context"

func WithServerDebug(b ...bool) OptionServer {
	return func(s *Server) { s.Debug(b...) }
}

func WithClientDebug() OptionClient {
	return func(ctx context.Context, c *Client) { c.Debug() }
}

func WithClientPrintBase(b ...bool) OptionClient {
	return func(ctx context.Context, c *Client) {
		c.Debug(b...)
		c.SetPrintWithBase()
	}
}

func WithServerPrintBase(b ...bool) OptionServer {
	return func(s *Server) {
		s.Debug(b...)
		s.SetPrintWithBase()
	}
}

func WithClientSetKey(key string) OptionClient {
	return func(ctx context.Context, c *Client) { c.SetKey(key) }
}

func WithClientReadWritePkg() OptionClient {
	return func(ctx context.Context, c *Client) {
		c.SetReadWriteWithPkg()
	}
}

func WithServerReadWritePkg() OptionServer {
	return func(s *Server) {
		s.SetReadWriteWithPkg()
	}
}
