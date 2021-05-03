package SURLTools

import (
	"log"
)

// MAX_URLS == 62^6 < 64^6 == 2^36, so MAX_URLS (and, subsequently, any index
// value) fits into 64-bit integer (here ^ denotes exponentiation)
const MAX_URLS uint64 = 62 * 62 * 62 * 62 * 62 * 62

// bigMultMod computes (a * b) % MAX_URLS for 36-bit integers.
// We can not write directly (a * b) % MAX_URLS, because product of two
// 36-bit integers may overflow uint64.
func bigMultMod(a, b uint64) (result uint64) {
	if a >= 1<<36 || b >= 1<<36 {
		log.Fatal("Error in bigMultMod: longer-than-36-bit integers supplied")
		// This should never happen. If it happens, it indicates a bug elsewhere
		// in the code.
	}
	bUpper := (b & 0xfffff0000) / 0x10000 // upper 20 bits
	bLower := b & 0xffff                  // lower 16 bits

	result = (a * bUpper) % MAX_URLS
	result = (result * 0x10000) % MAX_URLS
	result = (result + a*bLower) % MAX_URLS

	return result
}

// IndexToShortURL generates seemingly random strings.
func IndexToShortURL(index uint64) string {

	// ensure that the first generated string will not be "000000"
	index = (index + 1) % MAX_URLS

	// Since 35104476159 is coprime to MAX_URLS, the result of the following
	// line will be different for all indexes in range 0 .. MAX_URLS-1.
	index = bigMultMod(index, 35104476159)

	// By the way, the constant 35104476159 is the integer nearest to
	// 1/((1+sqrt(5))/2)*MAX_URLS coprime to MAX_URLS. I have chosen the factor
	// 1/((1+sqrt(5))/2), the reciprocal of the golden ratio, because it nicely
	// unpredictably jumps all over the range.

	// now convert to base 62

	strb := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		strb[i] = byte(index % 62)
		index = index / 62
		if 0 <= strb[i] && strb[i] < 10 {
			strb[i] = strb[i] + '0'
		} else if 10 <= strb[i] && strb[i] < 10+26 {
			strb[i] = strb[i] - 10 + 'a'
		} else { // 10+26 <= strb[i] && strb[i] < 10+26+26
			strb[i] = strb[i] - 10 - 26 + 'A'
		}
	}
	return string(strb)
}

// ShortURLToIndex transforms strings back into their indexes
func ShortURLToIndex(str string) (index uint64) {

	// first convert from base 62 back to base 10

	var x byte // base 62 digit
	for i := 0; i < 6; i++ {
		if '0' <= str[i] && str[i] <= '9' {
			x = str[i] - '0'
		} else if 'a' <= str[i] && str[i] <= 'z' {
			x = str[i] - 'a' + 10
		} else { // 'A' <= str[i] && str[i] <= 'Z'
			x = str[i] - 'A' + 10 + 26
		}
		index = 62*index + uint64(x)
	}

	// The following lines retrieve the original index. It works because
	// (35104476159 * 768306879) % MAX_URLS == 1
	index = bigMultMod(index, 768306879)

	// this line ensures that the (index - 1) on the next line is nonnegative
	// (it must be nonnegative since it is of unsigned type)
	index += MAX_URLS

	index = (index - 1) % MAX_URLS
	return
}
