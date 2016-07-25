package s2prot

import (
	"bytes"
	"testing"
)

func TestEOFD(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{}, bigEndian: true}
	if !bb.EOF() {
		t.Error("EOF falsely NOT reported.")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3}, bigEndian: true}

	if bb.EOF() {
		t.Error("EOF falsely reported.")
	}
	bb.readBits(1)
	if bb.EOF() {
		t.Error("EOF falsely reported.")
	}
	bb.readBits(7)
	if bb.EOF() {
		t.Error("EOF falsely reported.")
	}
	bb.readBits(1)
	if bb.EOF() {
		t.Error("EOF falsely reported.")
	}
	bb.readBits(12)
	if bb.EOF() {
		t.Error("EOF falsely reported.")
	}
	bb.readBits(3)
	if !bb.EOF() {
		t.Error("EOF falsely NOT reported.")
	}
}

func TestByteAlign(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3}, bigEndian: true}

	bb.byteAlign()
	if bb.readBits(8) != 1 {
		t.Error("Unexpected value!")
	}

	bb.readBits(1)
	bb.byteAlign()
	if bb.readBits(8) != 3 {
		t.Error("Unexpected value!")
	}
}

func TestReadBits1(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{0xaa, 0xaa}, bigEndian: true}

	for expected := false; !bb.EOF(); expected = !expected {
		if bb.readBits1() != expected {
			t.Error("Unexpected value!")
		}
	}
}

func TestReadBits8(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3, 4}, bigEndian: true}

	if bb.readBits8() != 1 {
		t.Error("Unexpected value!")
	}
	bb.readBits(3)
	if bb.readBits8() != 3 {
		t.Error("Unexpected value!")
	}
	if bb.readBits8() != 4 {
		t.Error("Unexpected value!")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3, 4}, bigEndian: false}

	if bb.readBits8() != 1 {
		t.Error("Unexpected value!")
	}
	bb.readBits(3)
	if bb.readBits8() != 0x60 {
		t.Error("Unexpected value!")
	}
	if bb.readBits8() != 0x80 {
		t.Error("Unexpected value!")
	}
}

func TestReadBits(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3, 4}, bigEndian: true}
	if bb.readBits(0) != 0 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(8) != 1 {
		t.Error("Unexpected value!")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3, 4}, bigEndian: true}

	if bb.readBits(3) != 1 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(13) != 2 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(1) != 1 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(15) != 0x0104 {
		t.Error("Unexpected value!")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3, 4}, bigEndian: false}

	if bb.readBits(3) != 1 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(13) != 0x40 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(1) != 1 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(15) != 0x0201 {
		t.Error("Unexpected value!")
	}
}

func TestReadBitsSpecial(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3, 4, 5, 6, 7, 8}, bigEndian: true}

	bb.readBits(3) // Non-empty cache
	if bb.readBits(8) != 2 {
		t.Error("Unexpected value!")
	}
	// 111111111 222222222 333333333 444444444 555555555 666666666 777777777 888888888
	// 0000 0001 0000 0010 0000 0011 0000 0100 0000 0101 0000 0110 0000 0111 0000 1000
	// 0000 0000 0001 1100
	if bb.readBits(16) != 0x001c {
		t.Error("Unexpected value!")
	}
	// 0000 0000 0010 1000 0011 0000 0011 1000
	if bb.readBits(32) != 0x00283038 {
		t.Error("Unexpected value!")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3, 4, 5, 6, 7}, bigEndian: false}

	// Empty cache
	if bb.readBits(8) != 1 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(16) != 0x0302 {
		t.Error("Unexpected value!")
	}
	if bb.readBits(32) != 0x07060504 {
		t.Error("Unexpected value!")
	}
}

func TestReadAligned(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3, 4, 5, 6, 7, 8}, bigEndian: true}

	if !bytes.Equal([]byte{}, bb.readAligned(0)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{1}, bb.readAligned(1)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{2, 3}, bb.readAligned(2)) {
		t.Error("Unexpected value!")
	}
	bb.readBits(3)
	if !bytes.Equal([]byte{5, 6, 7, 8}, bb.readAligned(4)) {
		t.Error("Unexpected value!")
	}
}

func TestReadUnaligned(t *testing.T) {
	bb := &bitPackedBuff{contents: []byte{1, 2, 3, 4, 5, 6, 7, 8}, bigEndian: true}

	if !bytes.Equal([]byte{}, bb.readUnaligned(0)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{1, 2}, bb.readUnaligned(2)) {
		t.Error("Unexpected value!")
	}
	bb.readBits(3)
	if !bytes.Equal([]byte{0x04, 0x05}, bb.readUnaligned(2)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{0x06, 0x07, 0x00}, bb.readUnaligned(3)) {
		t.Error("Unexpected value!")
	}

	bb = &bitPackedBuff{contents: []byte{1, 2, 3, 4, 5, 6, 7, 8}, bigEndian: false}

	if !bytes.Equal([]byte{}, bb.readUnaligned(0)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{1, 2}, bb.readUnaligned(2)) {
		t.Error("Unexpected value!")
	}
	bb.readBits(3)
	// 111111111 222222222 333333333 444444444 555555555 666666666 777777777 888888888
	// 0000 0001 0000 0010 0000 0011 0000 0100 0000 0101 0000 0110 0000 0111 0000 1000
	if !bytes.Equal([]byte{0x80, 0xa0}, bb.readUnaligned(2)) {
		t.Error("Unexpected value!")
	}
	if !bytes.Equal([]byte{0xc0, 0xe0, 0x00}, bb.readUnaligned(3)) {
		t.Error("Unexpected value!")
	}
}
