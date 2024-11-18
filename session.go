package gadmin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/samber/lo"
)

type Session struct {
	*Secret
	Values     map[string]any
	cookieName string
	saved      bool
}

func (S *Session) Fetch(name string) any {
	if v, ok := S.Values[name]; ok {
		delete(S.Values, name)
		return v
	}
	return nil
}
func (S *Session) Get(name string) any {
	return S.Values[name]
}
func (S *Session) Del(name string) *Session {
	delete(S.Values, name)
	return S
}
func (S *Session) Set(name string, val any) *Session {
	S.Values[name] = val
	return S
}

func (S *Session) Save(w http.ResponseWriter) error {
	if S.saved {
		panic("saved again?")
	}

	log.Printf("session save: %s", strings.Join(lo.Keys(S.Values), ","))

	bs, err := json.Marshal(S.Values)
	if err != nil {
		return err
	}

	dest := make([]byte, hex.EncodedLen(len(bs)))
	hex.Encode(dest, bs)

	// cookie := http.Cookie{
	// 	Name:  S.cookieName,
	// 	Value: S.Sign(dest),
	// }
	// w.Header().Set("Set-Cookie", cookie.String())

	// Maybe replace is better
	http.SetCookie(w, &http.Cookie{
		Name:  S.cookieName,
		Value: S.Sign(dest),
	})
	S.saved = true

	return nil
}

type _ctxkey int

var _sessionKey _ctxkey

func CurrentSession(r *http.Request) *Session {
	return must[*Session](r.Context().Value(_sessionKey).(*Session))
}

// Get or Create Session from current Request
func PatchSession(r *http.Request, admin *Admin) *http.Request {
	ns := Session{
		Secret:     admin.secret,
		Values:     map[string]any{},
		cookieName: "session", // TODO: admin.config
	}
	ns.ReadFrom(r)
	return r.Clone(context.WithValue(r.Context(), _sessionKey, &ns))
}

func (S *Session) ReadFrom(r *http.Request) error {
	c, err := r.Cookie(S.cookieName)
	if err != nil {
		return err
	}

	hexed_bytes, err := S.Unsign(c.Value)
	if err != nil {
		return err
	}

	json_bytes := make([]byte, hex.DecodedLen(len(hexed_bytes)))
	_, err = hex.Decode(json_bytes, hexed_bytes)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(json_bytes, &S.Values); err != nil {
		return err
	}
	return nil
}
