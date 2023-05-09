package io

func WithServerDebug(b ...bool) OptionServer {
	return func(s *Server) { s.Debug(b...) }
}

func WithClientDebug() OptionClient {
	return func(c *Client) { c.Debug() }
}

func WithClientPrintBase(b ...bool) OptionClient {
	return func(c *Client) {
		c.Debug(b...)
		c.SetPrintWithBase()
	}
}

func WithClientPrintErr(b ...bool) OptionClient {
	return func(c *Client) {
		c.Debug(b...)
		c.SetPrintWithErr()
	}
}

func WithServerPrintBase(b ...bool) OptionServer {
	return func(s *Server) {
		s.Debug(b...)
		s.SetPrintWithBase()
	}
}

func WithServerPrintErr(b ...bool) OptionServer {
	return func(s *Server) {
		s.Debug(b...)
		s.SetPrintWithErr()
	}
}

func WithClientSetKey(key string) OptionClient {
	return func(c *Client) { c.SetKey(key) }
}

func WithClientReadWritePkg() OptionClient {
	return func(c *Client) {
		c.SetReadWriteWithPkg()
	}
}

func WithServerReadWritePkg() OptionServer {
	return func(s *Server) {
		s.SetReadWriteWithPkg()
	}
}
