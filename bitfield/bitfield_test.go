package bitfield

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestHas(t *testing.T) {
    assert := assert.New(t)
    var bf Bitfield = []byte{0b01010101, 0b10101010, 0b11001100}
    // Check first byte
    for i, want := 0, false; i < 8; i++ {
        assert.Equal(want, bf.Has(i), fmt.Sprintf("Got wrong answer checking if we have piece %d", i))
        want = !want
    }
    // Check second byte
    for i, want := 8, true; i < 16; i++ {
        assert.Equal(want, bf.Has(i), fmt.Sprintf("Got wrong answer checking if we have piece %d", i))
        want = !want
    }
    // Check third byte
    for i, want := 16, true; i < 24; i += 2 {
        assert.Equal(want, bf.Has(i), fmt.Sprintf("Got wrong answer checking if we have piece %d", i))
        assert.Equal(want, bf.Has(i + 1), fmt.Sprintf("Got wrong answer checking if we have piece %d", i))
        want = !want
    }
}

func TestSet(t *testing.T) {
    assert := assert.New(t)
    var bf Bitfield = []byte{0b00000000, 0b11111111, 0b00000000}
    for i := 0; i < 24; i++ {
        bf.Set(i)
    }
    for i, want := 0, true; i < 24; i++ {
        assert.Equal(want, bf.Has(i), fmt.Sprintf("Error setting bit at index %d", i))
    }
}
