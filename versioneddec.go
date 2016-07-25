/*

Implementation of the versioned decoder.

*/

package s2prot

// Versioned decoder.
type versionedDec struct {
	*bitPackedBuff            // Data source: bit-packed buffer
	typeInfos      []typeInfo // Type descriptors
}

// newBitPackedDec creates a new bit-packed decoder.
func newVersionedDec(contents []byte, typeInfos []typeInfo) *versionedDec {
	return &versionedDec{
		bitPackedBuff: &bitPackedBuff{
			contents:  contents,
			bigEndian: true, // All versioned decoder uses big endian order
		},
		typeInfos: typeInfos,
	}
}

// instance decodes a value specified by its type id and returns the decoded value.
func (d *versionedDec) instance(typeid int) interface{} {
	b := d.bitPackedBuff // Local var for efficiency and more compact code

	ti := &d.typeInfos[typeid] // Pointer to avoid copying the struct

	switch ti.s2pType {
	case s2pInt:
		b.readBits8() // Field type (9)
		return readVarInt(b)
	case s2pStruct:
		b.readBits8() // Field type (5)
		// TODO order should be preserved! Map does not preserve it!
		s := Struct{}
		length := int(readVarInt(b))
		for i := 0; i < length; i++ {
			tag := int(readVarInt(b))
			var f *field
			for idx := range ti.fields {
				if ti.fields[idx].tag == tag {
					f = &ti.fields[idx]
					break
				}
			}
			if f == nil {
				// We don't have info about the field, skip it
				skipInstance(b)
				continue
			}
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
		b.readBits8() // Field type (3)
		tag := int(readVarInt(b))
		if tag > len(ti.fields) {
			return nil
		}
		f := ti.fields[tag]
		return Struct{f.name: d.instance(f.typeid)}
	case s2pArr:
		b.readBits8() // Field type (0)
		length := readVarInt(b)
		arr := make([]interface{}, length)
		for i := range arr {
			arr[i] = d.instance(ti.typeid)
		}
		return arr
	case s2pBitArr:
		b.readBits8() // Field type (1)
		length := int(readVarInt(b))
		return BitArr{Count: length, Data: b.readAligned((length + 7) / 8)}
	case s2pBlob:
		b.readBits8() // Field type (2)
		length := int(readVarInt(b))
		return string(b.readAligned(length))
	case s2pOptional:
		b.readBits8() // Field type (4)
		if b.readBits8() != 0 {
			return d.instance(ti.typeid)
		}
		return nil
	case s2pBool:
		b.readBits8() // Field type (6)
		return b.readBits8() != 0
	case s2pFourCC:
		b.readBits8() // Field type (7)
		return string(b.readAligned(4))
	case s2pNull:
		return nil
	}

	return nil
}

// readVarInt reads a variable-length int value.
// Format: read from input by 8 bits. Highest bit tells if have to read more bytes,
// lowest bit of the firt byte (first 8 bits) is not data but tells if the number is negative.
func readVarInt(b *bitPackedBuff) int64 {
	var data, value int64
	for shift := uint(0); ; shift += 7 {
		data = int64(b.readBits8())
		value |= (data & 0x7f) << shift
		if (data & 0x80) == 0 {
			if value&0x01 > 0 {
				return -(value >> 1)
			} else {
				return value >> 1
			}
		}
	}
}

// skipInstance reads and discards an instance whose type is deducted from the read Field type.
func skipInstance(b *bitPackedBuff) {
	fieldType := b.readBits8()
	switch fieldType {
	case 0: // array
		for i := readVarInt(b); i > 0; i-- {
			skipInstance(b)
		}
	case 1: // bit array
		b.readAligned((int(readVarInt(b)) + 7) / 8)
	case 2: // blob
		b.readAligned(int(readVarInt(b)))
	case 3: // choice
		readVarInt(b) // tag
		skipInstance(b)
	case 4: // optional
		if b.readBits8() != 0 {
			skipInstance(b)
		}
	case 5: // struct
		for i := readVarInt(b); i > 0; i-- {
			readVarInt(b) // tag
			skipInstance(b)
		}
	case 6: // uint8
		b.readBits8() // b.readAligned(1)
	case 7: // uint32
		b.readAligned(4)
	case 8: // uint64
		b.readAligned(8)
	case 9: // vint
		readVarInt(b)
	}
}
