package bitfield

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHas(t *testing.T) {
	assert := assert.New(t)

	var bf Bitfield = []byte{0b01010101, 0b10101010, 0b11001100}
	// Check first byte
	for i, want := 0, false; i < 8; i++ {
		assert.Equal(want, bf.Has(i), "Got wrong answer checking if we have piece "+strconv.Itoa(i))
		want = !want
	}
	// Check second byte
	for i, want := 8, true; i < 16; i++ {
		assert.Equal(want, bf.Has(i), "Got wrong answer checking if we have piece "+strconv.Itoa(i))
		want = !want
	}
	// Check third byte
	for i, want := 16, true; i < 24; i += 2 {
		assert.Equal(want, bf.Has(i), "Got wrong answer checking if we have piece "+strconv.Itoa(i))
		assert.Equal(want, bf.Has(i+1), "Got wrong answer checking if we have piece "+strconv.Itoa(i))
		want = !want
	}
}

func TestSet(t *testing.T) {
	assert := assert.New(t)

	var bf Bitfield = []byte{0b00000000, 0b11111111, 0b00000000}
	// Set all bits
	for i := 0; i < 24; i++ {
		bf.Set(i)
	}
	// Checking has for every index should return true
	for i := 0; i < 24; i++ {
		assert.Equal(true, bf.Has(i), "Error setting bit at index "+strconv.Itoa(i))
	}
}
