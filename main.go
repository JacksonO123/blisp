package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type DataType int

var dataTypes []string = []string{
	"Int",
	"String",
	"Float",
	"Bool",
	"List",
	"Ident",
	"Func",
	"Nil",
	"Tokens",
	"Struct",
	"BreakVals",
	"ReturnVals",
	"Function",
}

const (
	Int DataType = iota
	String
	Float
	Bool
	List
	Ident
	Func
	Nil
	Tokens
	Struct
	BreakVal  // dataType
	ReturnVal // dataType
)

type dataType struct {
	dataType DataType
	value    any
}

type variable struct {
	name    string
	isConst bool
	data    dataType
}

type function struct {
	name   string
	body   []token
	params []dataType
}

type structAttr struct {
	name string
	attr *dataType
}

type builtin struct {
	name string
	fn   func(*dataStore, int, []dataType) *[]dataType
}

type dataStore struct {
	vars             map[string][]variable
	scopedVars       [][]string
	scopedRedef      [][]string
	funcs            map[string][]function
	scopedFuncs      [][]string
	scopedRedefFuncs [][]string
	builtins         []builtin
	inFunc           bool
	inLoop           bool
}

func StrArrIncludes(arr []string, val ...string) bool {
	for _, v := range arr {
		for _, check := range val {
			if check == v {
				return true
			}
		}
	}
	return false
}

func StrArrMap(arr []string, f func(string) string) []string {
	res := arr
	for i, v := range arr {
		res[i] = f(v)
	}
	return res
}

func TokensToString(tokens []token) []string {
	res := []string{}
	for _, v := range tokens {
		res = append(res, fmt.Sprint(v.value))
	}
	return res
}

func GetVariableFrom(name string, val dataType, isConst bool) variable {
	var variable variable
	variable.name = name
	variable.data = val
	variable.isConst = isConst
	return variable
}

func RemoveScopedVars(ds *dataStore, keepScopes int) {
	for keepScopes < len(ds.scopedVars) {
		varArrToFree := ds.scopedVars[len(ds.scopedVars)-1]
		varScopesToPop := ds.scopedRedef[len(ds.scopedRedef)-1]
		for _, v := range varScopesToPop {
			if len(ds.vars[v]) > 0 {
				ds.vars[v] = ds.vars[v][:len(ds.vars[v])-1]
			}
		}
		ds.scopedRedef = ds.scopedRedef[:len(ds.scopedRedef)-1]
		for _, v := range varArrToFree {
			FreeVar(ds, v)
		}
		ds.scopedVars = ds.scopedVars[:len(ds.scopedVars)-1]
	}
	for keepScopes < len(ds.scopedFuncs) {
		funcArrToFree := ds.scopedFuncs[len(ds.scopedFuncs)-1]
		funcScopesToPop := ds.scopedRedefFuncs[len(ds.scopedRedefFuncs)-1]
		for _, v := range funcScopesToPop {
			if len(ds.funcs[v]) > 0 {
				ds.funcs[v] = ds.funcs[v][:len(ds.funcs[v])-1]
			}
		}
		ds.scopedRedefFuncs = ds.scopedRedefFuncs[:len(ds.scopedRedefFuncs)-1]
		for _, v := range funcArrToFree {
			FreeFunc(ds, v)
		}
		ds.scopedFuncs = ds.scopedFuncs[:len(ds.scopedFuncs)-1]
	}
}

func GetStrSlice(str string) (string, int) {
	res := []rune{}
	for i, v := range str {
		res = append(res, v)
		if i > 0 {
			if v == '"' && str[i-1] != '\\' {
				return string(res), i
			}
		}
	}
	return "", -1
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
			} else {
				res = res[:i] + "\\" + res[i:]
				i++
			}
		}
	}
	return res
}

func GetFuncEnd(f []token) ([]token, int) {
	parens := 1
	res := []token{}
	for i := 1; i < len(f); i++ {
		if f[i].tokenType == OpenParen {
			parens++
			res = append(res, f[i])
		} else if f[i].tokenType == CloseParen {
			parens--
			if parens > 0 {
				res = append(res, f[i])
			}
		} else {
			res = append(res, f[i])
		}
		if parens == 0 {
			return res, i
		}
	}
	return []token{}, 0
}

