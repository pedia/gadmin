package gadmin

import (
	"context"
	"net/http"

	"github.com/samber/lo"
)

// Mock flashed message in Flask
func newFlashed() *_Flashed {
	return &_Flashed{messages: []_flashedMessage{}}
}

type _flashedMessage struct {
	Category string // In template, category should be: info, danger, success
	Data     any
}

func (f *_flashedMessage) dict() map[string]any {
	return map[string]any{
		"category": f.Category,
		"data":     f.Data,
	}
}

type _Flashed struct {
	messages []_flashedMessage
}

func (F *_Flashed) Add(data any, category ...string) {
	F.messages = append(F.messages, _flashedMessage{
		Data:     data,
		Category: firstOr(category, "info"),
	})
}

// `category` is category filter
func (F *_Flashed) GetMessages(category ...string) []map[string]any {
	dictOf := func(ms []_flashedMessage) []map[string]any {
		return lo.Map(ms, func(m _flashedMessage, _ int) map[string]any {
			return m.dict()
		})
	}

	if len(category) > 0 {
		return dictOf(lo.Filter(F.messages, func(m _flashedMessage, _ int) bool {
			return m.Category == category[0]
		}))
	}
	return dictOf(F.messages)
}

type _key int

var _flashedKey _key

// Inject Flashed into cloned Request
func PatchFlashed(r *http.Request) *http.Request {
	return r.Clone(context.WithValue(r.Context(), _flashedKey, newFlashed()))
}

func FlashedFrom(r *http.Request) *_Flashed {
	return must[*_Flashed](r.Context().Value(_flashedKey).(*_Flashed))
}

// `Flash` Mock flask.flash
// `get_flashed_messages` as `FlashedFrom(r).GetMessages()`
func Flash(r *http.Request, data any, category ...string) {
	FlashedFrom(r).Add(data, category...)
}
