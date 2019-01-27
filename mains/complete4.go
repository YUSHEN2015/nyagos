package mains

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/gopher-lua"

	"github.com/zetamatta/nyagos/completion"
)

func complete4getter(L Lua) int {
	key, ok := L.Get(1).(lua.LString)
	if !ok {
		return lerror(L, "nyagos.complete_for[] too few arguments")
	}
	if _, ok := completion.CustomCompletion[string(key)]; ok {
		L.Push(lua.LTrue)
	} else {
		L.Push(lua.LNil)
	}
	return 1
}

func complete4setter(L Lua) int {
	key, ok := L.Get(-2).(lua.LString)
	if !ok {
		return lerror(L, "nyagos.complete_for[] too few arguments")
	}
	val := L.Get(-1)
	if val == lua.LNil {
		delete(completion.CustomCompletion, string(key))
		return 0
	}
	f, ok := val.(*lua.LFunction)
	if !ok {
		return lerror(L, "nyagos.complete_for[]= not function")
	}
	completion.CustomCompletion[string(key)] = func(ctx context.Context, args []string) ([]completion.Element, error) {
		LL, ok := ctx.Value(luaKey).(Lua)
		if !ok {
			return nil, errors.New("completion.CustomCompletion: no lua instance")
		}
		tbl := LL.NewTable()
		for i, arg1 := range args {
			LL.SetTable(tbl, lua.LNumber(i+1), lua.LString(arg1))
		}

		defer setContext(LL, getContext(LL))
		setContext(LL, ctx)

		LL.Push(f)
		LL.Push(tbl)
		if err := LL.PCall(1, 1, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		result := LL.Get(-1)
		if rtbl, ok := result.(*lua.LTable); ok {
			r := make([]completion.Element, 0, rtbl.Len())
			rtbl.ForEach(func(_ lua.LValue, val lua.LValue) {
				if s, ok := val.(lua.LString); ok {
					r = append(r, completion.Element1(string(s)))
				}
			})
			return r, nil
		} else if s, ok := result.(lua.LString); ok {
			list := strings.Split(string(s), "\n")
			r := make([]completion.Element, 0, len(list))
			for _, r1 := range list {
				r = append(r, completion.Element1(string(r1)))
			}
			return r, nil
		} else {
			return nil, errors.New("not a table or string")
		}
	}
	return 1
}
