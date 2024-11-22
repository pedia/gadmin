package gadmin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type Secret struct {
	key []byte
}

func NewSecret(secret string) *Secret {
	h := sha256.New()
	h.Write([]byte(secret))
	return &Secret{key: h.Sum(nil)}
}

func (S *Secret) hash(src []byte) []byte {
	h := hmac.New(sha256.New, S.key)
	h.Write(src)
	return h.Sum(nil)
}

// Generate string like: {src}.{hash of src}
func (S *Secret) Sign(src []byte, sep ...string) string {
	dest := S.hash(src)
	return string(src) + firstOr(sep, ".") + base64.RawURLEncoding.EncodeToString(dest)
}

func (S *Secret) Unsign(s string, sep ...string) ([]byte, error) {
	arr := strings.Split(s, firstOr(sep, "."))
	if len(arr) != 2 && len(arr[1]) != 64 {
		return nil, errInvalid
	}

	hashed_input, err := base64.RawURLEncoding.DecodeString(arr[1])
	if err != nil {
		return nil, errInvalid
	}

	hashed := S.hash([]byte(arr[0]))
	if !hmac.Equal(hashed, hashed_input) {
		return nil, errInvalid
	}
	return []byte(arr[0]), nil
}
