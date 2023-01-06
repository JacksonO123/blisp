package main

import (
	"fmt"
	"log"
	"math"
	"os"
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
	evalCache   map[string]string
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Eq(s1 string, s2 string) bool {
	return strings.Compare(s1, s2) == 0
}

func StrArrIncludes(arr []string, val string) bool {
	for _, v := range arr {
		if Eq(v, val) {
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
	} else if Eq(val, "true") || Eq(val, "false") {
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
	temp := ""
	listChars := strings.Split(list[1:len(list)-1], "")
	inString := false
	nestedLists := 0
	for i, v := range listChars {
		if Eq(v, "\"") {
			if i > 0 {
				if listChars[i-1] != "\\" {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if !inString {
			if Eq(v, ",") {
				if nestedLists == 0 {
					res = append(res, temp)
					temp = ""
				} else {
					temp += v
				}
			} else {
				temp += v
				if Eq(v, "[") {
					nestedLists++
				} else if Eq(v, "]") {
					nestedLists--
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
				temp = append(temp, c)
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
			temp = []rune(QuoteLiteralToQuote(string(temp)))
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

func Flatten(ds *dataStore, block string, caching bool) string {
	res := block[1 : len(block)-1]
	inString := false
	starts := []int{}
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
				starts = append(starts, i)
			} else if res[i] == ')' {
				slice := res[starts[len(starts)-1] : i+1]
				hasReturn, val := Eval(ds, slice, caching, len(starts)+1)
				RemoveScopedVars(ds, len(starts)+1)
				if hasReturn {
					res = res[:starts[len(starts)-1]] + val + res[i+1:]
					if len(starts)-1 == 0 {
						return res
					}
					i -= len(slice) - len(val)
					starts = starts[:len(starts)-1]
				}
			}
		}
	}
	return res
}

func SplitParams(str string) []string {
	res := []string{}
	temp := ""
	inString := false
	inArr := false
	strChars := strings.Split(str, "")
	for i, v := range strChars {
		if Eq(v, "\"") {
			if i > 0 {
				if strChars[i-1] != "\\" {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if !inString {
			if Eq(v, "[") {
				inArr = true
				temp += v
			} else if Eq(v, "]") {
				inArr = false
				temp += v
				res = append(res, temp)
				temp = ""
			} else if inArr {
				if Eq(v, " ") && temp[len(temp)-1] == ',' {
					continue
				} else {
					temp += v
				}
			} else {
				if Eq(v, " ") {
					res = append(res, temp)
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

func QuoteToQuoteLiteral(str string) string {
	return strings.Replace(str, "\"", "\\\"", -1)
}

func QuoteLiteralToQuote(str string) string {
	return strings.Replace(str, "\\\"", "\"", -1)
}

func Eval(ds *dataStore, code string, caching bool, scopes int) (bool, string) {
	blocks := GetBlocks(code)
	hasReturn := true
	toReturn := ""
	if len(blocks) != 1 {
		for _, block := range blocks {
			Eval(ds, block, caching, scopes+1)
			RemoveScopedVars(ds, scopes+1)
		}
	} else {
		flatBlock := ""
		if caching {
			if val, ok := ds.evalCache[blocks[0]]; ok {
				flatBlock = val
			} else {
				flatBlock = Flatten(ds, blocks[0], caching)
				ds.evalCache[blocks[0]] = flatBlock
			}
		} else {
			flatBlock = Flatten(ds, blocks[0], caching)
			ds.evalCache[blocks[0]] = flatBlock
		}
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
					hasReturn, toReturn = Eval(ds, params[0][1:len(params[0])-1], caching, scopes)
					if !hasReturn {
						toReturn = "\"(evaluating " + QuoteToQuoteLiteral(params[0]) + ")\""
					}
				} else {
					toReturn = "\"(evaluating " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
					for _, v := range params {
						if len(v) > 0 {
							Eval(ds, v[1:len(v)-1], caching, scopes)
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
	caching := false
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
		if StrArrIncludes(flags, "-c") {
			caching = true
		}
	}
	fileStart := time.Now()
	dat, err := os.ReadFile(fileName)
	if benchmark {
		fmt.Println("["+fileName+"] read in", time.Since(fileStart))
	}
	check(err)
	ds := new(dataStore)
	ds.vars = make(map[string][]variable)
	ds.scopedVars = [][]string{}
	ds.scopedRedef = [][]string{}
	ds.evalCache = make(map[string]string)
	start := time.Now()
	Eval(ds, string(dat), caching, 0)
	if benchmark {
		fmt.Println("Finished in", time.Since(start))
	}
}
