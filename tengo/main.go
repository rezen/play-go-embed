package main

// Missing script defined function that golang can execute
// https://github.com/d5/tengo/issues/420
import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/d5/tengo"
	"github.com/d5/tengo/stdlib"
)

func toTengo(data interface{}) tengo.Object {
	t := reflect.TypeOf(data)
	fmt.Println(t.Kind())

	switch data.(type) {
	case string:
		return &tengo.String{
			Value: data.(string),
		}
	case bool:
		if data.(bool) == true {
			return tengo.TrueValue
		} else {
			return tengo.FalseValue
		}
	case float64:
		return &tengo.Float{
			Value: data.(float64),
		}
	case []interface{}:
		holder := []tengo.Object{}
		for _, v := range data.([]interface{}) {
			holder = append(holder, toTengo(v))
		}
		return &tengo.Array{
			Value: holder,
		}

	case map[string]interface{}:
		holder := map[string]tengo.Object{}
		for key, v := range data.(map[string]interface{}) {
			holder[key] = toTengo(v)
		}
		return &tengo.Map{
			Value: holder,
		}
	case []map[string]interface{}:
		holder := []tengo.Object{}
		for _, v := range data.([]map[string]interface{}) {
			holder = append(holder, toTengo(v))
		}
		return &tengo.Array{
			Value: holder,
		}
	default:
		// @todo custom structs
		fmt.Printf("%v is unknown \n ", data)
		panic("Unprocssed type")
	}
}

func main() {
	// create a new Script instance
	script := tengo.NewScript([]byte(
		`
fmt := import("fmt")
enum := import("enum")

fmt.println(results)
enum.each(results,  func(k, x) {
	fmt.println(x)
})

x := myrepeat("X", "20")
fmt.println("printing this out ", x)

reduce := func(aggr, result) {
	fmt.println("HEY")
	return "test"
}

`))
	tmp := []map[string]interface{}{
		{
			"test": "Value",
			"key":  "value",
		},
	}

	results := toTengo(tmp)

	script.SetImports(stdlib.GetModuleMap("fmt", "enum"))

	err := script.Add("results", results)
	fmt.Println(err)
	err = script.Add("myrepeat", &tengo.UserFunction{
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 2 {
				return nil, tengo.ErrWrongNumArguments
			}
			str, ok := tengo.ToString(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name: "str",
				}
			}
			num, ok := tengo.ToInt(args[1])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name: "int",
				}
			}
			return &tengo.String{
				Value: strings.Repeat(str, num),
			}, nil
		}})

	fmt.Println(err)

	// run the script
	_, err = script.RunContext(context.Background())
	if err != nil {
		panic(err)
	}
}
