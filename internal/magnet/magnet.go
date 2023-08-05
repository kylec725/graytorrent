/*
Package magnet provides functionality to extract metainfo about
torrents from magnet links.
*/
package magnet

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const magnetStr = "magnet"
const urn = "urn"
const btih = "btih"

// Errors
var (
	ErrNotMagnet = errors.New("Link provided does not have the magnet schema")
	ErrXT        = errors.New("Magnet link must provide a valid xt value")
)

// Magnet is a struct holding the metadata from a torrent magnet link
type Magnet struct {
	InfoHash [20]byte
	dn       string // display name
	tr       string
	xpe      string
}

// New unpacks a magnet link string
func New(s string) (Magnet, error) {
	var m Magnet
	fmt.Println("link:", s)

	u, err := url.Parse(s)
	if err != nil {
		return Magnet{}, err
	}

	if u.Scheme != magnetStr {
		return Magnet{}, errors.Wrap(ErrNotMagnet, "Unmarshal")
	}

	fmt.Println("query:", u.RawQuery)
	q := u.Query()

	// xt is a required value
	if _, ok := q["xt"]; !ok {
		return Magnet{}, errors.Wrap(ErrXT, "Unmarshal")
	}
	xt := strings.Split(q["xt"][0], ":")
	fmt.Println("xt:", xt)

	if xt[0] != urn || xt[1] != btih {
		return Magnet{}, errors.Wrap(ErrXT, "Unmarshal")
	}

	return m, nil
}
