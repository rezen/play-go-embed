package main

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func toLValue(data interface{}) lua.LValue {
	t := reflect.TypeOf(data)
	fmt.Println(t.Kind())
	switch data.(type) {
	case string:
		return lua.LString(data.(string))
	case bool:
		if data.(bool) == true {
			return lua.LTrue
		} else {
			return lua.LFalse
		}
	case float64:
		return lua.LNumber(data.(float64))
	case []interface{}:
		holder := &lua.LTable{}
		for i, v := range data.([]interface{}) {
			holder.RawSetInt(i, toLValue(v))
		}
		return holder

	case map[string]interface{}:
		holder := &lua.LTable{}

		for key, v := range data.(map[string]interface{}) {
			holder.RawSetString(key, toLValue(v))
		}
		return holder
	case []map[string]interface{}:
		holder := &lua.LTable{}
		for i, v := range data.([]interface{}) {
			holder.RawSetInt(i, toLValue(v))
		}
		return holder
	default:
		// @todo custom structs
		fmt.Printf("%v is unknown \n ", data)
		panic("Unprocssed type")
	}
}

func myrepeat(L *lua.LState) int {
	str := L.ToString(1)
	num := L.ToInt(2)
	L.Push(lua.LString(strings.Repeat(str, num)))
	return 1
}

func main() {
	L := lua.NewState()

	L.SetGlobal("myrepeat", L.NewFunction(myrepeat))
	defer L.Close()

	// Load your Lua script
	err := L.DoString(`
		initial = {}
		x = myrepeat("x",12)

		print(x)

		function reducer(state, row)
			state[row.id] = row
			return state
		end
	`)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ret := L.GetGlobal("initial")
	fmt.Println(ret.Type())
	ret.(*lua.LTable).ForEach(func(k lua.LValue, v lua.LValue) {
		fmt.Println(k)
	})

	// Get the type of initial
	fmt.Println(ret.Type())
	// Get the value of initial
	fmt.Println(ret.String())
	// Get table keys

	// Get size of table
	table1 := toLValue(map[string]interface{}{
		"id": "bar",
	})

	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("reducer"),
		NRet:    1,
		Protect: true,
	}, ret, table1)

	if err != nil {
		panic(err)
	}
	ret = L.Get(-1) // returned value

	fmt.Println("RETURN", ret)

	table2 := toLValue(map[string]interface{}{
		"id": "foo",
	})

	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("reducer"),
		NRet:    1,
		Protect: true,
	}, ret, table2)

	if err != nil {
		panic(err)
	}
	ret = L.Get(-1) // returned value

	fmt.Println("RETURN", ret)

	ret.(*lua.LTable).ForEach(func(k lua.LValue, v lua.LValue) {
		fmt.Println(k, v)
	})

}
