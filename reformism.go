/*
Package reformism provides several utility functions for native text/template
*/
package gadmin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

// Pack represents packed arguments and original dot
type Pack struct {
	Origin interface{}
	Args   map[string]interface{}
}

// Witharg is used in pipe to pack argument with dot
func Witharg(k string, v interface{}, i interface{}) Pack {
	packT := reflect.TypeOf((*Pack)(nil)).Elem()
	if reflect.TypeOf(i) == packT {
		old := i.(Pack)
		old.Args[k] = v
		return old
	}
	return Pack{
		Origin: i,
		Args: map[string]interface{}{
			k: v,
		},
	}

}

// Done eats all pack passed to it and returns nil
func Done(Pack) interface{} {
	return nil
}

// Args extracts .Args field
func Args(p Pack) map[string]interface{} {
	return p.Args
}

// ArgCheckError may be raised in RequireArg
type ArgCheckError struct {
	detail string
}

// NewArgCheckError returns a new ArgCheckError instance from detailed message
func NewArgCheckError(s string) *ArgCheckError {
	return &ArgCheckError{
		detail: s,
	}
}

func (a ArgCheckError) Error() string {
	return a.detail
}

// RequireArg accepts packed dot(Pack), checks its validity, then returns the dot
func RequireArg(k string, trailingArgs ...interface{}) (interface{}, error) {
	if len(trailingArgs) != 1 && len(trailingArgs) != 2 {
		return nil, NewArgCheckError(`Invalid format. requireArg parameterName ["typeName"]`)
	}
	v := trailingArgs[len(trailingArgs)-1]

	if v, ok := v.(Pack); ok { // check whether last arg is Pack
		if _, ok := v.Args[k]; !ok { // check whether Pack contains arguments with name K
			return nil, NewArgCheckError(fmt.Sprintf("Required argument not found. Expected: %s, actual args: %v",
				k,
				v.Args))
		}
		if len(trailingArgs) == 2 { // check type
			if expectedTypeName, ok := trailingArgs[0].(string); ok {
				if reflect.TypeOf(v.Args[k]).Name() != expectedTypeName {
					return nil, NewArgCheckError(fmt.Sprintf("Unmatched type: Expected: %s, actual: %s",
						expectedTypeName,
						reflect.TypeOf(v.Args[k]).Name()))
				}
			} else {
				return nil, NewArgCheckError(fmt.Sprintf("The second argument of requireArg must be string! %v found",
					trailingArgs[0]))
			}
		}
		return trailingArgs[len(trailingArgs)-1], nil
	}
	return nil, NewArgCheckError("requireArg didn't receive argument modified by withArg")
}

func MakeSlice(args ...interface{}) []interface{} {
	return args
}

// MapError may be raised in MakeMap
type MapError struct {
	detail string
}

// NewMapError returns a new ArgCheckError instance from detailed message
func NewMapError(s string) *MapError {
	return &MapError{
		detail: s,
	}
}

func (a MapError) Error() string {
	return a.detail
}

func MakeMap(args ...interface{}) (map[string]interface{}, error) {
	if len(args) < 2 {
		return nil, NewMapError("arg num not required")
	}
	rawMap := make(map[string]interface{})
	if oldMap, ok := args[len(args)-1].(map[string]interface{}); ok {
		rawMap = oldMap
		args = args[:len(args)-1]
	}

	if len(args)%2 != 0 {
		return nil, NewMapError("arg should like key1 value1 key2 value2 ...")
	}
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			return nil, NewMapError("key should be string")
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

func MakeRange(args ...int) ([]int, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, errors.New("Arg number to make range unsatisfied: 1-3 is acceptable")
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

func AppendSlice(args ...interface{}) ([]interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("No arg found for appendSlice")
	}
	oldSlice := reflect.ValueOf(args[len(args)-1])
	if oldSlice.Kind() != reflect.Slice {
		return nil, errors.New("The last arg must be an slice")
	}
	slice := []interface{}{}
	for i := 0; i < oldSlice.Len(); i++ {
		slice = append(slice, oldSlice.Index(i).Interface())
	}
	for _, v := range args[:len(args)-1] {
		slice = append(slice, v)
	}

	return slice, nil
}

func SplitStr(sep, s string) []string {
	return strings.Split(s, sep)
}

func joinStr(sep string, a []string) string {
	return strings.Join(a, sep)
}

// FuncsText is a FuncMap which can be passed as argument of .Func of text/template
var FuncsText = template.FuncMap{
	"arg":     Witharg,
	"require": RequireArg,
	"done":    Done,
	"args":    Args,
	"slice":   MakeSlice,
	"map":     MakeMap,
	"rng":     MakeRange,
	"append":  AppendSlice,
	"split":   SplitStr,
	"join":    joinStr,
}

// FuncsHTML is a FuncMap which can be passed as argument of .Func of html/template
var FuncsHTML = FuncsText
