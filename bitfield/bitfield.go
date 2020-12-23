package bitfield

import (
	"fmt"
)

// Bitfield tracks the pieces of a torrent that one has
type Bitfield []byte

// Print the bitfield in a human-readable form
func (bf Bitfield) Print() {
	for _, n := range bf {
		fmt.Printf("%08b ", n)
	}
	fmt.Printf("\n")
}

// Has checks if we have a piece
func (bf Bitfield) Has(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	return bf[byteIndex]>>(7-offset)&1 == 1
}

// Set a bit to indicate a new piece
func (bf Bitfield) Set(index int) {
	byteIndex := index / 8
	offset := index % 8
	bf[byteIndex] |= 1 << (7 - offset)
}
