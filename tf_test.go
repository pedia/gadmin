package gadmin

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

type testCase struct {
	template       string
	argument       interface{}
	expectedResult string
	hasError       bool
}

var testCases = []testCase{
	{
		template: `
		{{define "foo"}}
		{{if $args := . | require "arg1" | require "arg2" "int" | args }}
		{{with .Origin }}
			{{.Bar}}
			{{$args.arg1}}
		{{ end }}
		{{ end }}
		{{ end }}

		{{template "foo" . | arg "arg1" "test" | arg "arg2" 42}}
		`,
		argument: map[string]string{
			"Bar": "bar",
		},
		expectedResult: "bartest",
		hasError:       false,
	},
	{
		template: `
		{{define "foo"}}
		{{if $args := . | require "arg1" | require "arg2" "string" | args }}
		{{with .Origin }}
			{{.Bar}}
			{{$args.arg1}}
		{{ end }}
		{{ end }}
		{{ end }}

		{{template "foo" . | arg "arg1" "test" | arg "arg2" 42}}
		`,
		argument: map[string]string{
			"Bar": "bar",
		},
		expectedResult: "bartest",
		hasError:       true,
	},
	{
		template: `
		{{ $x := slice 1 2 3 }}
		{{ range $y := $x }}
		{{$y}},
		{{end}}
		`,
		argument:       map[string]string{},
		expectedResult: "1,2,3,",
		hasError:       false,
	},
	{
		template: `
		{{ $m := map "foo" 1 | map "bar" 2 }}
		{{ range $k, $v := $m }}
		{{$k}}:{{$v}},
		{{end}}`,
		argument:       map[string]string{},
		expectedResult: "bar:2,foo:1,",
		hasError:       false,
	},
	{
		template: `
		{{ $r1 := rng 5 }}
		{{ range $e := $r1 }}
		{{$e}},
		{{end}}`,
		argument:       map[string]string{},
		expectedResult: "0,1,2,3,4,",
		hasError:       false,
	},
	{
		template: `
		{{ $r2 := rng 1 4 }}
		{{ range $e := $r2 }}
		{{$e}},
		{{end}}`,
		argument:       map[string]string{},
		expectedResult: "1,2,3,",
		hasError:       false,
	},
	{
		template: `
		{{ $r3 := rng 10 1 -3}}
		{{ range $e := $r3 }}
		{{$e}},
		{{end}}`,
		argument:       map[string]string{},
		expectedResult: "10,7,4,",
		hasError:       false,
	},
	{
		template: `
		{{ $r4 := rng 3 }}
		{{ $r4_1 := $r4 | append 3 4 }}
		{{ range $e := $r4_1 }}
		{{$e}},
		{{end}}`,
		argument:       map[string]string{},
		expectedResult: "0,1,2,3,4,",
		hasError:       false,
	},
	{
		template: `
		{{ . | split "," | join ";" }}`,
		argument:       "1,2,3",
		expectedResult: "1;2;3",
		hasError:       false,
	},
	{
		template: `
		{{ . | only "a" "b" }}`,
		argument:       map[string]any{"a": 1, "b": 2, "c": 3},
		expectedResult: "map[a:1 b:2]",
		hasError:       false,
	},
}

func removeWhite(s string) string {
	toRemove := []string{
		"\n", " ", "\t",
	}
	for _, r := range toRemove {
		s = strings.Replace(s, r, "", -1)
	}
	return s
}

func runTestCase(t *testing.T, tc testCase) {
	temp := template.Must(
		template.New("test_template").Funcs(
			FuncsText,
		).Parse(tc.template))
	buf := new(bytes.Buffer)
	err := temp.Execute(buf, tc.argument)
	if (err != nil) != tc.hasError {
		t.Errorf("haserror status unexpected. Expected %v, actual error %v", tc.hasError, err)
	}
	if !tc.hasError && removeWhite(buf.String()) != removeWhite(tc.expectedResult) {
		t.Errorf("Unexpected result. Expected: %s, Actual: %s", tc.expectedResult, buf.String())
	}
}

func TestAll(t *testing.T) {
	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
