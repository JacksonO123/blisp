package main

import (
	"log"
)

func InitBuiltins(ds *dataStore) {
	ds.builtins["print"] = func(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
		Print(ds, params...)
		toPrint := ""
		for i, v := range params {
			toPrint += GetStrValue(v)
			if i < len(params)-1 {
				toPrint += ", "
			}
		}
		return true, []dataType{{dataType: String, value: "\"(printing " + toPrint + ")\""}}
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
			if params[1].dataType == Float {
				index = int(params[1].value.(float64))
			} else {
				index = params[1].value.(int)
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
	// ds.builtins["loop"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	// if len(params) == 3 {
	// 	// 	ds.inLoop = true
	// 	// 	valType := GetValueType(ds, params[0])
	// 	// 	if valType == "List" {
	// 	// 		LoopListIterator(ds, scopes, params[0], params[1], Tokenize(params[2].value))
	// 	// 		ds.inLoop = false
	// 	// 		return true, [][]token{GetToken("\"(looping over " + params[0].value + ")\"")}
	// 	// 	} else if valType == "Int" {
	// 	// 		LoopTo(ds, scopes, params[0], params[1], Tokenize(params[2].value))
	// 	// 		ds.inLoop = false
	// 	// 		return true, [][]token{GetToken("\"(looping to " + params[0].value + ")\"")}
	// 	// 	} else {
	// 	// 		log.Fatal("Expecting first param to be \"List\" or \"Int\", got:", valType)
	// 	// 	}
	// 	// } else if len(params) == 4 {
	// 	// 	ds.inLoop = true
	// 	// 	valType := GetValueType(ds, params[0])
	// 	// 	if valType == "List" {
	// 	// 		LoopListIndexIterator(ds, scopes, params[0], params[1], params[2], Tokenize(params[3].value))
	// 	// 		ds.inLoop = false
	// 	// 		return true, [][]token{GetToken("\"(looping over " + params[0].value + ")\"")}
	// 	// 	} else if valType == "Int" {
	// 	// 		LoopFromTo(ds, scopes, params[0], params[1], params[2], Tokenize(params[3].value))
	// 	// 		ds.inLoop = false
	// 	// 		return true, [][]token{GetToken("\"(looping from " + params[0].value + " to " + params[1].value + ")\"")}
	// 	// 	} else {
	// 	// 		log.Fatal("Expecting first param to be list, got: ", valType)
	// 	// 	}
	// 	// }
	// 	return false, []token{}
	// }
	// ds.builtins["scan-line"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) == 0 {
	// 		line := ""
	// 		fmt.Scanln(&line)
	// 		return true, []token{GetToken("\"" + line + "\"")}
	// 	} else if len(params) == 1 {
	// 		line := ""
	// 		fmt.Scanln(&line)
	// 		if _, ok := ds.vars[params[0].value]; ok {
	// 			// SetVar(ds, params[0], GetToken("\""+line+"\""))
	// 			return true, []token{GetToken("\"(setting " + params[0].value + " to " + line + ")\"")}
	// 		} else {
	// 			log.Fatal("Unable to assign value to", params[0])
	// 		}
	// 	} else {
	// 		log.Fatal("Invalid number of parameters to \"scan-line\". Expected 0 or 2 found ", len(params))
	// 	}
	// 	return false, []token{}
	// }
	// ds.builtins["if"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) == 2 || len(params) == 3 {
	// 		return If(ds, scopes, params...)
	// 	} else {
	// 		log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
	// 	}
	// 	return false, [][]token{}
	// }
	// ds.builtins["eq"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) > 0 {
	// 		return true, [][]token{GetToken(fmt.Sprint(Eq(ds, params...)))}
	// 	} else {
	// 		log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found ", len(params))
	// 	}
	// 	return false, [][]token{}
	// }
	// ds.builtins["append"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) < 2 {
	// 		log.Fatal("Invalid number of parameters to \"append\". Expected 2 or more found ", len(params))
	// 	}
	// 	res := ListFunc(ds, AppendToList, params...)
	// 	if _, ok := ds.vars[params[0].value]; ok {
	// 		// SetVar(ds, params[0], res)
	// 		return true, [][]token{GetToken("\"(appending [" + QuoteToQuoteLiteral(strings.Join(TokensToValue(params[1:]), ",")) + "] to " + params[0].value + ")\"")}
	// 	} else {
	// 		return true, [][]token{res}
	// 	}
	// }
	// ds.builtins["prepend"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) < 2 {
	// 		log.Fatal("Invalid number of parameters to \"prepend\". Expected 2 or more found ", len(params))
	// 	}
	// 	res := ListFunc(ds, PrependToList, params...)
	// 	if _, ok := ds.vars[params[0].value]; ok {
	// 		// SetVar(ds, params[0], res)
	// 		return true, [][]token{GetToken("\"(prepending [" + QuoteToQuoteLiteral(strings.Join(TokensToValue(params[1:]), " ")) + "] to " + params[0].value + ")\"")}
	// 	} else {
	// 		return true, [][]token{res}
	// 	}
	// }
	// ds.builtins["concat"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	return true, [][]token{GetToken("\"" + Concat(ds, params...) + "\"")}
	// }
	// ds.builtins["exit"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) != 0 {
	// 		log.Fatal("Invalid number of parameters to \"exit\". Expected 0 found ", len(params))
	// 	}
	// 	os.Exit(0)
	// 	return false, [][]token{}
	// }
	// ds.builtins["break"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	return true, [][]token{GetToken("break")}
	// }
	// ds.builtins["pop"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) != 1 {
	// 		log.Fatal("Invalid number of parameters to \"pop\". Expected 1 found ", len(params))
	// 	}
	// 	return true, [][]token{Pop(ds, params[0])}
	// }
	// ds.builtins["remove"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) != 2 {
	// 		log.Fatal("Invalid number of parameters to \"remove\". Expected 2 found ", len(params))
	// 	}
	// 	return true, [][]token{Remove(ds, params[0], params[1])}
	// }
	// ds.builtins["len"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) != 1 {
	// 		log.Fatal("Invalid number of parameters to \"len\". Expected 1 found ", len(params))
	// 	}
	// 	return true, [][]token{GetToken(fmt.Sprint(Len(ds, params[0])))}
	// }
	// ds.builtins["and"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) == 0 {
	// 		log.Fatal("Invalid number of parameters to \"and\". Expected 1 or more found ", len(params))
	// 	}
	// 	return true, [][]token{GetToken(fmt.Sprint(And(ds, params...)))}
	// }
	// ds.builtins["or"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) == 0 {
	// 		log.Fatal("Invalid number of parameters to \"or\". Expected 1 or more found ", len(params))
	// 	}
	// 	return true, [][]token{GetToken(fmt.Sprint(Or(ds, params...)))}
	// }
	// ds.builtins["not"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) != 1 {
	// 		log.Fatal("Invalid number of parameters to \"not\". Expected 1 or more found ", len(params))
	// 	}
	// 	return true, [][]token{GetToken(fmt.Sprint(Not(ds, params[0])))}
	// }
	// ds.builtins["func"] = func(ds *dataStore, scopes int, params []token) (bool, []token) {
	// 	if len(params) < 3 {
	// 		log.Fatal("Invalid number of parameters to \"func\". Expected 3 or more found ", len(params))
	// 	}
	// 	MakeFunction(ds, scopes, params...)
	// 	return true, [][]token{GetToken("\"(setting function " + params[0].value + " with " + strings.Join(TokensToValue(params[1:]), " ") + ")\"")}
	// }
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
