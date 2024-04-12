package io

var (
	Null       = &null{}
	NullCloser = &nullCloser{}
)

type null struct{}

func (this *null) ReadAt(p []byte, off int64) (int, error) { return 0, nil }

func (this *null) WriteAt(p []byte, off int64) (int, error) { return len(p), nil }

func (this *null) Write(p []byte) (int, error) { return len(p), nil }

func (this *null) Read(p []byte) (int, error) { return 0, nil }

func (this *null) Close() error { return nil }

type nullCloser struct{}

func (this *nullCloser) Close() error { return nil }
