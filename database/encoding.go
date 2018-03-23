package database

import (
	"encoding/binary"
	"math"
)

func uint64ToByte(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func byteToUint64(arr []byte) uint64 {
	return binary.BigEndian.Uint64(arr)
}

func intToByte(v int) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, int64(v))
	return b
}

func byteToInt(arr []byte) int {
	val, _ := binary.Varint(arr)
	return int(val)
}

func float64ToByte(f float64) []byte {
	return uint64ToByte(math.Float64bits(f))
}

func byteToFloat64(arr []byte) float64 {
	return math.Float64frombits(byteToUint64(arr))
}
