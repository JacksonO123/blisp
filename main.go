package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fastjson/fastfloat"
)

type VariableType int

const (
	Int VariableType = iota
	String
	Float
	Bool
	List
)

type variable struct {
	variableType VariableType
	name         string
	value        string
}

type dataStore struct {
	vars        map[string][]variable
	scopedVars  [][]string
	scopedRedef [][]string
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func StrArrIncludes(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func StringArrMap(arr []string, f func(string) string) []string {
	newArr := arr
	for i := range arr {
		newArr[i] = f(newArr[i])
	}
	return newArr
}

func GetVariableInfo(name string, val string) variable {
	var variable variable
	variable.name = name
	if strings.Index(val, "\"") == 0 {
		variable.variableType = String
	} else if strings.Index(val, "[") == 0 {
		variable.variableType = List
	} else if val == "true" || val == "false" {
		variable.variableType = Bool
	} else {
		num, err := fastfloat.Parse(val)
		if err != nil {
			if math.Floor(num) != num {
				variable.variableType = Float
			} else {
				variable.variableType = Int
			}
		}
	}
	variable.value = val
	return variable
}

func SplitList(list string) []string {
	res := []string{}
	temp := []rune{}
	newList := list[1 : len(list)-1]
	inString := false
	nestedLists := 0
	for i, v := range newList {
		if v == '"' {
			if i > 0 {
				if newList[i-1] != '\\' {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if !inString {
			if v == ' ' {
				if nestedLists == 0 {
					res = append(res, string(temp))
					temp = []rune{}
				} else {
					temp = append(temp, v)
				}
			} else {
				temp = append(temp, v)
				if v == '[' {
					nestedLists++
				} else if v == ']' {
					nestedLists--
				}
			}
		} else {
			temp = append(temp, v)
		}
	}
	if len(temp) > 0 {
		res = append(res, string(temp))
	}
	return res
}

func GetBlocks(code string) []string {
	temp := []rune{}
	var blocks []string
	inString := false
	parenScopes := 0
	for i, c := range code {
		if c == '"' {
			if i > 0 {
				if code[i-1] != '\\' {
					inString = !inString
				}
			} else {
				inString = !inString
			}
			temp = append(temp, c)
			continue
		} else if c == '\t' {
			if len(temp) > 0 && temp[len(temp)-1] != ' ' {
				temp = append(temp, ' ')
			}
			continue
		} else if !inString {
			if c == '(' {
				parenScopes++
			} else if c == ')' {
				parenScopes--
			} else if c == '\n' {
				continue
			}
		}
		if parenScopes > 0 {
			if c == ' ' && len(temp) > 0 && temp[len(temp)-1] != ' ' {
				temp = append(temp, c)
			} else if c != ' ' {
				temp = append(temp, c)
			}
		} else if parenScopes == 0 && len(temp) > 0 {
			temp = append(temp, ')')
			temp = []rune(string(temp))
			blocks = append(blocks, string(temp))
			temp = []rune{}
		}
	}
	return blocks
}

func RemoveScopedVars(ds *dataStore, keepScopes int) {
	for keepScopes < len(ds.scopedVars) {
		arrToFree := ds.scopedVars[len(ds.scopedVars)-1]
		scopesToPop := ds.scopedRedef[len(ds.scopedRedef)-1]
		for _, v := range scopesToPop {
			if len(ds.vars[v]) > 0 {
				ds.vars[v] = ds.vars[v][:len(ds.vars[v])-1]
			}
		}
		for _, v := range arrToFree {
			FreeVar(ds, v)
		}
		ds.scopedVars = ds.scopedVars[:len(ds.scopedVars)-1]
	}
}

func Flatten(ds *dataStore, block string) string {
	res := block[1 : len(block)-1]
	inString := false
	starts := []int{}
	funcNames := []string{""}
	hasCurrentFunc := false
	for i := 0; i < len(res); i++ {
		if res[i] == '"' {
			if i > 0 {
				if res[i-1] != '\\' {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		} else if !inString {
			if res[i] == '(' {
				hasCurrentFunc = false
				funcNames = append(funcNames, "")
				starts = append(starts, i)
			} else if res[i] == ')' {
				hasCurrentFunc = false
				if StrArrIncludes(funcNames, "body") {
					if funcNames[len(funcNames)-1] == "body" {
						slice := res[starts[len(starts)-1]+6 : i]
						replaceWith := QuoteToQuoteLiteral(slice)
						res = res[:starts[len(starts)-1]] + "\"" + replaceWith + "\"" + res[i+1:]
						i -= len(slice) + 5 - len(replaceWith)
					}
					starts = starts[:len(starts)-1]
					funcNames = funcNames[:len(funcNames)-1]
				} else {
					slice := res[starts[len(starts)-1] : i+1]
					hasReturn, val := Eval(ds, slice, len(starts)+1)
					RemoveScopedVars(ds, len(starts)+1)
					if hasReturn {
						res = res[:starts[len(starts)-1]] + val + res[i+1:]
						i -= len(slice) - len(val)
						starts = starts[:len(starts)-1]
					}
				}
			} else if res[i] == ' ' {
				if !hasCurrentFunc {
					hasCurrentFunc = true
				}
			} else if !hasCurrentFunc {
				funcNames[len(funcNames)-1] += string(res[i])
			}
		}
	}
	res = FixQuoteLiterals(res)
	return res
}

func SplitParams(str string) []string {
	res := []string{}
	temp := ""
	inString := false
	inArr := false
	strChars := strings.Split(str, "")
	for i, v := range strChars {
		if v == "\"" {
			if i > 0 {
				if strChars[i-1] != "\\" {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if !inString {
			if v == "[" {
				inArr = true
				temp += v
			} else if v == "]" {
				inArr = false
				temp += v
				res = append(res, temp)
				temp = ""
			} else if inArr {
				if v == " " && temp[len(temp)-1] == ',' {
					continue
				} else {
					temp += v
				}
			} else {
				if v == " " {
					if len(temp) > 0 {
						res = append(res, temp)
					}
					temp = ""
				} else {
					temp += v
				}
			}
		} else {
			temp += v
		}
	}
	if len(temp) > 0 {
		res = append(res, temp)
	}
	return res
}

func QuoteLiteralToQuote(str string) string {
	return strings.Replace(str, "\\\"", "\"", -1)
}

func QuoteToQuoteLiteral(str string) string {
	res := str
	for i := 0; i < len(res); i++ {
		if res[i] == '"' {
			if i > 0 {
				if res[i-1] == '\\' {
					continue
				} else {
					res = res[:i] + "\\" + res[i:]
					i++
				}
			}
		}
	}
	return res
}

func FixQuoteLiterals(str string) string {
	temp := str
	startLen := len(temp)
	diff := startLen - len(strings.ReplaceAll(temp, "\"", ""))
	res := str
	quoteNum := 0
	for i := 0; i < len(res); i++ {
		if res[i] == '"' {
			quoteNum++
			if i > 0 {
				if res[i-1] == '\\' {
					if quoteNum == 1 || quoteNum == diff {
						res = res[:i-1] + res[i:]
					}
					continue
				}
			}
		}
	}
	return res
}

func Eval(ds *dataStore, code string, scopes int) (bool, string) {
	blocks := GetBlocks(code)
	hasReturn := true
	toReturn := ""
	if len(blocks) != 1 {
		for _, block := range blocks {
			Eval(ds, FixQuoteLiterals(block), scopes+1)
			RemoveScopedVars(ds, scopes+1)
		}
	} else {
		blocks[0] = FixQuoteLiterals(blocks[0])
		flatBlock := Flatten(ds, blocks[0])
		parts := SplitParams(flatBlock)
		params := parts[1:]
		switch parts[0] {
		case "print":
			{
				toReturn = "\"(printing " + QuoteToQuoteLiteral(QuoteLiteralToQuote(strings.Join(params, ", "))) + ")\""
				Print(ds, parts[1:]...)
			}
		case "+":
			{
				toReturn = fmt.Sprint(Add(ds, params...))
			}
		case "-":
			{
				toReturn = fmt.Sprint(Sub(ds, params...))
			}
		case "*":
			{
				toReturn = fmt.Sprint(Mult(ds, params...))
			}
		case "/":
			{
				toReturn = fmt.Sprint(Divide(ds, params...))
			}
		case "^":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"^\". Expected 2 found", len(params))
				}
				toReturn = fmt.Sprint(Exp(ds, params[0], params[1]))
			}
		case "%":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"%\". Expected 2 found", len(params))
				}
				toReturn = fmt.Sprint(Mod(ds, params[0], params[1]))
			}
		case "eval":
			{
				if len(params) == 1 {
					hasReturn, toReturn = Eval(ds, params[0][1:len(params[0])-1], scopes)
					if !hasReturn {
						toReturn = "\"(evaluating " + QuoteToQuoteLiteral(params[0]) + ")\""
					}
				} else {
					toReturn = "\"(evaluating " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
					for _, v := range params {
						if len(v) > 0 {
							Eval(ds, v[1:len(v)-1], scopes)
						}
					}
				}
			}
		case "var":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"var\". Expected 2 found", len(params))
				}
				toReturn = "\"(initializing " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
				MakeVar(ds, scopes, params[0], params[1])
			}
		case "set":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"set\". Expected 2 found", len(params))
				}
				toReturn = "\"(setting " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
				SetVar(ds, params[0], params[1])
			}
		case "free":
			{
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"free\". Expected 1 found", len(params))
				}
				toReturn = "\"(freeing " + QuoteToQuoteLiteral(params[0]) + ")\""
				FreeVar(ds, params[0])
			}
		case "type":
			{
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"type\". Expected 1 found", len(params))
				}
				toReturn = "\"" + GetValueType(ds, params[0]) + "\""
			}
		case "get":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"get\". Expected 2 found", len(params))
				}
				toReturn = GetValueFromList(ds, params[0], params[1])
			}
		case "loop":
			{
				if len(params) == 3 {
					valType := GetValueType(ds, params[0])
					if valType == "List" {
						LoopListIterator(ds, scopes, params[0], params[1], params[2])
						toReturn = "\"(looping over " + params[0] + ")\""
					} else if valType == "Int" {
						LoopTo(ds, scopes, params[0], params[1], params[2])
						toReturn = "\"(looping to " + params[0] + ")\""
					} else {
						log.Fatal("Expecting first param to be \"List\" or \"Int\", got:", valType)
					}
				} else if len(params) == 4 {
					valType := GetValueType(ds, params[0])
					if valType == "List" {
						LoopListIndexIterator(ds, scopes, params[0], params[1], params[2], params[3])
						toReturn = "\"(looping over " + params[0] + ")\""
					} else if valType == "Int" {
						LoopFromTo(ds, scopes, params[0], params[1], params[2], params[3])
						toReturn = "\"(looping from " + params[0] + " to " + params[1] + ")\""
					} else {
						log.Fatal("Expecting first param to be list, got:", valType)
					}
				}
			}
		case "scan-line":
			{
				if len(params) == 0 {
					line := ""
					fmt.Scanln(&line)
					toReturn = line
				} else if len(params) == 1 {
					line := ""
					fmt.Scanln(&line)
					if _, ok := ds.vars[params[0]]; ok {
						SetVar(ds, params[0], line)
						toReturn = "\"(setting " + params[0] + " to " + line + ")\""
					} else {
						log.Fatal("Unable to assign value to", params[0])
					}
				} else {
					log.Fatal("Invalid number of parameters to \"scan-line\". Expected 0 or 2 found ", len(params))
				}
			}
		case "if":
			{
				if len(params) == 2 {
					info := GetValue(ds, params[0])
					if info.variableType == Bool {
						if val, err := strconv.ParseBool(info.value); err == nil && val {
							Eval(ds, params[1][1:len(params[1])-1], scopes)
						}
					} else {
						log.Fatal("Error in \"if\", expected type: \"Bool\" found ", info.variableType)
					}
				} else {
					log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
				}
			}
		case "eq":
			{
				if len(params) > 0 {
					eq := true
					for i := 0; i < len(params)-1; i++ {
						if GetValue(ds, params[i]).value != GetValue(ds, params[i+1]).value {
							eq = false
						}
					}
					toReturn = fmt.Sprint(eq)
				} else {
					log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found", len(params))
				}
			}
		default:
			{
				hasReturn = false
				fmt.Println("default", parts)
			}
		}
	}
	return hasReturn, toReturn
}

func main() {
	args := os.Args[1:]
	fileName := ""
	benchmark := false
	if len(args) > 0 {
		fileName = args[0]
		if !strings.Contains(fileName, ".blisp") {
			fileName += ".blisp"
		}
	} else {
		fileName = "main.blisp"
	}
	fmt.Println("Running [" + fileName + "]")
	if len(args) > 1 {
		flags := args[1:]
		if StrArrIncludes(flags, "-b") {
			benchmark = true
		}
	}
	fileStart := time.Now()
	dat, err := os.ReadFile(fileName)
	if benchmark {
		fmt.Println("["+fileName+"] read in", time.Since(fileStart))
		fmt.Println()
	}
	check(err)
	ds := new(dataStore)
	ds.vars = make(map[string][]variable)
	ds.scopedVars = [][]string{}
	ds.scopedRedef = [][]string{}
	start := time.Now()
	Eval(ds, string(dat), 0)
	if benchmark {
		fmt.Println("\nFinished in", time.Since(start))
	}
}
