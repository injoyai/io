package rpc

/*
私有RPC协议


*/

type Type int

const (
	String byte = 0x01
	Bool   byte = 0x02
	Int    byte = 0x03
	Float  byte = 0x04
	Object byte = 0x05
	Array  byte = 0x06
)
