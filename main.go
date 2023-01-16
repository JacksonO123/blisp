package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/valyala/fastjson/fastfloat"
)

type VariableType string

const (
	Int    VariableType = "Int"
	String VariableType = "String"
	Float  VariableType = "Float"
	Bool   VariableType = "Bool"
	List   VariableType = "List"
)

type variable struct {
	variableType VariableType
	name         string
	value        string
}

type function struct {
	name   string
	body   token
	params variable
}

type dataStore struct {
	vars             map[string][]variable
	scopedVars       [][]string
	scopedRedef      [][]string
	funcs            map[string][]function
	scopedFuncs      [][]string
	scopedRedefFuncs [][]string
	builtins         map[string]func(*dataStore, int, []token) (bool, []token)
	inFunc           bool
	inLoop           bool
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
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

func StringArrMap(arr []string, f func(string) string) []string {
	newArr := arr
	for i := range arr {
		newArr[i] = f(newArr[i])
	}
	return newArr
}

func TokensToValue(tokens []token) []string {
	res := []string{}
	for _, v := range tokens {
		res = append(res, v.value)
	}
	return res
}

func GetVariableInfo(name string, val token) variable {
	var variable variable
	variable.name = name
	if val.tokenType == StringToken {
		variable.variableType = String
	} else if val.tokenType == ListToken {
		variable.variableType = List
	} else if val.tokenType == BoolToken {
		variable.variableType = Bool
	} else {
		num, err := fastfloat.Parse(val.value)
		if err == nil {
			if math.Floor(num) != num {
				variable.variableType = Float
			} else {
				variable.variableType = Int
			}
		} else {
			log.Fatal("Unknown value: ", val.value)
		}
	}
	variable.value = val.value
	return variable
}

// fix this
func SplitList(list string) []string {
	res := []string{}
	return res
}

func GetBlocks(code []token) [][]token {
	temp := []token{}
	var blocks [][]token
	parenScopes := 0
	for _, c := range code {
		if c.tokenType == OpenParen {
			parenScopes++
		} else if c.tokenType == CloseParen {
			parenScopes--
		}
		temp = append(temp, c)
		if parenScopes == 0 && len(temp) > 0 {
			blocks = append(blocks, temp)
			temp = []token{}
		}
	}
	return blocks
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
			t := GetToken(v)
			FreeVar(ds, t)
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

func GetScopeEnd(block []token) int {
	scopes := 1
	for i, v := range block {
		if v.tokenType == OpenParen {
			scopes++
		} else if v.tokenType == CloseParen {
			scopes--
		}
		if scopes == 0 {
			return i
		}
	}
	return len(block)
}

func JoinTokens(tokens []token) string {
	res := ""
	for i, v := range tokens {
		if len(res) > 0 {
			if tokens[i-1].tokenType == Identifier {
				res += " " + v.value
			} else {
				res += v.value
			}
		} else {
			res += v.value
		}
	}
	return res
}

func Flatten(ds *dataStore, block []token) []token {
	res := block[1 : len(block)-1]
	starts := []int{}
	for i := 0; i < len(res); i++ {
		if res[i].tokenType == OpenParen {
			starts = append(starts, i)
		} else if res[i].tokenType == CloseParen {
			slice := res[starts[len(starts)-1] : i+1]
			hasReturn, val := Eval(ds, slice, len(starts)+1, false)
			RemoveScopedVars(ds, len(starts)+1)
			if hasReturn {
				if JoinTokens(val) == "(break)" {
					if ds.inLoop {
						return []token{GetToken("break")}
					} else {
						log.Fatal("Not in loop, cannot break")
					}
				} else if len(val) > 1 {
					if val[1].value == "return" {
						if ds.inFunc {
							return val
						} else {
							log.Fatal("Not in func, cannot return")
						}
					}
				}
				start := res[:starts[len(starts)-1]]
				end := res[i+1:]
				tempRes := append(start, val...)
				tempRes = append(tempRes, end...)
				res = tempRes
				i -= len(slice) - len(val)
			}
			starts = starts[:len(starts)-1]
		} else if res[i].tokenType == Identifier {
			if i > 0 && res[i-1].tokenType == OpenParen && res[i].value == "body" {
				bodyStart := i - 1
				i += GetScopeEnd(res[i:]) + 1
				body := res[bodyStart:i]
				var t token
				t.tokenType = UnTokenized
				t.value = JoinTokens(body)

				start := res[:starts[len(starts)-1]]
				end := res[i:]
				tempRes := append(start, t)
				tempRes = append(tempRes, end...)
				res = tempRes

				starts = starts[:len(starts)-1]
			}
		}
	}

	return res
}

func GetArr(str string) (string, int) {
	brackets := 0
	for i, v := range str {
		if v == '[' {
			brackets++
		} else if v == ']' {
			brackets--
		}

		if brackets == 0 {
			return str[:i+1], i + 1
		}
	}
	return "", -1
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

func Eval(ds *dataStore, code []token, scopes int, root bool) (bool, []token) {
	blocks := GetBlocks(code)
	hasReturn := true
	toReturn := []token{}
	if len(blocks) > 1 {
		hasReturn = false
		for _, block := range blocks {
			blockReturn, blockVal := Eval(ds, block, scopes+1, false)
			if blockReturn {
				if JoinTokens(blockVal) == "(break)" {
					hasReturn = true
					toReturn = blockVal
					break
				}
			}
			RemoveScopedVars(ds, scopes+1)
		}
	} else {
		parts := Flatten(ds, blocks[0])
		if root {
			scopes++
		}
		hasReturn, toReturn = HandleFunc(ds, scopes, parts...)
	}
	return hasReturn, toReturn
}

func main() {
	args := os.Args[1:]
	scanner := bufio.NewScanner(os.Stdin)
	fileName := ""
	benchmark := false
	ds := new(dataStore)
	ds.vars = make(map[string][]variable)
	ds.funcs = make(map[string][]function)
	ds.scopedVars = [][]string{}
	ds.scopedRedef = [][]string{}
	ds.scopedFuncs = [][]string{}
	ds.scopedRedefFuncs = [][]string{}
	ds.builtins = make(map[string]func(*dataStore, int, []token) (bool, []token))
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
		fmt.Println()
	}
	check(err)
	start := time.Now()
	Eval(ds, Tokenize(string(dat)), 0, true)
	if benchmark {
		fmt.Println("\nFinished in", time.Since(start))
	}
}
