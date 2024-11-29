package gadmin

import (
	"github.com/spf13/cast"
)

type Config struct {
	dict map[string]any
}

// config.Bool("debug.verbose", true)
// net.Dial(config.String("xx.address", "10.0.0.1:3389"))
var config = &Config{dict: map[string]any{}}

func (c *Config) Bool(name string, default_value ...bool) bool {
	if v, ok := c.dict[name]; ok {
		return cast.ToBool(v)
	}
	return firstOr(default_value)
}

func (c *Config) String(name string, default_value ...string) string {
	if v, ok := c.dict[name]; ok {
		return cast.ToString(v)
	}
	return firstOr(default_value)
}

func (c *Config) Int(name string, default_value ...int) int {
	if v, ok := c.dict[name]; ok {
		return cast.ToInt(v)
	}
	return firstOr(default_value)
}

func (c *Config) Put(name string, v any) {
	c.dict[name] = v
}

// prefix.name => value
// return name => value
func (c *Config) GetSection(section_name string) map[string]any {
	// prefix := section_name + "."
	// res := map[string]string{}

	// for _, k := range c.dict.Keys() {
	// 	if strings.HasPrefix(k, prefix) {
	// 		key_without_prefix := k[len(prefix):]
	// 		res[key_without_prefix] = c.dict.Get(k)
	// 	}
	// }
	return nil
}
