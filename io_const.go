package io

const (
	Simplex    = "Simplex"    //单工
	HalfDuplex = "HalfDuplex" //半双工
	FullDuplex = "FullDuplex" //全双工

	KB1               = 1 << 10 //1KB
	KB4               = 4 << 10 //4KB
	MB1               = 1 << 20 //1MB
	DefaultBufferSize = KB4
)
