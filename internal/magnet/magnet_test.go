package magnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMagnetUnmarshal(t *testing.T) {
	assert := assert.New(t)
	link := "magnet:?xt=urn:btih:<info-hash>&dn=<name>&tr=<tracker-url>&x.pe=<peer-address>"

	_, err := Unmarshal(link)
	assert.Nil(err)
}
