/*

Implementation of a byte buffer whose content can be accessed/interpreted by bits.

*/

package s2prot

// Bit masks having as many ones at the lowest bits as the index.
var bitMasks = [...]byte{0x00, 0x01, 0x03, 0x07, 0x0f, 0x1f, 0x3f, 0x7f, 0xff}

// The wrapper around a []byte providing access by arbitrary number of bits.
type bitPackedBuff struct {
	contents  []byte // Source of bits
	bigEndian bool   // Tells if numbers constucted from read bits are coming in big endian byte order.
	idx       int    // Index of the next byte from contents (this equals to bytes already read/processed)
	cache     byte   // Cache of the byte whose bits are next
	cacheBits byte   // Unused bits in cache
}

// EOF tells if end of buffer reached.
func (b *bitPackedBuff) EOF() bool {
	return b.cacheBits == 0 && b.idx >= len(b.contents)
}

// byteAlign aligns the buffer to byte boundary.
// This means if there are unused bits from the cached, last read byte, they are thrown away.
func (b *bitPackedBuff) byteAlign() {
	b.cacheBits = 0
}

// readBits1 reads 1 bit and returns true if the bit is 1, and returns false if the bit is 0.
// This method is more efficient than but has the same effect as the code:
//     readBits(1) != 0
func (b *bitPackedBuff) readBits1() bool {
	// No need to check endianness, we only need 1 bit (it can't be split in multiple bytes)

	if b.cacheBits == 0 {
		cache := b.contents[b.idx]
		b.cache = cache >> 1
		b.idx++
		b.cacheBits = 7
		return cache&0x01 == 1
	}

	res := b.cache&0x01 == 1
	b.cache >>= 1
	b.cacheBits--
	return res
}

// readBits8 reads 8 bits and returns it as a byte.
// This method is more efficient than but has the same effect as the code:
//     readBits(8)
func (b *bitPackedBuff) readBits8() (r byte) {
	// No need to update b.cacheBits because we read 8 bits (and would be the same)

	if b.cacheBits == 0 {
		// No need to check endianness, we need the next complete byte as-is
		r = b.contents[b.idx]
		b.idx++
		return
	}

	compBits := 8 - b.cacheBits // complementary bits, bits needed from the next byte
	if b.bigEndian {
		r = b.cache << compBits // no need to mask, we need all cache bits
		b.cache = b.contents[b.idx]
		b.idx++
		r |= b.cache & bitMasks[compBits]
	} else {
		r = b.cache // no need to mask, we need all cache bits
		b.cache = b.contents[b.idx]
		b.idx++
		r |= (b.cache & bitMasks[compBits]) << b.cacheBits
	}
	b.cache >>= compBits
	return
}

// readBits returns a number constructed from the next n bits.
func (b *bitPackedBuff) readBits(n byte) int64 {
	// n might be 0!
	if n == 0 {
		return 0
	}

	if b.bigEndian {
		// If applicable, call the optimized version:
		// (I omit optimizing the extremely rare case of n being multiple of 8 AND cache being empty.)
		if n&0x07 == 0 && b.cacheBits != 0 {
			return b.readBitsBigByte(n)
		}
		return b.readBitsBig(n)
	}

	// Highly optimized case for cache being empty and n being multiple of 8.
	// Actually this is true 100% of the cases (little endian is only used to decode attributes events).
	if n&0x07 == 0 && b.cacheBits == 0 {
		// Remember: n > 0 (n == 0 is already handled)
		value := int64(b.contents[b.idx])
		b.idx++
		for i := byte(8); i < n; i += 8 {
			value |= int64(b.contents[b.idx]) << i
			b.idx++
		}
		return value
	}
	return b.readBitsLittle(n)
}

// readBitsBigByte returns a number constructed from the next n bits, using big-endian byte order.
// This is a highly optimized version for a special and frequent case of:
//     - n must be a multiple of 8 and must be greater than 0
//     - cache must not be empty (cacheBits != 0).
func (b *bitPackedBuff) readBitsBigByte(n byte) (value int64) {
	// Cache bits
	value = int64(b.cache) // no need to mask, we need all cache bits

	// Read whole bytes
	for ; n > 8; n -= 8 {
		value = (value << 8) | int64(b.contents[b.idx])
		b.idx++
	}

	b.cache = b.contents[b.idx]
	b.idx++

	compBits := 8 - b.cacheBits // complementary bits, bits needed from the last byte
	value = (value << compBits) | int64(b.cache&bitMasks[compBits])
	b.cache >>= compBits

	return value
}

