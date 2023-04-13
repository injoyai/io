package io

import "context"

func WithServerDebug(b ...bool) func(s *Server) {
	return func(s *Server) { s.Debug(b...) }
}

func WithClientDebug() func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) { c.Debug() }
}

func WithClientPrintBase(b ...bool) func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) {
		c.Debug(b...)
		c.SetPrintWithBase()
	}
}

func WithServerPrintBase(b ...bool) func(s *Server) {
	return func(s *Server) {
		s.Debug(b...)
		s.SetPrintWithBase()
	}
}

func WithClientSetKey(key string) func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) { c.SetKey(key) }
}

func WithClientReadWritePkg() func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) {
		c.SetReadWriteWithPkg()
	}
}

func WithServerReadWritePkg() func(s *Server) {
	return func(s *Server) {
		s.SetReadWriteWithPkg()
	}
}
