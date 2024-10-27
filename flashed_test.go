package gadmin

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlashed(t *testing.T) {
	is := assert.New(t)

	r, _ := http.NewRequest("GET", "/", nil)
	r2 := PatchFlashed(r)

	f2 := FlashedFrom(r2)
	f2.Add(42)
	f2.Add(34, "error")

	f3 := FlashedFrom(r2)
	is.Equal([]any{42, 34}, f3.GetMessages())
	is.Equal([]any{42}, f3.GetMessages("message"))
	is.Equal([]any{34}, f3.GetMessages("error"))
}
