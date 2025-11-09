package gadmin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
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

	bs, err := json.Marshal(S.Values)
	if err != nil {
		return err
	}

	dest := make([]byte, hex.EncodedLen(len(bs)))
	hex.Encode(dest, bs)

	cookie := http.Cookie{
		Name:     S.cookieName,
		Value:    S.Sign(dest),
		Path:     "/",
		HttpOnly: true,
	}
	v := cookie.String()
	if v == "" {
		panic("cookie invalid")
	}
	w.Header().Add("Set-Cookie", v)
	S.saved = true

	return nil
}

type _ctxkey int

var _sessionKey _ctxkey

func CurrentSession(r *http.Request) *Session {
	// TODO:
	// return must(r.Context().Value(_sessionKey).(*Session))
	return &Session{
		Secret:     NewSecret("TODO"),
		Values:     map[string]any{},
		cookieName: "session", // TODO: admin.config
	}
}

// Get or Create Session from current Request
func PatchSession(r *http.Request, admin *Admin) *http.Request {
	ns := Session{
		Secret:     admin.secret,
		Values:     map[string]any{},
		cookieName: "session", // TODO: admin.config
	}
	if err := ns.ReadFrom(r); err != nil {
		// log.Printf("session %s read %s, got %d", r.URL, err, len(ns.Values))
	}

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
