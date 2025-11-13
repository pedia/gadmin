package gadm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	is := assert.New(t)
	is.NotNil(config)

	config.Put("word", "foo")
	config.Put("bool", true)
	config.Put("int", 42)

	is.Equal("foo", config.String("word"))
	is.Equal("foo", config.String("word", "bar"))
	is.Equal("bar", config.String("world", "bar"))

	is.Equal(42, config.Int("int"))
	is.Equal(33, config.Int("nint", 33))

	is.Equal(true, config.Bool("bool"))
	is.Equal(false, config.Bool("nbool"))
}
