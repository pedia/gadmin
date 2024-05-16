/*
Package reformism provides several utility functions for native text/template
*/
package gadmin

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"
)

// Pack represents packed arguments and original dot
type Pack struct {
	Origin any
	Args   map[string]any
}

// witharg is used in pipe to pack argument with dot
func witharg(k string, v any, i any) Pack {
	packT := reflect.TypeOf((*Pack)(nil)).Elem()
	if reflect.TypeOf(i) == packT {
		old := i.(Pack)
		old.Args[k] = v
		return old
	}
	return Pack{
		Origin: i,
		Args: map[string]any{
			k: v,
		},
	}

}

// done eats all pack passed to it and returns nil
func done(Pack) any {
	return nil
}

// args extracts .args field
func args(p Pack) map[string]any {
	return p.Args
}

// argCheckError may be raised in RequireArg
type argCheckError struct {
	detail string
}

// newArgCheckError returns a new ArgCheckError instance from detailed message
func newArgCheckError(s string) *argCheckError {
	return &argCheckError{
		detail: s,
	}
}

func (a argCheckError) Error() string {
	return a.detail
}

// requireArg accepts packed dot(Pack), checks its validity, then returns the dot
func requireArg(k string, trailingArgs ...any) (any, error) {
	if len(trailingArgs) != 1 && len(trailingArgs) != 2 {
		return nil, newArgCheckError(`Invalid format. requireArg parameterName ["typeName"]`)
	}
	v := trailingArgs[len(trailingArgs)-1]

	if v, ok := v.(Pack); ok { // check whether last arg is Pack
		if _, ok := v.Args[k]; !ok { // check whether Pack contains arguments with name K
			return nil, newArgCheckError(fmt.Sprintf("Required argument not found. Expected: %s, actual args: %v",
				k,
				v.Args))
		}
		if len(trailingArgs) == 2 { // check type
			if expectedTypeName, ok := trailingArgs[0].(string); ok {
				if reflect.TypeOf(v.Args[k]).Name() != expectedTypeName {
					return nil, newArgCheckError(fmt.Sprintf("Unmatched type: Expected: %s, actual: %s",
						expectedTypeName,
						reflect.TypeOf(v.Args[k]).Name()))
				}
			} else {
				return nil, newArgCheckError(fmt.Sprintf("The second argument of requireArg must be string! %v found",
					trailingArgs[0]))
			}
		}
		return trailingArgs[len(trailingArgs)-1], nil
	}
	return nil, newArgCheckError("requireArg didn't receive argument modified by withArg")
}

func makeSlice(args ...any) []any {
	return args
}

// mapError may be raised in MakeMap
type mapError struct {
	detail string
}

// newMapError returns a new ArgCheckError instance from detailed message
func newMapError(s string) *mapError {
	return &mapError{
		detail: s,
	}
}

func (a mapError) Error() string {
	return a.detail
}

func makeMap(args ...any) (map[string]any, error) {
	if len(args) < 2 {
		return nil, newMapError("arg num not required")
	}
	rawMap := make(map[string]any)
	if oldMap, ok := args[len(args)-1].(map[string]any); ok {
		rawMap = oldMap
		args = args[:len(args)-1]
	}

	if len(args)%2 != 0 {
		return nil, newMapError("arg should like key1 value1 key2 value2 ...")
	}
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			return nil, newMapError("key should be string")
		}
		rawMap[key] = args[i+1]
	}
	return rawMap, nil
}

func inRange(start, end, n int) bool {
	if start <= end {
		return n >= start && n < end
	}
	return n <= start && n > end
}

func makeSequence(args ...int) ([]int, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, errors.New("arg number to make range unsatisfied: 1-3 is acceptable")
	}
	var result []int
	if len(args) == 1 {
		for i := 0; i < args[0]; i++ {
			result = append(result, i)
		}
	} else {
		start := args[0]
		end := args[1]
		var step int
		if end >= start {
			step = 1
		} else {
			step = -1
		}
		if len(args) == 3 {
			step = args[2]
		}
		if step == 0 {
			return nil, errors.New("step=0 is illegal")
		}
		for i := start; inRange(start, end, i); i += step {
			result = append(result, i)
		}
	}
	return result, nil
}

func appendSlice(args ...any) ([]any, error) {
	if len(args) == 0 {
		return nil, errors.New("no arg found for appendSlice")
	}
	oldSlice := reflect.ValueOf(args[len(args)-1])
	if oldSlice.Kind() != reflect.Slice {
		return nil, errors.New("the last arg must be an slice")
	}
	slice := []any{}
	for i := 0; i < oldSlice.Len(); i++ {
		slice = append(slice, oldSlice.Index(i).Interface())
	}
	slice = append(slice, args[:len(args)-1]...)

	return slice, nil
}

func splitStr(sep, s string) []string {
	return strings.Split(s, sep)
}

func joinStr(sep string, a []string) string {
	return strings.Join(a, sep)
}

func only(args ...any) (map[string]any, error) {
	if len(args) < 2 {
		return nil, errors.New("filter need more args")
	}
	last := args[len(args)-1]
	pm, ok := last.(map[string]any)
	if !ok {
		pack, ok := last.(Pack)
		if !ok {
			return nil, errors.New("filter should be added to pipeline")
		}
		pm = pack.Args
	}

	res := map[string]any{}
	for i := 0; i < len(args)-1; i++ {
		k := args[i].(string)
		if v, ok := pm[k]; ok {
			res[k] = v
		}
	}
	return res, nil
}

func deleteItem(k string, args ...any) (map[string]any, error) {
	if len(args) < 1 {
		return nil, errors.New("delete need more args")
	}
	last := args[len(args)-1]
	pm, ok := last.(map[string]any)
	if !ok {
		pack, ok := last.(Pack)
		if !ok {
			return nil, errors.New("delete should be added to pipeline")
		}
		pm = pack.Args
	}

	delete(pm, k)
	return pm, nil
}

// {{ add "key" "val"}}
func mapSet(k string, v any, args ...any) (map[string]any, error) {
	if len(args) != 1 {
		return nil, errors.New("add need more args")
	}
	last := args[len(args)-1]
	pm, ok := last.(map[string]any)
	if !ok {
		pack, ok := last.(Pack)
		if !ok {
			return nil, errors.New("add should be added to pipeline")
		}
		pm = pack.Args
	}

	pm[k] = v
	return pm, nil
}

// {{ default . "key" "value" }}
func defaultOf(c any, k string, dv any) (any, error) {
	m, ok := c.(map[string]any)
	if !ok {
		pack, ok := c.(Pack)
		if !ok {
			return nil, errors.New("a map of pack descired before pipeline")
		}
		m = pack.Args
	}
	if v, ok := m[k]; ok {
		return v, nil
	}
	return dv, nil
}

func logto(format string, v ...any) string {
	log.Printf(format, v...)
	return ""
}

// Funcs is a FuncMap which can be passed as argument of .Func of text/template
var Funcs = template.FuncMap{
	"arg":     witharg,
	"require": requireArg,
	"done":    done,
	"args":    args,
	"slice":   makeSlice,
	"map":     makeMap,
	"seq":     makeSequence,
	"append":  appendSlice,
	"split":   splitStr,
	"join":    joinStr,
	"only":    only,
	"delete":  deleteItem,
	"set":     mapSet,
	"default": defaultOf,
	"log":     logto,
	// sprig add return int64
	"add": func(is ...int) int {
		var a int = 0
		for _, b := range is {
			a += b
		}
		return a
	},
	"sub": func(a, b int) int { return a - b },
}
