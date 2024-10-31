package gadmin

import (
	"net/http"

	"github.com/samber/lo"
)

// Mock flashed message in Flask
func flashed(session *Session) *_Flashed {
	return &_Flashed{Session: session, name: " flashed"}
}

type _message struct {
	Category string // In template, category should be: info, danger, success
	Data     any
}

func (f *_message) dict() map[string]any {
	return map[string]any{
		"category": f.Category,
		"data":     f.Data,
	}
}

type _Flashed struct {
	*Session
	name string
}

func (F *_Flashed) get() []_message {
	ms := []_message{}
	if ma := F.Get(F.name); ma != nil {
		ms = ma.([]_message)
	}
	return ms
}

func (F *_Flashed) Add(data any, category ...string) {
	ms := append(F.get(), _message{
		Data:     data,
		Category: firstOr(category, "info"),
	})
	F.Set(F.name, ms)
}

// `category` is category filter
func (F *_Flashed) GetMessages(category ...string) []map[string]any {
	dictOf := func(ms []_message) []map[string]any {
		return lo.Map(ms, func(m _message, _ int) map[string]any {
			return m.dict()
		})
	}

	if len(category) > 0 {
		return dictOf(lo.Filter(F.get(), func(m _message, _ int) bool {
			return m.Category == category[0]
		}))
	}
	return dictOf(F.get())
}

// `Flash` Mock flask.flash
// `get_flashed_messages` as `FlashedFrom(r).GetMessages()`
func Flash(r *http.Request, data any, category ...string) {
	flashed(CurrentSession(r)).Add(data, category...)
}

func GetFlashedMessages(r *http.Request, category ...string) []map[string]any {
	return flashed(CurrentSession(r)).GetMessages(category...)
}
