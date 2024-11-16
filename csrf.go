package gadmin

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// CSRF
func NewCSRF(session *Session) *CSRF {
	return &CSRF{
		Session: session,
		timeout: time.Duration(1) * time.Hour,
		fnow:    time.Now,
	}
}

type CSRF struct {
	*Session
	timeout time.Duration
	fnow    func() time.Time // for unit test
}

func (C *CSRF) GenerateToken() string {
	buf := hex.EncodeToString(C.rand(8))
	now := C.fnow()

	C.Set(buf, now)

	const stampLayout = "20060102150405"
	return now.Format(stampLayout) + "#" + buf
}

var errTokenExpired = errors.New("token expired")
var errTokenInvalid = errors.New("token invalid")

func (C *CSRF) Validate(token string) error {
	arr := strings.Split(token, "#")
	if len(arr) != 2 {
		return errTokenInvalid
	}
	sv := C.Fetch(arr[1])
	if sv == nil {
		return errTokenInvalid
	}

	when, ok := sv.(time.Time)
	if !ok {
		return errTokenInvalid
	}

	d := C.fnow().Sub(when)
	if d < 0 || d > C.timeout {
		return errTokenExpired
	}
	return nil
}

func (C *CSRF) rand(size int) []byte {
	bs := make([]byte, size)
	rand.Read(bs)
	return bs
}
