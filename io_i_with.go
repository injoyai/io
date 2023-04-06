package io

import "context"

func WithServerDebug(b ...bool) func(s *Server) {
	return func(s *Server) { s.Debug(b...) }
}

func WithClientDebug() func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) { c.Debug() }
}

func WithClientSetKey(key string) func(ctx context.Context, c *Client) {
	return func(ctx context.Context, c *Client) { c.SetKey(key) }
}
