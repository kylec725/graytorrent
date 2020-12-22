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

func main() {
    var one Bitfield = make([]byte, 8)
    var two Bitfield = []byte{0b01010101, 0b10101010}
    for i := 0; i < 8; i++ {
        one[i] |= 0xff
    }
    one[1] &= 0xff >> 1
    one[3] &= 0xff >> 1
    one.Print()
    fmt.Println(one.Has(0))
    fmt.Println(one.Has(8))
    one.Set(3)
    one.Print()
    two.Print()
}
