package connect

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

const debugConnect = true

func TestGetOpenPort(t *testing.T) {
    assert := assert.New(t)

    port, err := GetOpenPort([]int{ 6881, 6889 })
    assert.Nil(err)

    if debugConnect {
        fmt.Println("Got open port:", port)
    }
}
