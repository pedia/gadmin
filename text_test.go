package gadmin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/leonelquinteros/gotext.v1"
)

func TestText(t *testing.T) {
	is := assert.New(t)

	gotext.Configure("translations", "zh_Hant_TW", "admin")
	is.Equal("首頁", gotext.Get("Home"))
	is.Equal(`檔案 "foo" 已經存在。`, gotext.Get(`File "%s" already exists.`, "foo"))
}
