/*

Types describing decoding instructions for protocol types.

*/

package s2prot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// S2protocol type
type s2pType int

// S2protocol types
const (
	s2pInt      s2pType = iota // An integer number
	s2pStruct                  // A structure (list of fields)
	s2pChoice                  // A choice of multiple types (one of multiple)
	s2pArr                     // List of elements of the same type
	s2pBitArr                  // List of bits (packed into a byte array)
	s2pBlob                    // A byte array
	s2pOptional                // Optionally a value (of a specified type)
	s2pBool                    // A bool value
	s2pFourCC                  // 4 bytes data, usually interpreted as string
	s2pNull                    // Exactly as its name says: nothing
)

// Precached map from type names to S2pType value (for faster parsing).
// First 2 character (excluding the underscore '_') is unique, so just use that:
var nameS2pTypes = map[string]s2pType{"in": s2pInt, "st": s2pStruct, "ch": s2pChoice, "ar": s2pArr,
	"bi": s2pBitArr, "bl": s2pBlob, "op": s2pOptional, "bo": s2pBool, "fo": s2pFourCC, "nu": s2pNull}

// Describes a field in structures.
// Fields used for structures (stStruct) have/use the tag attribute,
// fields used for choices (stChoice) omit the tag.
type field struct {
	name   string // Name of the field
	typeid int    // Type id (index) of the type info of the field's value
	tag    int    // Optional tag of the field (often used for field index).

	isNameParent bool // Tells if field name equals to "__parent" (for optimization purposes, it is checked many times and the result is constant)
}

// Decoding info for a specific type.
type typeInfo struct {
	s2pType s2pType // Type selector; specifies how to read the value and what further fields are valid/filled

	// Optional parameters for decoding, filled values depend on typeSel

	// Bounds for int (and also for choice and array and bitarray and blob)
	offset32 int32 // 32-bit offset to add to the read value
	offset64 int64 // 64-bit offset to add to the read value
	bits     int   // Number of bits to read

	// For struct, and also for choice
	fields []field // List of fields (in case of struct)

	// For array, also used for optional
	typeid int // Type id (index) of the elements of the array
}

// parseTypeInfo parses a TypeInfo from a python string representation.
// Panics if input is in invalid format.
func parseTypeInfo(s string) typeInfo {
	var err error

	// Decode type name, example:
	// ('_int',[(0,7)]),  #0
	s = s[strings.IndexByte(s, '\'')+2:] // All start with an underscore '_', cut that also

	// Map keys are the first 2 characters of the names
	ti := typeInfo{s2pType: nameS2pTypes[s[:2]]}

	if ti.s2pType == s2pOptional {
		// In case of Optional no parenthesis follows, only skip 1 character, 2nd is part of the number
		s = s[strings.IndexByte(s, '[')+1:]
	} else {
		s = s[strings.IndexByte(s, '[')+2:]
	}

	// Helper function to read intbounds specified in the form of "(0,7)" (positioned after the parenthesis)
	// Returns the last index (closing parenthesis)
	readBounds := func() int {
		// Parameters: offset and bits which will provide an integer value
		i := strings.IndexByte(s, ',')
		j := strings.IndexByte(s, ')')
		if ti.bits, err = strconv.Atoi(s[i+1 : j]); err != nil {
			panic(err)
		}
		if ti.offset64, err = strconv.ParseInt(s[:i], 10, 64); err != nil {
			panic(err)
		}
		if ti.bits <= 32 {
			ti.offset32 = int32(ti.offset64)
		}
		return j
	}

	switch ti.s2pType {
	case s2pInt: // ('_int',[(0,7)]),  #0
		// Parameters: offset and bits which will provide the integer value
		readBounds()
	case s2pStruct: // ('_struct',[[('m_name',71,-3),('m_type',6,-2),('m_data',20,-1)]]),  #73
		// Parameters: list of fields
		fields := make([]field, 0, 8)
		for {
			i := strings.IndexByte(s, '\'')
			if i < 0 {
				break // No more fields
			}
			s = s[i+1:]
			i = strings.IndexByte(s, '\'')
			f := field{name: s[:i]}
			f.isNameParent = f.name == "__parent"
			// Most field names start with "m_". Cut that off.
			if strings.HasPrefix(f.name, "m_") {
				f.name = f.name[2:]
			}
			s = s[i+2:]
			i = strings.IndexByte(s, ',')
			j := strings.IndexByte(s, ')')
			if f.typeid, err = strconv.Atoi(s[:i]); err != nil {
				panic(err)
			}
			if f.tag, err = strconv.Atoi(s[i+1 : j]); err != nil {
				panic(err)
			}
			fields = append(fields, f)
		}
		// Copy a trimmed version of this to type info:
		ti.fields = make([]field, len(fields))
		copy(ti.fields, fields)
	case s2pChoice: // ('_choice',[(0,2),{0:('None',91),1:('TargetPoint',93),2:('TargetUnit',94),3:('Data',6)}]),  #95
		// Parameters: offset and bits which will provide the index integer value to choose
		// from the following field list
		i := readBounds()
		s = s[i+1:]
		fields := make([]field, 0, 8)
		for {
			if s[1] == '}' {
				break // No more fields
			}
			s = s[2:]
			i := strings.IndexByte(s, ':')
			f := field{}
			if f.tag, err = strconv.Atoi(s[:i]); err != nil {
				panic(err)
			}
			s = s[strings.IndexByte(s, '\'')+1:]
			i = strings.IndexByte(s, '\'')
			f.name = s[:i]
			s = s[i+2:]
			i = strings.IndexByte(s, ')')
			if f.typeid, err = strconv.Atoi(s[:i]); err != nil {
				panic(err)
			}
			s = s[i:]
			fields = append(fields, f)
		}
		// Copy a trimmed version of this to type info:
		ti.fields = make([]field, len(fields))
		copy(ti.fields, fields)
	case s2pArr: // ('_array',[(16,0),10]),  #14
		// Parameters: offset+bits which will provide the array length, and a typeid (element type)
		s = s[readBounds()+2:]
		j := strings.IndexByte(s, ']')
		if ti.typeid, err = strconv.Atoi(s[:j]); err != nil {
			panic(err)
		}
	case s2pBitArr: // ('_bitarray',[(0,6)]),  #52
		// Parameters: offset and bits which will provide the number of bits
		readBounds()
	case s2pBlob: // ('_blob',[(0,8)]),  #9
		// Parameters: offset and bits which will provide the array length (number of bytes)
		readBounds()
	case s2pOptional: // ('_optional',[14]),  #15
		// Parameters: typeid (type of the value that optionally follows)
		j := strings.IndexByte(s, ']')
		if ti.typeid, err = strconv.Atoi(s[:j]); err != nil {
			panic(err)
		}
	case s2pBool: // ('_bool',[]),  #13
		// We're done, nothing to do (no parameters)
	case s2pFourCC: // ('_fourcc',[]),  #19
		// We're done, nothing to do (no parameters)
	case s2pNull: // ('_null',[]),  #91
		// We're done, nothing to do (no parameters)
	}

	return ti
}

