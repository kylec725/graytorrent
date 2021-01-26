package connect

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugConnect = false

func TestOpenPort(t *testing.T) {
    assert := assert.New(t)

    port, err := OpenPort([]int{ 6881, 6889 })
    assert.Nil(err)

    if debugConnect {
        fmt.Println("Got open port:", port)
    }
}

func TestPortFromAddr(t *testing.T) {
    assert := assert.New(t)

    port, err := PortFromAddr("23493:5000")
    if assert.Nil(err) {
        assert.Equal(uint16(5000), port)
    }
}
