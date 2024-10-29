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
	is.Equal([]map[string]interface{}{
		{"category": "info", "data": 42},
		{"category": "error", "data": 34}},
		f3.GetMessages())
	is.Len(f3.GetMessages("info"), 1)
	is.Len(f3.GetMessages("error"), 1)
}
