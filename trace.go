package gadm

import (
	"bytes"
	"container/list"
	"iter"
	"strings"

	"fmt"
	"net/http"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// During http request, some gorm SQLs will be excuted.
// Record as trace.

type Entry struct {
	URL string
	Log string
}

type Trace struct {
	m        sync.RWMutex
	entries  *list.List
	buf      *bytes.Buffer
	logger   logger.Interface
	MaxCount int
}

// Trace gorm sqls during a http.Request
func NewTrace() *Trace {
	t := &Trace{
		buf:      bytes.NewBuffer(make([]byte, 0, 1000)),
		entries:  list.New(),
		MaxCount: 100,
	}

	t.logger = logger.New(t, logger.Config{
		LogLevel:                  logger.Info,
		Colorful:                  false,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      false,
	})
	return t
}

func (t *Trace) Trace(db *gorm.DB) {
	if db.Logger != t.logger {
		// set logger level
		db.Config.Logger = logger.Default.LogMode(logger.Info)

		// replace logger to mine
		db.Logger = t.logger
	}
}

// collect once, should call in ServeHTTP
//
//	func ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	  mux.ServeHTTP(w, r)
//	  defer trace.CollectOnce(r)
//	}
func (t *Trace) CollectOnce(r *http.Request) {
	t.m.Lock()
	defer t.m.Unlock()

	if t.buf.Len() == 0 {
		return
	}

	// TODO: remove this
	text := strings.ReplaceAll(t.buf.String(), "/Users/mord/t/", "")
	t.entries.PushBack(Entry{r.URL.String(), text})

	if t.entries.Len() > t.MaxCount {
		t.entries.Remove(t.entries.Front())
	}
	t.buf.Reset()
}

func (t *Trace) CheckTrace(r *http.Request) {
	if t != nil {
		t.CollectOnce(r)
	}
}

func (t *Trace) Printf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	fmt.Fprintf(t.buf, format+"\n", args...)
}

// light loop with lock
func (t *Trace) Entries() iter.Seq[Entry] {
	t.m.RLock()
	defer t.m.RUnlock()

	return func(yield func(Entry) bool) {
		for e := t.entries.Back(); e != nil; e = e.Prev() {
			pair := e.Value.(Entry)
			if !yield(pair) {
				break
			}
		}
	}
}
