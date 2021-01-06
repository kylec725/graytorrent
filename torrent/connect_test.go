package torrent

import (
    "testing"
    "fmt"

    "github.com/stretchr/testify/assert"
)

func TestGetOpenPort(t *testing.T) {
    assert := assert.New(t)

    port, err := getOpenPort()
    assert.Nil(err)
    fmt.Println("Got open port:", port)
}