// readBitsBig returns a number constructed from the next n bits, using big-endian byte order.
// Here n is not allowed to be 0.
func (b *bitPackedBuff) readBitsBig(n byte) (value int64) {
	for {
		if b.cacheBits == 0 {
			b.cache = b.contents[b.idx]
			b.idx++
			b.cacheBits = 8
		}

		// How many bits to use from cache?
		switch {
		case n > b.cacheBits: // All bits from cache are needed, and it's not even enough
			value = (value << b.cacheBits) | int64(b.cache) // no need to mask, we need all cache bits
			n -= b.cacheBits
			b.cacheBits = 0
			// Nothing left in cache, no need to shift it (will be overridden on next read)
		case n < b.cacheBits: // Some bits from cache are needed but not all
			value = (value << n) | int64(b.cache&bitMasks[n])
			b.cacheBits -= n
			b.cache >>= n
			return
		default: // n == b.cacheBits: cache contains exactly as many as needed
			value = (value << n) | int64(b.cache) // no need to mask, we need all cache bits
			b.cacheBits = 0
			// Nothing left in cache, no need to shift it (will be overridden on next read)
			return
		}
	}
}

// readBitsLittle returns a number constructed from the next n bits, using little-endian byte order.
// Here n is not allowed to be 0.
func (b *bitPackedBuff) readBitsLittle(n byte) (value int64) {
	var valueBits byte // Bits already set in value
	for {
		if b.cacheBits == 0 {
			b.cache = b.contents[b.idx]
			b.idx++
			b.cacheBits = 8
		}

		// How many bits to use from cache?
		switch {
		case n > b.cacheBits: // All bits from cache are needed, and it's not even enough
			value |= int64(b.cache) << valueBits // no need to mask, we need all cache bits
			n -= b.cacheBits
			valueBits += b.cacheBits
			b.cacheBits = 0
			// Nothing left in cache, no need to shift it (will be overridden on next read)
		case n < b.cacheBits: // Some bits from cache are needed but not all
			value |= int64(b.cache&bitMasks[n]) << valueBits
			b.cacheBits -= n
			b.cache >>= n
			return
		default: // n == b.cacheBits: cache contains exactly as many as needed
			value |= int64(b.cache) << valueBits // no need to mask, we need all cache bits
			b.cacheBits = 0
			// Nothing left in cache, no need to shift it (will be overridden on next read)
			return
		}
	}
}

// readAligned first aligns to a byte and reads and returns n bytes.
func (b *bitPackedBuff) readAligned(n int) (buff []byte) {
	b.byteAlign()

	buff = make([]byte, n)
	b.idx += copy(buff, b.contents[b.idx:])

	return
}

// readUnaligned reads and returns n bytes (or more precisely n*8 bits).
func (b *bitPackedBuff) readUnaligned(n int) (buff []byte) {
	buff = make([]byte, n)
	if n == 0 {
		return
	}

	// A quick check: if we're at a byte boundary,
	// then reading unaligned is the same as reading aligned which is much faster:
	// no bit shift and masking is required, just copy the bytes
	if b.cacheBits == 0 {
		b.idx += copy(buff, b.contents[b.idx:])
		return
	}

	// Simplest / naive solution, fast(er) only if n is small:
	if n <= 2 {
		for i := range buff {
			buff[i] = byte(b.readBits(8))
		}
		return
	}

	// Highly optimized version for the case: b.cacheBits != 0
	// Since we read bits by 8, cacheBits will not change at the end,
	// and we never have to modify/update it as we know we need 8-cacheBits from the next byte.

	// local vars for efficiency
	idx := b.idx

	complCacheBits := 8 - b.cacheBits
	if b.bigEndian {
		// In case of bigEndian we don't even have to shift the cache, just apply bitmasks.
		// We only have to shift (back) the current value in the cache because it is already shifted.
		// And once we're done, shift the remaining last value in the cache as expected.
		mask := bitMasks[complCacheBits]
		umask := ^mask
		cache := b.cache << complCacheBits
		for i := range buff {
			value := cache & umask
			cache = b.contents[idx]
			idx++
			buff[i] = value | (cache & mask)
		}
		b.cache = cache >> complCacheBits
	} else {
		cacheBits := b.cacheBits
		mask := bitMasks[cacheBits]
		cache := b.cache
		for i := range buff {
			value := cache & mask // Masking only needed from 2nd iteration but doesn't hurt in the first
			cache = b.contents[idx]
			idx++
			buff[i] = value | (cache << cacheBits) // No need to mask as cache is byte and shifting right will shift out the bits
			cache >>= complCacheBits
		}
		b.cache = cache
	}

	// restore from local vars:
	b.idx = idx

	return
}
