package gadmin

import (
	"net/url"
	"reflect"
	"strconv"
)

func first_or_empty[T any](as ...T) T {
	if len(as) > 0 {
		return as[0]
	}
	var t T
	return t
}

func must[T any](xs ...any) T {
	// try return with (x, error)
	err, ok := xs[len(xs)-1].(error)
	if ok && err != nil {
		panic(err)
	}

	if !ok {
		// try return with (x, bool)
		if b, ok := xs[len(xs)-1].(bool); ok && !b {
			panic("not ok")
		}
	}

	return xs[0].(T)
}

func map_into_values(m map[string]any) url.Values {
	values := url.Values{}
	for key, value := range m {
		// Convert value to string based on its type
		var strVal string
		switch reflect.TypeOf(value).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strVal = strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strVal = strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
		case reflect.Float32, reflect.Float64:
			strVal = strconv.FormatFloat(reflect.ValueOf(value).Float(), 'f', -1, 64) // Use -1 for automatic precision
		case reflect.Bool:
			strVal = strconv.FormatBool(reflect.ValueOf(value).Bool())
		case reflect.String:
			strVal = reflect.ValueOf(value).String()
		default:
			// Handle any other types as needed, or skip if unsupported
			continue
		}
		values.Set(key, strVal)
	}
	return values
}
