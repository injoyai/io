package io

import "io"

var (

	// Discard 丢弃写入的数据,实现接口
	Discard = io.Discard

	// NopCloser io.Reader 转 io.ReadCloser
	NopCloser = io.NopCloser

	// TeeReader 读取reader数据的同时写入writer
	TeeReader = io.TeeReader

	// NewSectionReader readerAt 转 reader
	NewSectionReader = io.NewSectionReader

	// LimitReader 读取固定长度数据
	LimitReader = io.LimitReader

	// Copy 从reader中读取所有数据到writer
	// Copy = io.Copy

	// CopyBuffer 从reader中读取所有数据到writer,设置读取缓存大小
	CopyBuffer = io.CopyBuffer

	// CopyN same Copy 限制了每次读取的大小
	CopyN = io.CopyN

	// WriteString same writer.Write([]byte(s))
	WriteString = io.WriteString

	// ReadAtLeast 限制了至少读取数量
	ReadAtLeast = io.ReadAtLeast

	// ReadFull 读满buf大小的数据
	ReadFull = io.ReadFull

	// MultiReader 合并多个reader成1个reader
	// 从第一个reader开始读取,待读取全部,依次按顺序向后读取
	MultiReader = io.MultiReader

	JoinReader = io.MultiReader

	// MultiWriter 合并多个writer成1个writer
	// 写入会写入到每一个writer
	MultiWriter = io.MultiWriter

	// Pipe creates a synchronous in-memory pipe.
	// It can be used to connect code expecting an io.Reader
	// with code expecting an io.Writer.
	//
	// Reads and Writes on the pipe are matched one to one
	// except when multiple Reads are needed to consume a single Write.
	// That is, each Write to the PipeWriter blocks until it has satisfied
	// one or more Reads from the PipeReader that fully consume
	// the written data.
	// The data is copied directly from the Write to the corresponding
	// Read (or Reads); there is no internal buffering.
	//
	// It is safe to call Read and Write in parallel with each other or with Close.
	// Parallel calls to Read and parallel calls to Write are also safe:
	// the individual calls will be gated sequentially.
	Pipe = io.Pipe

	// ReadAll 一次性读取全部
	ReadAll = io.ReadAll
)

const (
	SeekStart   = io.SeekStart   // seek relative to the origin of the file
	SeekCurrent = io.SeekCurrent // seek relative to the current offset
	SeekEnd     = io.SeekEnd     // seek relative to the end
)

var (

	// ErrShortWrite means that a write accepted fewer bytes than requested
	// but failed to return an explicit error.
	// "short write"
	ErrShortWrite = io.ErrShortWrite

	// ErrShortBuffer means that a read required a longer buffer than was provided.
	// "short buffer"
	ErrShortBuffer = io.ErrShortBuffer

	// EOF is the error returned by Read when no more input is available.
	// (Read must return EOF itself, not an error wrapping EOF,
	// because callers will test for EOF using ==.)
	// Functions should return EOF only to signal a graceful end of input.
	// If the EOF occurs unexpectedly in a structured data stream,
	// the appropriate error is either ErrUnexpectedEOF or some other error
	// giving more detail.
	// "EOF"
	EOF = io.EOF

	// ErrUnexpectedEOF means that EOF was encountered in the
	// middle of reading a fixed-size block or data structure.
	// "unexpected EOF"
	ErrUnexpectedEOF = io.ErrUnexpectedEOF

	// ErrNoProgress is returned by some clients of an Reader when
	// many calls to Read have failed to return any data or error,
	// usually the sign of a broken Reader implementation.
	// "multiple Read calls return no data or error"
	ErrNoProgress = io.ErrNoProgress

	// ErrClosedPipe is the error used for read or write operations on a closed pipe.
	// "io: read/write on closed pipe"
	ErrClosedPipe = io.ErrClosedPipe
)

type (
	Reader          = io.Reader
	Writer          = io.Writer
	Closer          = io.Closer
	Seeker          = io.Seeker
	ReadWriter      = io.ReadWriter
	ReadCloser      = io.ReadCloser
	WriteCloser     = io.WriteCloser
	ReadWriteCloser = io.ReadWriteCloser
	ReadSeekCloser  = io.ReadSeekCloser
	WriteSeeker     = io.WriteSeeker
	ReadWriteSeeker = io.ReadWriteSeeker
	ReaderFrom      = io.ReaderFrom
	WriterTo        = io.WriterTo
	ReaderAt        = io.ReaderAt
	WriterAt        = io.WriterAt
	ByteReader      = io.ByteReader
	ByteScanner     = io.ByteScanner
	ByteWriter      = io.ByteWriter
	RuneReader      = io.RuneReader
	RuneScanner     = io.RuneScanner
	StringWriter    = io.StringWriter

	PipeReader = io.PipeReader
	PipeWriter = io.PipeWriter
)
