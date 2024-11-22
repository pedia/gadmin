package gadmin

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	is := assert.New(t)

	admin := NewAdmin("Foo", nil)

	r := httptest.NewRequest("GET", "/foo", nil)
	r1 := PatchSession(r, admin)
	s := CurrentSession(r1)
	is.NotNil(s)

	w1 := httptest.NewRecorder()
	s.Set("pai", 3.14)
	is.Nil(s.Save(w1))
	is.Equal(3.14, s.Get("pai"))

	c1 := readCookies(w1.Header(), s.cookieName)[0]
	is.NotNil(c1)

	r2 := httptest.NewRequest("GET", "/foo", nil)
	r2.Header.Add("cookie", fmt.Sprintf(`%s=%s`, s.cookieName, c1.Value))
	r2 = PatchSession(r2, admin)
	s2 := CurrentSession(r2)
	is.Equal(3.14, s2.Get("pai"))

	// test Sign/Unsign
	plain := []byte("blue sky")
	ss := s2.Sign(plain)
	src, err := s2.Unsign(ss)
	is.Nil(err)
	is.Equal(plain, src)

	// CSRF
	c := NewCSRF(s2)

	// generate csrf token
	c.fnow = func() time.Time { return time.Date(2024, 10, 30, 20, 0, 0, 0, time.Local) }
	t1 := c.GenerateToken()
	c.fnow = func() time.Time { return time.Date(2024, 10, 30, 20, 10, 0, 0, time.Local) }
	er1 := c.Validate(t1)
	is.Nil(er1)

	c.fnow = func() time.Time { return time.Date(2024, 10, 30, 20, 0, 0, 0, time.Local) }
	t2 := c.GenerateToken()
	c.fnow = func() time.Time { return time.Date(2024, 10, 30, 21, 0, 1, 0, time.Local) }
	er2 := c.Validate(t2)
	is.Equal(errExpired, er2)

	is.Equal(errInvalid, c.Validate("a#b"))

	// Flash
	Flash(r2, 42)
	Flash(r2, 34, "error")

	f3 := flashed(s2)

	is.Equal([]map[string]interface{}{
		{"category": "info", "data": 42},
		{"category": "error", "data": 34}},
		f3.GetMessages())
	// is.Len(f3.GetMessages("info"), 1)
	// is.Len(f3.GetMessages("error"), 1)

	// is.Equal([]map[string]interface{}{
	// 	{"category": "info", "data": 42},
	// 	{"category": "error", "data": 34}},
	// 	GetFlashedMessages(r2))

	// is.Equal([]map[string]interface{}{
	// 	{"category": "info", "data": 42}},
	// 	GetFlashedMessages(r2, "info"))
}
