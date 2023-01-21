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
	"Identifier",
	"Func",
	"Nil",
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
)

type dataType struct {
	dataType DataType
	value    any
}

type variable struct {
	name string
	data dataType
}

type function struct {
	name   string
	body   []token
	params []dataType
}

type dataStore struct {
	vars             map[string][]variable
	scopedVars       [][]string
	scopedRedef      [][]string
	funcs            map[string][]function
	scopedFuncs      [][]string
	scopedRedefFuncs [][]string
	builtins         map[string]func(*dataStore, int, []dataType) (bool, []dataType)
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

func GetVariableFrom(name string, val dataType) variable {
	var variable variable
	variable.name = name
	variable.data = val
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
		for _, v := range funcArrToFree {
			FreeFunc(ds, v)
		}
		ds.scopedFuncs = ds.scopedFuncs[:len(ds.scopedFuncs)-1]
	}
}

func JoinTokens(tokens []token) string {
	res := ""
	for i, v := range tokens {
		if i > 0 {
			if tokens[i].tokenType == Identifier {
				if tokens[i-1].tokenType == CloseParen ||
					tokens[i-1].tokenType == CloseBracket ||
					tokens[i-1].tokenType == Identifier {
					res += " " + fmt.Sprint(v.value)
				} else {
					res += fmt.Sprint(v.value)
				}
			} else if (tokens[i-1].tokenType == Identifier ||
				tokens[i-1].tokenType == CloseParen ||
				tokens[i-1].tokenType == StringToken) &&
				tokens[i].tokenType != CloseParen &&
				tokens[i].tokenType != CloseBracket {
				res += " " + fmt.Sprint(v.value)
			} else {
				res += fmt.Sprint(v.value)
			}
		} else {
			res += fmt.Sprint(v.value)
		}
	}
	return res
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
	res := []token{{tokenType: OpenParen, value: "("}}
	for i := 1; i < len(f); i++ {
		if f[i].tokenType == OpenParen {
			parens++
			res = append(res, f[i])
		} else if f[i].tokenType == CloseParen {
			parens--
			res = append(res, f[i])
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
			d.value = t.value.(string)
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
	default:
		{
			log.Fatal("Cannot infer datatype from: ", t)
		}
	}
	return d
}

func EvalFunc(ds *dataStore, scopes int, info []dataType) (bool, []dataType) {
	return ds.builtins[info[0].value.(string)](ds, scopes, info[1:])
}

func Eval(ds *dataStore, code []token, scopes int, root bool) (bool, []dataType) {
	funcCall := [][]dataType{}
	funcNames := []string{}
	hasReturn := true
	toReturn := []dataType{}
	for i := 0; i < len(code); i++ {
		if len(funcNames) > 0 && funcNames[len(funcNames)-1] == "body" {
			funcCall = funcCall[:len(funcCall)-1]
			funcNames = funcNames[:len(funcNames)-1]
			i++
			bodyDataTokens, index := GetFuncEnd(code[i:])
			i += index + 2
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], dataType{dataType: Tokens, value: bodyDataTokens})
		}
		if code[i].tokenType == OpenParen {
			funcCall = append(funcCall, []dataType{})
			funcNames = append(funcNames, code[i+1].value.(string))
		} else if code[i].tokenType == CloseParen {
			if len(funcCall) == 0 {
				continue
			}
			funcReturns, val := EvalFunc(ds, len(funcCall)+scopes, funcCall[len(funcCall)-1])
			RemoveScopedVars(ds, len(funcCall)+scopes)
			funcCall = funcCall[:len(funcCall)-1]
			funcNames = funcNames[:len(funcNames)-1]
			if len(funcCall) > 0 && funcReturns {
				funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], val...)
			}
		} else if code[i].tokenType == OpenBracket {
			arr, index := GetArr(code[i:])
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], arr)
			i += index + 1
		} else if code[i].tokenType == Identifier ||
			code[i].tokenType == StringToken ||
			code[i].tokenType == IntToken ||
			code[i].tokenType == BoolToken ||
			code[i].tokenType == FloatToken {
			funcCall[len(funcCall)-1] = append(funcCall[len(funcCall)-1], GetDataTypeFromToken(code[i]))
		}
	}
	return hasReturn, toReturn
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
	ds.builtins = make(map[string]func(*dataStore, int, []dataType) (bool, []dataType))
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
			hasReturn, val := Eval(ds, Tokenize(line), 0, true)
			if hasReturn {
				fmt.Println(val)
			}
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
	Eval(ds, tokens, 0, true)
	if benchmark {
		evalEnd := time.Since(start)
		fmt.Println("\nFinished in", evalEnd)
	}
}
