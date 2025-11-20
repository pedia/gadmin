package gadm

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	is := assert.New(t)

	g := NewGenerator("sqlite:examples/sqla/sample.db")

	err := g.Run(nil, io.Discard)
	is.Nil(err)
}
