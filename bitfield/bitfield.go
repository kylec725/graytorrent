package bitfield

import (
    "fmt"
)

type Bitfield []byte

func (bf Bitfield) Print() {
    for _, n := range(bf) {
        fmt.Printf("%08b ", n)
    }
    fmt.Printf("\n")
}

func (bf Bitfield) Has(index int) bool {
    byteIndex := index / 8
    offset := index % 8
    return bf[byteIndex] >> (7 - offset) & 1 == 1
}

func (bf Bitfield) Set(index int) {
    byteIndex := index / 8
    offset := index % 8
    bf[byteIndex] |= 1 << (7 - offset)
}
