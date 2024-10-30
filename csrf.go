package gadmin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// CSRF
func NewCSRF(secret string) *CSRF {
	h := sha256.New()
	h.Write([]byte(secret))
	return &CSRF{
		secret:     h.Sum(nil),
		timeout:    time.Duration(1) * time.Hour,
		fnow:       time.Now,
		cookieName: "gadmin_csrf"}
}

type CSRF struct {
	secret     []byte
	timeout    time.Duration
	fnow       func() time.Time // for unit test
	cookieName string
}

func (C *CSRF) GenerateToken() string {
	// token: time+rand.hash(time+rand)
	src := C.fnow().Format(stampLayout) + hex.EncodeToString(C.rand(8))

	return C.Sign([]byte(src), "#")
}

var errExpired = errors.New("token expired")
var errInvalid = errors.New("token invalid")

func (C *CSRF) Validate(token string) error {
	if src, err := C.Unsign(token, "#"); err == nil {
		stamp := src[:len(stampLayout)]
		if when, err := time.ParseInLocation(stampLayout, string(stamp), time.Local); err == nil {
			d := C.fnow().Sub(when)
			if d > 0 && d < C.timeout {
				return nil
			}
			return errExpired
		}
	}
	return errInvalid
}

const stampLayout = "20060102150405"

func (C *CSRF) rand(size int) []byte {
	bs := make([]byte, size)
	rand.Read(bs)
	return bs
}

func (C *CSRF) hash(src []byte) []byte {
	h := hmac.New(sha256.New, C.secret)
	h.Write(src)
	return h.Sum(nil)
}

// Generate string like: {src}.{hash of src}
func (C *CSRF) Sign(src []byte, sep ...string) string {
	dest := C.hash(src)
	return string(src) + firstOr(sep, ".") + base64.RawURLEncoding.EncodeToString(dest)
}

func (C *CSRF) Unsign(s string, sep ...string) ([]byte, error) {
	arr := strings.Split(s, firstOr(sep, "."))
	if len(arr) != 2 && len(arr[1]) != 64 {
		return nil, errInvalid
	}

	hashed_input, err := base64.RawURLEncoding.DecodeString(arr[1])
	if err != nil {
		return nil, errInvalid
	}

	hashed := C.hash([]byte(arr[0]))
	if !hmac.Equal(hashed, hashed_input) {
		return nil, errInvalid
	}
	return []byte(arr[0]), nil
}
