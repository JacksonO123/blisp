package main

import (
	"fmt"
	"log"
	"os"
)

func validateRangeParam(name string, rangeToCheck [2]int, numOfParams int) {
	if numOfParams < rangeToCheck[0] || numOfParams > rangeToCheck[1] {
		log.Fatal("Invalid number of parameters to \"", name, "\", expected range from ", rangeToCheck[0], " to ", rangeToCheck[1], " found ", numOfParams)
	}
}

func validateNumParam(name string, numToCheck int, numOfParams int) {
	if numToCheck != numOfParams {
		log.Fatal("Invalid number of parameters to \"", name, "\", expected ", numToCheck, " found ", numOfParams)
	}
}

func InitBuiltins(ds *dataStore) {
	ds.builtins = []builtin{
		{
			name: "print",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				Print(ds, params...)
				return nil
			},
		},
		{
			name: "+",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{Add(ds, params...)}
			},
		},
		{
			name: "-",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{Sub(ds, params...)}
			},
		},
		{
			name: "*",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{Mult(ds, params...)}
			},
		},
		{
			name: "/",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{Divide(ds, params...)}
			},
		},
		{
			name: "^",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"^\". Expected 2 found ", len(params))
				}
				return &[]dataType{Exp(ds, params[0], params[1])}
			},
		},
		{
			name: "%",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"%\". Expected 2 found ", len(params))
				}
				return &[]dataType{Mod(ds, params[0], params[1])}
			},
		},
		{
			name: "eval",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				strVal := params[0].value.(string)
				if len(params) == 1 {
					toEval := PrepQuotesString(strVal)
					return Eval(ds, Tokenize(toEval), scopes)
				} else {
					for _, v := range params {
						strVal := v.value.(string)
						Eval(ds, Tokenize(strVal), scopes)
					}
					return nil
				}
			},
		},
		{
			name: "var",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"var\". Expected 2 found ", len(params))
				}
				MakeVar(ds, scopes, params[0].value.(string), params[1], false)
				return nil
			},
		},
		{
			name: "const",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				MakeVar(ds, scopes, params[0].value.(string), params[1], true)
				return nil
			},
		},
		{
			name: "set",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if _, ok := ds.vars[params[0].value.(string)]; !ok {
					log.Fatal("Cannot set variable: ", params[0].value, ", variable is not initialized")
				}
				if len(params) == 2 {
					SetVar(ds, params[0].value.(string), params[1])
				} else if len(params) == 3 {
					SetValue(ds, params[0], params[1], params[2])
				} else {
					log.Fatal("Error in \"set\", invalid number of parameters: ", len(params))
				}
				return nil
			},
		},
		{
			name: "free",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"free\". Expected 1 found ", len(params))
				}
				FreeVar(ds, params[0].value.(string))
				return nil
			},
		},
		{
			name: "type",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"type\". Expected 1 found ", len(params))
				}
				return &[]dataType{{dataType: String, value: GetType(ds, params[0])}}
			},
		},
		{
			name: "get",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"get\". Expected 2 found ", len(params))
				}
				return &[]dataType{GetFromValue(ds, params[0], params[1])}
			},
		},
		{
			name: "loop",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 3 {
					val := params[0]
					if val.dataType == Ident {
						val = GetDsValue(ds, val)
					}
					if val.dataType == List {
						res := LoopListIterator(ds, scopes, val, params[1], params[2])
						ds.inLoop = false
						return res
					} else if val.dataType == Int {
						res := LoopTo(ds, scopes, val, params[1], params[2])
						ds.inLoop = false
						return res
					} else {
						log.Fatal("Error in \"Loop\". Expected first param to be \"List\" or \"Int\", found ", dataTypes[val.dataType])
					}
				} else if len(params) == 4 {
					ds.inLoop = true
					val := params[0]
					if val.dataType == Ident {
						val = GetDsValue(ds, val)
					}
					if val.dataType == List {
						res := LoopListIndexIterator(ds, scopes, val, params[1], params[2], params[3])
						ds.inLoop = false
						return res
					} else if val.dataType == Int {
						res := LoopFromTo(ds, scopes, val, params[1], params[2], params[3])
						ds.inLoop = false
						return res
					} else {
						log.Fatal("Error in \"Loop\". Expected first param to be list, got: ", dataTypes[val.dataType])
					}
				}
				return nil
			},
		},
		{
			name: "scan-line",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 0 {
					line := ""
					fmt.Scanln(&line)
					return &[]dataType{{dataType: String, value: line}}
				} else if len(params) == 1 {
					line := ""
					fmt.Scanln(&line)
					if params[0].dataType == Ident {
						SetVar(ds, params[0].value.(string), dataType{dataType: String, value: line})
						return nil
					} else {
						log.Fatal("Unable to assign value to", params[0])
					}
				} else {
					log.Fatal("Invalid number of parameters to \"scan-line\". Expected 0 or 2 found ", len(params))
				}
				return nil
			},
		},
		{
			name: "if",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 2 || len(params) == 3 {
					val := If(ds, scopes, params...)
					return val
				} else {
					log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
				}
				return nil
			},
		},
		{
			name: "body",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return Eval(ds, params[0].value.([]token), scopes)
			},
		},
		{
			name: "eq",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) > 0 {
					return &[]dataType{{dataType: Bool, value: Eq(ds, params...)}}
				} else {
					log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found ", len(params))
				}
				return nil
			},
		},
		{
			name: "append",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) < 2 {
					log.Fatal("Invalid number of parameters to \"append\". Expected 2 or more found ", len(params))
				}
				res := ListFunc(ds, AppendToList, params...)
				return &[]dataType{res}
			},
		},
		{
			name: "prepend",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) < 2 {
					log.Fatal("Invalid number of parameters to \"prepend\". Expected 2 or more found ", len(params))
				}
				res := ListFunc(ds, PrependToList, params...)
				return &[]dataType{res}
			},
		},
		{
			name: "concat",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{{dataType: String, value: Concat(ds, params...)}}
			},
		},
		{
			name: "exit",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 0 {
					os.Exit(0)
				} else if len(params) == 1 {
					os.Exit(params[0].value.(int))
				} else {
					log.Fatal("Invalid number of parameters to \"exit\". Expected 0, 1, or more found ", len(params))
				}
				return nil
			},
		},
		{
			name: "break",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{{dataType: BreakVal, value: nil}}
			},
		},
		{
			name: "pop",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"pop\". Expected 1 found ", len(params))
				}
				return &[]dataType{Pop(ds, params[0])}
			},
		},
		{
			name: "remove",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"remove\". Expected 2 found ", len(params))
				}
				return &[]dataType{Remove(ds, params[0], params[1])}
			},
		},
		{
			name: "len",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"len\". Expected 1 found ", len(params))
				}
				return &[]dataType{{dataType: Int, value: Len(ds, params[0])}}
			},
		},
		{
			name: "and",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 0 {
					log.Fatal("Invalid number of parameters to \"and\". Expected 1 or more found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: And(ds, params...)}}
			},
		},
		{
			name: "or",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 0 {
					log.Fatal("Invalid number of parameters to \"or\". Expected 1 or more found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: Or(ds, params...)}}
			},
		},
		{
			name: "not",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"not\". Expected 1 or more found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: Not(ds, params[0])}}
			},
		},
		{
			name: "func",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) < 2 {
					log.Fatal("Invalid number of parameters to \"func\". Expected 3 or more found ", len(params))
				}
				f := MakeFunction(ds, scopes, params[0], params[1:])
				// f can be nil, but only when returned is false
				if f == nil {
					return nil
				}
				return &[]dataType{*f}
			},
		},
		{
			name: "return",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if !ds.inFunc {
					log.Fatal("Not in func, cannot return")
				}
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"return\". Expected 1 found ", len(params))
				}
				val := GetDsValue(ds, params[0])
				return &[]dataType{{dataType: ReturnVal, value: val}}
			},
		},
		{
			name: "parse",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				var res dataType
				if len(params) == 1 {
					res = Parse(ds, params[0])
				} else {
					log.Fatal("Invalid number of parameters to \"parse\". Expected 1 found ", len(params))
				}
				return &[]dataType{res}
			},
		},
		{
			name: "<",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"<\". Expected 2 found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: LessThan(ds, params[0], params[1])}}
			},
		},
		{
			name: "<=",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"<=\". Expected 2 found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: LessThanOrEqualTo(ds, params[0], params[1])}}
			},
		},
		{
			name: ">",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \">\". Expected 2 found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: LessThan(ds, params[1], params[0])}}
			},
		},
		{
			name: ">=",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \">=\". Expected 2 found ", len(params))
				}
				return &[]dataType{{dataType: Bool, value: LessThanOrEqualTo(ds, params[1], params[0])}}
			},
		},
		{
			name: "read",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"read\". Expected 1 found ", len(params))
				}
				return &[]dataType{{dataType: String, value: GetFile(ds, params[0])}}
			},
		},
		{
			name: "write",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"write\". Expected 2 found ", len(params))
				}
				WriteFile(ds, params[0], params[1])
				return nil
			},
		},
		{
			name: "substr",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 2 {
					return &[]dataType{{dataType: String, value: SubstrEnd(ds, params[0], params[1])}}
				} else if len(params) == 3 {
					return &[]dataType{{dataType: String, value: Substr(ds, params[0], params[1], params[2])}}
				}
				return nil
			},
		},
		{
			name: "struct",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return &[]dataType{MakeStruct(ds, params...)}
			},
		},
		{
			name: "shift",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"shift\". Expected 1 found ", len(params))
				}
				return &[]dataType{Shift(ds, params[0])}
			},
		},
		{
			name: ".",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) < 2 {
					log.Fatal("Invalid number of parameters to \".\". Expected 2 or more found ", len(params))
				}
				return CallProp(ds, scopes, params)
			},
		},
		{
			name: "while",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				return WhileLoop(ds, scopes, params)
			},
		},
		{
			name: "++",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"++\", expected 1 found ", len(params))
				}
				return &[]dataType{AddOne(ds, params[0])}
			},
		},
		{
			name: "--",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"--\", expected 1 found ", len(params))
				}
				return &[]dataType{SubOne(ds, params[0])}
			},
		},
		{
			name: "+=",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"+=\", expected 2 found ", len(params))
				}
				return &[]dataType{AddMany(ds, params[0], params[1])}
			},
		},
		{
			name: "-=",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"-=\", expected 2 found ", len(params))
				}
				return &[]dataType{SubMany(ds, params[0], params[1])}
			},
		},
		{
			name: "require",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"require\", expected 1 found ", len(params))
				}

				file := GetFile(ds, params[0])
				tokens := Tokenize(string(file))
				Eval(ds, tokens, scopes-1)
				return nil
			},
		},
		{
			name: "from-char-code",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"from-char-code\", expected 1 found ", len(params))
				}
				return &[]dataType{FromCharCode(ds, params[0])}
			},
		},
		{
			name: "char-code-from",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"char-code-from\", expected 1 found ", len(params))
				}
				return &[]dataType{CharCodeFrom(ds, params[0])}
			},
		},
		{
			name: "split",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) == 1 {
					return &[]dataType{Split(ds, params[0], dataType{dataType: String, value: ""})}
				} else if len(params) == 2 {
					return &[]dataType{Split(ds, params[0], params[1])}
				} else {
					log.Fatal("Invalid number of parameters to \"split\", expected 1 or 2 found ", len(params))
					return nil
				}
			},
		},
		{
			name: "is-letter",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"is-letter\", expected 1 found ", len(params))
				}

				return &[]dataType{IsLetter(ds, params[0])}
			},
		},
		{
			name: "keys",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"keys\", expected 1 found ", len(params))
				}

				return &[]dataType{GetKeys(ds, params[0])}
			},
		},
		{
			name: "values",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"values\", expected 1 found ", len(params))
				}

				return &[]dataType{GetValues(ds, params[0])}
			},
		},
		{
			name: "floor",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"floor\", expected 1 found ", len(params))
				}

				return &[]dataType{Floor(ds, params[0])}
			},
		},
		{
			name: "ceil",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"ceil\", expected 1 found ", len(params))
				}

				return &[]dataType{Ceil(ds, params[0])}
			},
		},
		{
			name: "float",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"float\", expected 1 found ", len(params))
				}
				return &[]dataType{CastFloat(ds, params[0])}
			},
		},
		{
			name: "int",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"int\", expected 1 found ", len(params))
				}
				return &[]dataType{CastInt(ds, params[0])}
			},
		},
		{
			name: "string",
			fn: func(ds *dataStore, scopes int, params []dataType) *[]dataType {
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"string\", expected 1 found ", len(params))
				}
				return &[]dataType{CastString(ds, params[0])}
			},
		},
	}
}
