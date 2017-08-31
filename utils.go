package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

func streamToInt8(data []byte) (int8, error) {
	var y int8
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	return y, err
}

func streamToInt16(data []byte) (int16, error) {
	var y int16
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	return y, err
}

func streamToInt32(data []byte) (int32, error) {
	var y int32
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	if y>>31 == 1 {
		y *= -1
	}
	return y, err
}

func streamToInt64(data []byte) (int64, error) {
	var y int64
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	return y, err
}

func streamToFloat32(data []byte) (float32, error) {
	var y float32
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	return y, err
}

func twos_complement(input int32) int32 {
	mask := int32(math.Pow(2, 31))
	return -(input & mask) + (input &^ mask)
}

func streamToTime(data []byte) (time.Time, error) {
	miliseconds, err := streamToInt64(data)
	seconds := int64(float64(miliseconds) / 1000.0)
	nanoseconds := int64(miliseconds % 1000)

	return time.Unix(seconds, nanoseconds), err
}
