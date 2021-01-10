package connect

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestGetOpenPort(t *testing.T) {
    assert := assert.New(t)

    port, err := GetOpenPort([2]int{ 6881, 6889 })
    assert.Nil(err)
    fmt.Println("Got open port:", port)
}