func PrepQuotesString(str string) string {
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

func GetDataTypeFromToken(t token) dataType {
	var d dataType
	switch t.tokenType {
	case Identifier:
		{
			d.dataType = Ident
			d.value = t.value.(string)
		}
	case StringToken:
		{
			d.dataType = String
			str := t.value.(string)
			d.value = str[1 : len(str)-1]
		}
	case BoolToken:
		{
			d.dataType = Bool
			d.value = t.value.(bool)
		}
	case IntToken:
		{
			d.dataType = Int
			d.value = t.value.(int)
		}
	case FloatToken:
		{
			d.dataType = Float
			d.value = t.value.(float64)
		}
	case NilToken:
		{
			d.dataType = Nil
			d.value = nil
		}
	default:
		{
			log.Fatal("Cannot infer datatype from: ", t)
		}
	}
	return d
}

func IsBuiltin(ds *dataStore, name string) *builtin {
	for _, v := range ds.builtins {
		if v.name == name {
			return &v
		}
	}
	return nil
}

func EvalFunc(ds *dataStore, scopes int, info []dataType) (bool, *[]dataType) {
	ds.inFunc = true
	if info[0].dataType == Func {
		returnValue := CallInlineFunc(ds, scopes, "lambda", info[0].value.(function), info[1:])
		return true, returnValue
	} else if b := IsBuiltin(ds, info[0].value.(string)); b != nil {
		isCustom := false
		if b.name == "." {
			isCustom = true
		}
		return isCustom, b.fn(ds, scopes, info[1:])
	} else {
		v := CallFunc(ds, scopes, info[0], info[1:])
		return true, v
	}

	// f, ok := ds.builtins[info[0].value.(string)]; ok {
	// v := f(ds, scopes, info[1:])
	// isCustom := false
	// if info[0].value.(string) == "." {
	// 	isCustom = true
	// }
	// return isCustom, v
}

func Eval(ds *dataStore, code []token, scopes int) *[]dataType {
	funcCall := [][]dataType{}
	funcNames := []string{}
	var toReturn *[]dataType = nil
	reachedBlockEnd := false
	for i := 0; i < len(code); i++ {
		if len(funcNames) > 0 && funcNames[len(funcNames)-1] == "while" {
			if code[i+1].tokenType == OpenParen {
				funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], GetDataTypeFromToken(code[i]))
				conditionTokens, num := GetFuncEnd(code[i+1:])
				openParen := token{
					tokenType: OpenParen,
					value:     "(",
				}
				closeParen := token{
					tokenType: CloseParen,
					value:     ")",
				}
				conditionTokens = append([]token{openParen}, conditionTokens...)
				conditionTokens = append(conditionTokens, closeParen)
				tokens := dataType{dataType: Tokens, value: conditionTokens}
				i += num + 2
				funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], tokens)
			} else if code[i+1].tokenType == Identifier || code[i+1].tokenType == BoolToken {
				tokens := dataType{dataType: Tokens, value: []token{code[i+1]}}
				funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], GetDataTypeFromToken(code[i]))
				funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], tokens)
				i += 2
			}
		}
		if len(funcNames) > 0 && funcNames[len(funcNames)-1] == "body" {
			funcCall = funcCall[:len(funcCall)-1]
			funcNames = funcNames[:len(funcNames)-1]
			bodyDataTokens, num := GetFuncEnd(code[i:])
			tokens := dataType{dataType: Tokens, value: bodyDataTokens}
			i += num + 1
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], tokens)
		}
		if code[i].tokenType == OpenParen {
			funcCall = append(funcCall, []dataType{})
			funcNames = append(funcNames, code[i+1].value.(string))
		} else if code[i].tokenType == CloseParen {
			if len(funcCall) == 0 {
				continue
			}
			fromCustom, valP := EvalFunc(ds, len(funcCall)+scopes, funcCall[len(funcCall)-1])
			RemoveScopedVars(ds, len(funcCall)+scopes)
			if valP != nil {
				val := *valP
				if len(val) > 0 && (val[0].dataType == BreakVal || val[0].dataType == ReturnVal) {
					if !fromCustom {
						return valP
					} else {
						valP = &[]dataType{val[0].value.(dataType)}
					}
				}
			}

			funcCall = funcCall[:len(funcCall)-1]
			funcNames = funcNames[:len(funcNames)-1]
			if len(funcCall) > 0 {
				if valP != nil {
					funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], *valP...)
				}
			} else if len(funcCall) == 0 {
				toReturn = valP // do stuff
			}
		} else if code[i].tokenType == OpenBracket {
			arr, index := GetArr(code[i:]) // possibly eval here
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], arr)
			i += index + 1
		} else if (code[i].tokenType == Identifier ||
			code[i].tokenType == StringToken ||
			code[i].tokenType == IntToken ||
			code[i].tokenType == BoolToken ||
			code[i].tokenType == FloatToken ||
			code[i].tokenType == NilToken) && len(funcCall) > 0 {
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], GetDataTypeFromToken(code[i]))
		}
		if len(funcCall) == 0 {
			reachedBlockEnd = true
		} else if len(funcCall) > 0 && reachedBlockEnd {
			toReturn = nil
		}
	}
	return toReturn
}

var benchmark bool = false

func main() {
	args := os.Args[1:]
	scanner := bufio.NewScanner(os.Stdin)
	fileName := ""
	ds := new(dataStore)
	ds.vars = make(map[string][]variable)
	ds.funcs = make(map[string][]function)
	ds.scopedVars = [][]string{}
	ds.scopedRedef = [][]string{}
	ds.scopedFuncs = [][]string{}
	ds.scopedRedefFuncs = [][]string{}
	ds.builtins = []builtin{}
	ds.inFunc = false
	ds.inLoop = false
	InitBuiltins(ds)
	if len(args) > 0 {
		fileName = args[0]
		if !strings.Contains(fileName, ".blisp") {
			fileName += ".blisp"
		}
	} else {
		// repl
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT)
		go func() {
			<-sigs
			fmt.Println(" | Closing...")
			fmt.Println()
			os.Exit(0)
		}()
		for {
			fmt.Print("> ")
			line := ""
			scanner.Text()
			if scanner.Scan() {
				line = scanner.Text()
			}
			val := Eval(ds, Tokenize(line), 0)
			fmt.Println(val)
		}
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
	}
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	tokensStart := time.Now()
	tokens := Tokenize(string(dat))
	if benchmark {
		tokenEnd := time.Since(tokensStart)
		fmt.Println("Tokenized in", tokenEnd)
		fmt.Println()
	}
	Eval(ds, tokens, 0)
	if benchmark {
		evalEnd := time.Since(start)
		fmt.Println("\nFinished in", evalEnd)
	}
}
