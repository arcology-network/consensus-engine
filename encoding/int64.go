package encoding

import (
	"encoding/binary"
)

const (
	INT_LEN = 8
)

type Int64 int64

func (this *Int64) Get() interface{} {
	return *this
}

func (this *Int64) Set(v interface{}) {
	*this = v.(Int64)
}

func (this Int64) Size() uint32 {
	return uint32(INT_LEN)
}

func (this Int64) Encode() []byte {
	buffer := make([]byte, INT_LEN)
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this Int64) EncodeToBuffer(buffer []byte) {
	binary.LittleEndian.PutUint64(buffer, uint64(this))
}

func (this Int64) Decode(buffer []byte) interface{} {
	return Int64(int64(binary.LittleEndian.Uint64(buffer)))
}

type Int64s []Int64

func (this Int64s) Encode() []byte {
	buffer := make([]byte, len(this)*INT_LEN)
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this Int64s) EncodeToBuffer(buffer []byte) {
	for i := 0; i < len(this); i++ {
		binary.LittleEndian.PutUint64(buffer[i*INT_LEN:], uint64(this[i]))
	}
}

func (this Int64s) Decode(buffer []byte) Int64s {
	for i := 0; i < len(this); i++ {
		this[i] = Int64(int64(binary.LittleEndian.Uint64(buffer)))
	}
	return Int64s(this)
}

func (this Int64s) Sum() int64 {
	sum := int64(0)
	for i := range this {
		sum += int64(this[i])
	}
	return sum
}

func (this Int64s) Accumulate() []Int64 {
	if len(this) == 0 {
		return []Int64{}
	}

	values := make([]Int64, len(this))
	values[0] = this[0]
	for i := 1; i < len(this); i++ {
		values[i] = values[i-1] + this[i]
	}
	return values
}