// Struct represents a decoded struct.
// It is a dynamic struct modelled with a general map with helper methods to access its content.
//
// Tip: use the encoding/json package to nicely format Struct values, e.g.:
//
//     data, _ := json.MarshalIndent(someStruct, "", "  ")
//     fmt.Printf("Full Struct:\n%s\n", data)
type Struct map[string]interface{}

// Value returns the value specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Value(path ...string) interface{} {
	if len(path) == 0 {
		return nil
	}

	ss, ok := *s, false

	last := len(path) - 1
	for i := 0; i < last; i++ {
		if ss, ok = ss[path[i]].(Struct); !ok {
			return nil
		}
	}

	return ss[path[last]]
}

// Structv returns the (sub) Struct specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Structv(path ...string) (v Struct) {
	v, _ = s.Value(path...).(Struct)
	return
}

// Int returns the integer specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Int(path ...string) (v int64) {
	v, _ = s.Value(path...).(int64)
	return
}

// Bool returns the bool specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Bool(path ...string) (v bool) {
	v, _ = s.Value(path...).(bool)
	return
}

// Bytes returns the []byte specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Bytes(path ...string) (v []byte) {
	v, _ = s.Value(path...).([]byte)
	return
}

// Text returns the []byte specified by the path converted to string.
// zero value is returned if path is invalid.
func (s *Struct) Text(path ...string) string {
	v, ok := s.Value(path...).([]byte)
	if ok {
		return string(v)
	}
	return ""
}

// Stringv returns the string specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Stringv(path ...string) (v string) {
	v, _ = s.Value(path...).(string)
	return
}

// Array returns the array (of empty interfaces) specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) Array(path ...string) (v []interface{}) {
	v, _ = s.Value(path...).([]interface{})
	return
}

// BitArr returns the bit array specified by the path.
// zero value is returned if path is invalid.
func (s *Struct) BitArr(path ...string) (v BitArr) {
	v, _ = s.Value(path...).(BitArr)
	return
}

// String returns the idented JSON string representation of the Struct.
func (s *Struct) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// Event
type Event struct {
	Struct
	*EvtType // Pointer only to avoid copying
}

// Loop returns the loop (time) of the event.
func (e *Event) Loop() int64 {
	return e.Int("loop")
}

// UserId returns the id of the user that issued the event.
func (e *Event) UserId() int64 {
	return e.Int("userid")
}

// Bit array which stores the bits in a byte slice.
type BitArr struct {
	Count int    // Bits count
	Data  []byte // Data holding the bits
}

// Bit masks having exactly 1 one bit at the position specified by the index (zero-based).
var singleBitMasks = [...]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}

// Bit tells if the bit at the specified position (zero-based) is 1.
func (b *BitArr) Bit(n int) bool {
	return b.Data[n>>3]&singleBitMasks[n&0x07] != 0
}

// Cached array which tells the nubmer of 1 bits in the number specified by the index.
var ones [256]int

func init() {
	// Initialize / compute the ones array.
	for i := range ones {
		c := 0
		for j := i; j > 0; j >>= 1 {
			if j&0x01 != 0 {
				c++
			}
		}
		ones[i] = c
	}
}

// Ones returns the number of 1 bits in the bit array.
func (b *BitArr) Ones() (c int) {
	for _, d := range b.Data {
		c += ones[d]
	}
	return
}

// String returns the string representation of the bit array in hexadecimal form.
// Using value receiver so printing a BitArr value will call this method.
func (b BitArr) String() string {
	return fmt.Sprintf("0x%s (count=%d)", hex.EncodeToString(b.Data), b.Count)
}

// MarshalJSON produces a custom JSON string for a more informative and more compact representation of the bitarray.
// The essence is that the Data slice is presented in hex format (instead of the default Base64 encoding).
func (b BitArr) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"Count":%d,"Data": "0x%s"}`, b.Count, hex.EncodeToString(b.Data))), nil
}
