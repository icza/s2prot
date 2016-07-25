/*

Implementation of the bit-packed decoder.

*/

package s2prot

// Bit-packed decoder.
type bitPackedDec struct {
	*bitPackedBuff            // Data source: bit-packed buffer
	typeInfos      []typeInfo // Type descriptors
}

// newBitPackedDec creates a new bit-packed decoder.
func newBitPackedDec(contents []byte, typeInfos []typeInfo) *bitPackedDec {
	return &bitPackedDec{
		bitPackedBuff: &bitPackedBuff{
			contents:  contents,
			bigEndian: true, // All bit-packed decoder uses big endian order
		},
		typeInfos: typeInfos,
	}
}

// instance decodes a value specified by its type id and returns the decoded value.
func (d *bitPackedDec) instance(typeid int) interface{} {
	b := d.bitPackedBuff // Local var for efficiency and more compact code

	ti := &d.typeInfos[typeid] // Pointer to avoid copying the struct

	// Helper function to read an integer specified by the type info
	readInt := func() int64 {
		return ti.offset64 + b.readBits(byte(ti.bits))
	}

	switch ti.s2pType {
	case s2pInt:
		return readInt()
	case s2pStruct:
		// TODO order should be preserved! Map does not preserve it!
		s := Struct{}
		for _, f := range ti.fields {
			if f.isNameParent {
				parent := d.instance(f.typeid)
				if s2, ok := parent.(Struct); ok {
					// Copy s2 into s
					for k, v := range s2 {
						s[k] = v
					}
				} else if len(ti.fields) == 1 {
					return parent
				} else {
					s[f.name] = parent
				}
			} else {
				s[f.name] = d.instance(f.typeid)
			}
		}
		return s
	case s2pChoice:
		tag := int(readInt())
		if tag > len(ti.fields) {
			return nil
		}
		f := ti.fields[tag]
		return Struct{f.name: d.instance(f.typeid)}
	case s2pArr:
		length := readInt()
		arr := make([]interface{}, length)
		for i := range arr {
			arr[i] = d.instance(ti.typeid)
		}
		return arr
	case s2pBitArr:
		// length may be > 64, so simple readBits() is not enough
		length := int(readInt())
		buf := make([]byte, (length+7)/8)    // Number of required bytes
		copy(buf, b.readUnaligned(length/8)) // Number of whole bytes:
		if remaining := byte(length % 8); remaining != 0 {
			buf[len(buf)-1] = byte(b.readBits(remaining))
		}
		return BitArr{Count: length, Data: buf}
	case s2pBlob:
		length := readInt()
		return string(b.readAligned(int(length)))
	case s2pOptional:
		if b.readBits1() {
			return d.instance(ti.typeid)
		}
		return nil
	case s2pBool:
		return b.readBits1()
	case s2pFourCC:
		return string(b.readUnaligned(4))
	case s2pNull:
		return nil
	}

	return nil
}
