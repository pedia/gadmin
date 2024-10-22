package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	is := assert.New(t)

	is.Equal("a", firstOrEmpty("a", "b"))
	is.Equal("", firstOrEmpty[string]())
	is.Equal(0, firstOrEmpty[int]())
}

func TestPairToQuery(t *testing.T) {
	is := assert.New(t)

	is.Equal("a=1", pairToQueryString("a", "1"))
	is.Equal("a=1&a=2", pairToQueryString("a", "1", "a", "2"))
	is.Equal("a=+", pairToQueryString("a", " "))

	// abnormal input
	is.Equal("", pairToQueryString("a"))
	is.Equal("a=b", pairToQueryString("a", "b", "c"))
}

func TestXxx(t *testing.T) {
	is := assert.New(t)

	is.Equal([]string{"a", "b"}, strings.SplitN("a.b", ".", 2))
	is.Equal([]string{"a", "b.c"}, strings.SplitN("a.b.c", ".", 2))
	is.Equal([]string{"", "b"}, strings.SplitN(".b", ".", 2))
}
