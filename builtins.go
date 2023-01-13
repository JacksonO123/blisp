package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func InitBuiltins(ds *dataStore) {
	ds.builtins["print"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		Print(ds, params...)
		return true, "\"(printing " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
	}
	ds.builtins["+"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, fmt.Sprint(Add(ds, params...))
	}
	ds.builtins["-"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, fmt.Sprint(Sub(ds, params...))
	}
	ds.builtins["*"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, fmt.Sprint(Mult(ds, params...))
	}
	ds.builtins["/"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, fmt.Sprint(Divide(ds, params...))
	}
	ds.builtins["^"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"^\". Expected 2 found ", len(params))
		}
		return true, fmt.Sprint(Exp(ds, params[0], params[1]))
	}
	ds.builtins["%"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"%\". Expected 2 found ", len(params))
		}
		return true, fmt.Sprint(Mod(ds, params[0], params[1]))
	}
	ds.builtins["eval"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 1 {
			hasReturn, toReturn := Eval(ds, params[0][1:len(params[0])-1], scopes, false)
			if !hasReturn {
				return true, "\"(evaluating " + QuoteToQuoteLiteral(params[0]) + ")\""
			}
			return true, toReturn
		} else {
			for _, v := range params {
				if len(v) > 0 {
					Eval(ds, v[1:len(v)-1], scopes, false)
				}
			}
			return true, "\"(evaluating " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
		}
	}
	ds.builtins["var"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"var\". Expected 2 found ", len(params))
		}
		MakeVar(ds, scopes, params[0], params[1])
		return true, "\"(initializing " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
	}
	ds.builtins["set"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 2 {
			SetVar(ds, params[0], params[1])
			return true, "\"(setting " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
		} else if len(params) == 3 {
			SetIndex(ds, params[0], params[1], params[2])
			return true, "\"(setting " + QuoteToQuoteLiteral(params[0]) + " at index " + params[1] + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
		} else {
			log.Fatal("Invalid number of parameters to \"set\". Expected 2 or 3 found ", len(params))
		}
		return false, ""
	}
	ds.builtins["free"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"free\". Expected 1 found ", len(params))
		}
		FreeVar(ds, params[0])
		return true, "\"(freeing " + QuoteToQuoteLiteral(params[0]) + ")\""
	}
	ds.builtins["type"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"type\". Expected 1 found ", len(params))
		}
		return true, "\"" + GetValueType(ds, params[0]) + "\""
	}
	ds.builtins["get"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"get\". Expected 2 found ", len(params))
		}
		return true, GetValueFromList(ds, params[0], params[1])
	}
	ds.builtins["loop"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 3 {
			ds.inLoop = true
			valType := GetValueType(ds, params[0])
			if valType == "List" {
				LoopListIterator(ds, scopes, params[0], params[1], params[2])
				ds.inLoop = false
				return true, "\"(looping over " + params[0] + ")\""
			} else if valType == "Int" {
				LoopTo(ds, scopes, params[0], params[1], params[2])
				ds.inLoop = false
				return true, "\"(looping to " + params[0] + ")\""
			} else {
				log.Fatal("Expecting first param to be \"List\" or \"Int\", got:", valType)
			}
		} else if len(params) == 4 {
			ds.inLoop = true
			valType := GetValueType(ds, params[0])
			if valType == "List" {
				LoopListIndexIterator(ds, scopes, params[0], params[1], params[2], params[3])
				ds.inLoop = false
				return true, "\"(looping over " + params[0] + ")\""
			} else if valType == "Int" {
				LoopFromTo(ds, scopes, params[0], params[1], params[2], params[3])
				ds.inLoop = false
				return true, "\"(looping from " + params[0] + " to " + params[1] + ")\""
			} else {
				log.Fatal("Expecting first param to be list, got:", valType)
			}
		}
		return false, ""
	}
	ds.builtins["scan-line"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 0 {
			line := ""
			fmt.Scanln(&line)
			return true, line
		} else if len(params) == 1 {
			line := ""
			fmt.Scanln(&line)
			if _, ok := ds.vars[params[0]]; ok {
				SetVar(ds, params[0], line)
				return true, "\"(setting " + params[0] + " to " + line + ")\""
			} else {
				log.Fatal("Unable to assign value to", params[0])
			}
		} else {
			log.Fatal("Invalid number of parameters to \"scan-line\". Expected 0 or 2 found ", len(params))
		}
		return false, ""
	}
	ds.builtins["if"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 2 || len(params) == 3 {
			return If(ds, scopes, params...)
		} else {
			log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
		}
		return false, ""
	}
	ds.builtins["eq"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) > 0 {
			return true, fmt.Sprint(Eq(ds, params...))
		} else {
			log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found ", len(params))
		}
		return false, ""
	}
	ds.builtins["body"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return Eval(ds, flatBlock[6:len(flatBlock)-1], scopes, false)
	}
	ds.builtins["append"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) < 2 {
			log.Fatal("Invalid number of parameters to \"append\". Expected 2 or more found ", len(params))
		}
		res := ListFunc(ds, AppendToList, params...)
		if _, ok := ds.vars[params[0]]; ok {
			SetVar(ds, params[0], res)
			return true, "\"(appending [" + QuoteToQuoteLiteral(strings.Join(params[1:], ",")) + "] to " + params[0] + ")\""
		} else {
			return true, res
		}
	}
	ds.builtins["prepend"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) < 2 {
			log.Fatal("Invalid number of parameters to \"prepend\". Expected 2 or more found ", len(params))
		}
		res := ListFunc(ds, PrependToList, params...)
		if _, ok := ds.vars[params[0]]; ok {
			SetVar(ds, params[0], res)
			return true, "\"(prepending [" + QuoteToQuoteLiteral(strings.Join(params[1:], " ")) + "] to " + params[0] + ")\""
		} else {
			return true, res
		}
	}
	ds.builtins["concat"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, "\"" + Concat(ds, params...) + "\""
	}
	ds.builtins["exit"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 0 {
			log.Fatal("Invalid number of parameters to \"exit\". Expected 0 found ", len(params))
		}
		os.Exit(0)
		return false, ""
	}
	ds.builtins["break"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		return true, "(break)"
	}
	ds.builtins["pop"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"pop\". Expected 1 found ", len(params))
		}
		return true, Pop(ds, params[0])
	}
	ds.builtins["remove"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"remove\". Expected 2 found ", len(params))
		}
		return true, Remove(ds, params[0], params[1])
	}
	ds.builtins["len"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"len\". Expected 1 found ", len(params))
		}
		return true, fmt.Sprint(Len(ds, params[0]))
	}
	ds.builtins["and"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 0 {
			log.Fatal("Invalid number of parameters to \"and\". Expected 1 or more found ", len(params))
		}
		return true, fmt.Sprint(And(ds, params...))
	}
	ds.builtins["or"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) == 0 {
			log.Fatal("Invalid number of parameters to \"or\". Expected 1 or more found ", len(params))
		}
		return true, fmt.Sprint(Or(ds, params...))
	}
	ds.builtins["not"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"not\". Expected 1 or more found ", len(params))
		}
		return true, fmt.Sprint(Not(ds, params[0]))
	}
	ds.builtins["func"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if len(params) < 3 {
			log.Fatal("Invalid number of parameters to \"func\". Expected 3 or more found ", len(params))
		}
		MakeFunction(ds, scopes, params...)
		return true, "\"(setting function " + params[0] + " with " + strings.Join(params[1:], " ") + ")\""
	}
	ds.builtins["return"] = func(ds *dataStore, scopes int, flatBlock string, params []string) (bool, string) {
		if !ds.inFunc {
			log.Fatal("Not in func, cannot return")
		}
		vals := []string{}
		for _, v := range params {
			vals = append(vals, GetValue(ds, v).value)
		}
		return true, strings.Join(vals, " ")
	}
}
