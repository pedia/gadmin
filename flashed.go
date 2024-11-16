package gadmin

import (
	"net/http"

	"github.com/samber/lo"
)

// Mock flashed message in Flask
func flashed(session *Session) *_Flashed {
	return &_Flashed{Session: session, name: " flashed"}
}

type _message map[string]any

type _Flashed struct {
	*Session
	name string // session name
}

func (F *_Flashed) Add(data any, category ...string) {
	ms := append(F.get(), _message(map[string]any{
		"data":     data,
		"category": firstOr(category, "info"),
	}))
	F.Set(F.name, ms)
}

func (F *_Flashed) get() []map[string]any {
	ms := []map[string]any{}
	if ma := F.Session.Fetch(F.name); ma != nil {
		ms, _ = ma.([]map[string]any)
	}
	return ms
}

// `category` is category filter
func (F *_Flashed) GetMessages(category ...string) []map[string]any {
	if len(category) > 0 {
		return lo.Filter(F.get(), func(m map[string]any, _ int) bool {
			return m["category"] == category[0]
		})

		// TODO: put not-matched-message back
	}
	return F.get()
}

// `Flash` Mock flask.flash
// `get_flashed_messages` as `FlashedFrom(r).GetMessages()`
// `bootstrap` affect category: info, success, danger
// https://getbootstrap.com/docs/4.0/components/alerts/
func Flash(r *http.Request, data any, category ...string) {
	flashed(CurrentSession(r)).Add(data, category...)
}

func GetFlashedMessages(r *http.Request, category ...string) []map[string]any {
	return flashed(CurrentSession(r)).GetMessages(category...)
}
