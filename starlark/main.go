package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"go.starlark.net/starlark"
)

func toStarlark(data interface{}) starlark.Value {
	t := reflect.TypeOf(data)
	fmt.Println(t.Kind())
	switch data.(type) {
	case string:
		return starlark.String(data.(string))
	case bool:
		if data.(bool) == true {
			return starlark.True
		} else {
			return starlark.False
		}
	case float64:
		return starlark.Float(data.(float64))
	case []interface{}:
		holder := []starlark.Value{}
		for _, v := range data.([]interface{}) {
			holder = append(holder, toStarlark(v))
		}
		return starlark.NewList(holder)

	case map[string]interface{}:
		holder := starlark.NewDict(len(data.(map[string]interface{})))

		for key, v := range data.(map[string]interface{}) {
			holder.SetKey(starlark.String(key), toStarlark(v))
		}
		return holder
	case []map[string]interface{}:
		holder := []starlark.Value{}
		for _, v := range data.([]map[string]interface{}) {
			holder = append(holder, toStarlark(v))
		}
		return starlark.NewList(holder)
	default:
		// @todo custom structs
		fmt.Printf("%v is unknown \n ", data)
		panic("Unprocssed type")
	}
}

func getReducerDefault(v *starlark.Function) starlark.Value {
	fmt.Printf("type=%s string=%s\n", v.Type(), v.String())

	fmt.Println(v.Param(0))
	x := v.ParamDefault(0)
	fmt.Println("x", reflect.TypeOf(x))
	switch x.(type) {
	case *starlark.Dict:
		return starlark.NewDict(0)
	default:
		return starlark.MakeInt(0)
	}
}

func main() {
	const data = `
print("HEY", results)
y = 22
print(myrepeat("A", y))

def cats():
	return "cats"

def map(result):
	pass

def reduce(aggr={}, result=None):
	aggr[result] = 1
	return aggr

def filter(result):
	pass

def test(results):
	for x in results:
		for key in x:
			print("key", key)

test(results)

initial = {}
`

	// The Thread defines the behavior of the built-in 'print' function.
	thread := &starlark.Thread{
		Name:  "example",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
	}

	tmp := []map[string]interface{}{
		{
			"test": "Value",
			"key":  "value",
		},
	}
	yy := toStarlark(tmp)

	predeclared := starlark.StringDict{
		"results": yy,
		"myrepeat": starlark.NewBuiltin("myrepeat", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var s string
			var n int = 1
			if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s", &s, "n?", &n); err != nil {
				return nil, err
			}
			return starlark.String(strings.Repeat(s, n)), nil
		}),
	}

	// Execute a program.
	globals, err := starlark.ExecFile(thread, "apparent/filename.star", data, predeclared)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	fmt.Println("INITIAL", globals["initial"])
	// Print the global environment.
	fmt.Println("\nGlobals:")
	aggr := starlark.NewDict(0)
	_, err = starlark.Call(thread, globals["reduce"], starlark.Tuple{aggr, starlark.MakeInt(10)}, nil)

	iter := aggr.Iterate()
	defer iter.Done()
	var key starlark.Value
	for iter.Next(&key) {
		val, _, _ := aggr.Get(key)
		fmt.Println("Iterator", key, val)
	}

	for _, name := range globals.Keys() {
		v := globals[name]
		fmt.Printf("%s (%s) = %s\n", name, v.Type(), v.String())

		switch v.(type) {
		case *starlark.Function:
			getReducerDefault(v.(*starlark.Function))

		default:
			fmt.Print("x")
		}
	}

}
