package util

import (
	"bytes"
	"encoding/binary"
)

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrMsg(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}
