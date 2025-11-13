package gadm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	is := assert.New(t)

	g := NewGenerator("sqlite:examples/sqla/sample.db")

	err := g.Run(nil, os.Stdout)
	is.Nil(err)
}
