package main

import (
	"fmt"
	"log"
	"os"
)

func InitBuiltins(ds *dataStore) {
	ds.builtins["print"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		Print(ds, params...)
		return true, []dataType{{dataType: String, value: "\"(printing)\""}}
	}
	ds.builtins["+"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{Add(ds, params...)}
	}
	ds.builtins["-"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{Sub(ds, params...)}
	}
	ds.builtins["*"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{Mult(ds, params...)}
	}
	ds.builtins["/"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{Divide(ds, params...)}
	}
	ds.builtins["^"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"^\". Expected 2 found ", len(params))
		}
		return true, []dataType{Exp(ds, params[0], params[1])}
	}
	ds.builtins["%"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"%\". Expected 2 found ", len(params))
		}
		return true, []dataType{Mod(ds, params[0], params[1])}
	}
	ds.builtins["eval"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		strVal := params[0].value.(string)
		if len(params) == 1 {
			toEval := PrepQuotesString(strVal[1 : len(strVal)-1])
			hasReturn, toReturn := Eval(ds, Tokenize(toEval), scopes, false)
			if !hasReturn {
				return true, []dataType{{dataType: String, value: "\"(evaluating " + QuoteToQuoteLiteral(strVal) + ")\""}}
			}
			return true, toReturn
		} else {
			for _, v := range params {
				strVal := v.value.(string)
				Eval(ds, Tokenize(strVal[1:len(strVal)-1]), scopes, false)
			}
			toPrint := ""
			for i, v := range params {
				toPrint += GetStrValue(v)
				if i < len(params)-1 {
					toPrint += ", "
				}
			}
			return true, []dataType{{dataType: String, value: "\"(evaluating " + QuoteToQuoteLiteral(toPrint) + ")\""}}
		}
	}
	ds.builtins["var"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		MakeVar(ds, scopes, params[0].value.(string), params[1])
		return true, []dataType{{dataType: String, value: "\"(initializing " + QuoteToQuoteLiteral(params[0].value.(string)) + " to " + GetStrValue(params[1]) + ")\""}}
	}
	ds.builtins["set"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if _, ok := ds.vars[params[0].value.(string)]; !ok {
			log.Fatal("Cannot set variable: ", params[0].value, ", variable is not initialized")
		}
		if len(params[2:]) == 0 {
			SetVar(ds, params[0].value.(string), params[1])
			return true, []dataType{{dataType: String, value: "\"(setting " + QuoteToQuoteLiteral(params[0].value.(string)) + " to " + GetStrValue(params[1]) + ")\""}}
		} else {
			var index int
			if params[1].dataType == Int {
				index = params[1].value.(int)
			} else {
				log.Fatal("Cannot set index of array with type \"Float\"")
			}
			SetIndex(ds, params[0], index, params[2])
			return true, []dataType{{dataType: String, value: "\"(setting " + QuoteToQuoteLiteral(params[0].value.(string)) + " at index " + GetStrValue(params[1]) + " to " + QuoteToQuoteLiteral(GetStrValue(params[1])) + ")\""}}
		}
	}
	ds.builtins["free"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"free\". Expected 1 found ", len(params))
		}
		FreeVar(ds, params[0].value.(string))
		return true, []dataType{{dataType: String, value: "\"(freeing " + QuoteToQuoteLiteral(params[0].value.(string)) + ")\""}}
	}
	ds.builtins["type"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"type\". Expected 1 found ", len(params))
		}
		return true, []dataType{{dataType: String, value: dataTypes[params[0].dataType]}}
	}
	ds.builtins["get"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"get\". Expected 2 found ", len(params))
		}
		return true, []dataType{GetValueFromList(ds, params[0], params[1])}
	}
	ds.builtins["loop"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 3 {
			ds.inLoop = true
			val := params[0]
			if val.dataType == Ident {
				val = GetDsValue(ds, val)
			}
			if val.dataType == List {
				LoopListIterator(ds, scopes, val, params[1], params[2])
				ds.inLoop = false
				return true, []dataType{{dataType: String, value: "\"(looping over " + GetStrValue(val) + ")\""}}
			} else if val.dataType == Int {
				LoopTo(ds, scopes, val, params[1], params[2])
				ds.inLoop = false
				return true, []dataType{{dataType: String, value: "\"(looping over " + GetStrValue(val) + ")\""}}
			} else {
				log.Fatal("Expecting first param to be \"List\" or \"Int\", got:", dataTypes[val.dataType])
			}
		} else if len(params) == 4 {
			ds.inLoop = true
			val := params[0]
			if val.dataType == Ident {
				val = GetDsValue(ds, val)
			}
			if val.dataType == List {
				LoopListIndexIterator(ds, scopes, val, params[1], params[2], params[3])
				ds.inLoop = false
				return true, []dataType{{dataType: String, value: "\"(looping over " + GetStrValue(val) + ")\""}}
			} else if val.dataType == Int {
				LoopFromTo(ds, scopes, val, params[1], params[2], params[3])
				ds.inLoop = false
				return true, []dataType{{dataType: String, value: "\"(looping from " + GetStrValue(val) + " to " + GetStrValue(params[1]) + ")\""}}
			} else {
				log.Fatal("Expecting first param to be list, got: ", dataTypes[val.dataType])
			}
		}
		return false, []dataType{}
	}
	ds.builtins["scan-line"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 0 {
			line := ""
			fmt.Scanln(&line)
			return true, []dataType{{dataType: String, value: "\"" + line + "\""}}
		} else if len(params) == 1 {
			line := ""
			fmt.Scanln(&line)
			if params[0].dataType == Ident {
				SetVar(ds, params[0].value.(string), dataType{dataType: String, value: "\"" + line + "\""})
				return true, []dataType{{dataType: String, value: "\"(setting " + GetStrValue(params[0]) + " to " + line + ")\""}}
			} else {
				log.Fatal("Unable to assign value to", params[0])
			}
		} else {
			log.Fatal("Invalid number of parameters to \"scan-line\". Expected 0 or 2 found ", len(params))
		}
		return false, []dataType{}
	}
	ds.builtins["if"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 2 || len(params) == 3 {
			v1, v2 := If(ds, scopes, params...)
			return v1, v2
		} else {
			log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
		}
		return false, []dataType{}
	}
	ds.builtins["body"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return Eval(ds, params[0].value.([]token), scopes, false)
	}
	ds.builtins["eq"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) > 0 {
			return true, []dataType{{dataType: Bool, value: Eq(ds, params...)}}
		} else {
			log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found ", len(params))
		}
		return false, []dataType{}
	}
	ds.builtins["append"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) < 2 {
			log.Fatal("Invalid number of parameters to \"append\". Expected 2 or more found ", len(params))
		}
		res := ListFunc(ds, AppendToList, params...)
		return true, []dataType{res}
	}
	ds.builtins["prepend"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) < 2 {
			log.Fatal("Invalid number of parameters to \"prepend\". Expected 2 or more found ", len(params))
		}
		res := ListFunc(ds, PrependToList, params...)
		return true, []dataType{res}
	}
	ds.builtins["concat"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{{dataType: String, value: "\"" + Concat(ds, params...) + "\""}}
	}
	ds.builtins["exit"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 0 {
			os.Exit(0)
		} else if len(params) == 1 {
			os.Exit(params[0].value.(int))
		} else {
			log.Fatal("Invalid number of parameters to \"exit\". Expected 0, 1, or more found ", len(params))
		}
		return false, []dataType{}
	}
	ds.builtins["break"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		return true, []dataType{{dataType: BreakVals, value: params}}
	}
	ds.builtins["pop"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"pop\". Expected 1 found ", len(params))
		}
		return true, []dataType{Pop(ds, params[0])}
	}
	ds.builtins["remove"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 2 {
			log.Fatal("Invalid number of parameters to \"remove\". Expected 2 found ", len(params))
		}
		return true, []dataType{Remove(ds, params[0], params[1])}
	}
	ds.builtins["len"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"len\". Expected 1 found ", len(params))
		}
		return true, []dataType{{dataType: Int, value: Len(ds, params[0])}}
	}
	ds.builtins["and"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 0 {
			log.Fatal("Invalid number of parameters to \"and\". Expected 1 or more found ", len(params))
		}
		return true, []dataType{{dataType: Bool, value: And(ds, params...)}}
	}
	ds.builtins["or"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) == 0 {
			log.Fatal("Invalid number of parameters to \"or\". Expected 1 or more found ", len(params))
		}
		return true, []dataType{{dataType: Bool, value: Or(ds, params...)}}
	}
	ds.builtins["not"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) != 1 {
			log.Fatal("Invalid number of parameters to \"not\". Expected 1 or more found ", len(params))
		}
		return true, []dataType{{dataType: Bool, value: Not(ds, params[0])}}
	}
	ds.builtins["func"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		if len(params) < 2 {
			log.Fatal("Invalid number of parameters to \"func\". Expected 3 or more found ", len(params))
		}
		MakeFunction(ds, scopes, params[0], params[1:])
		return true, []dataType{{dataType: String, value: "\"(setting function " + params[0].value.(string) + " with " + GetStrValue(dataType{dataType: List, value: params[1:]}) + ")\""}}
	}
	// ds.builtins["return"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if !ds.inFunc {
	// 		log.Fatal("Not in func, cannot return")
	// 	}
	// 	vals := [][]token{}
	// 	// for _, v := range params {
	// 	// 	vals = append(vals, GetValue(ds, v).value...)
	// 	// }
	// 	return true, vals
	// }
}
