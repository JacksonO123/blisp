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
		inString = StringToggle(newList, i, inString)
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

func GetScopeEnd(block string) int {
	scopes := 1
	inString := false
	for i, v := range block {
		inString = StringToggle(block, i, inString)
		if !inString {
			if v == '(' {
				scopes++
			} else if v == ')' {
				scopes--
			}
		}
		if scopes == 0 {
			return i
		}
	}
	return len(block) - 1
}

func Flatten(ds *dataStore, block string) string {
	res := block[1 : len(block)-1]
	inString := false
	starts := []int{}
	funcName := ""
	hasCurrentFunc := false
	for i := 0; i < len(res); i++ {
		inString = StringToggle(res, i, inString)
		if !inString {
			if res[i] == '(' {
				hasCurrentFunc = false
				funcName = ""
				starts = append(starts, i)
			} else if res[i] == ')' {
				hasCurrentFunc = false
				if len(starts) > 0 {
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
					if funcName == "body" {
						i += GetScopeEnd(res[i:])
					}
				}
			} else if !hasCurrentFunc {
				funcName += string(res[i])
			}
		}
	}
	res = FixQuoteLiterals(res)
	return res
}

func StringToggle(str string, index int, inString bool) bool {
	if str[index] == '"' {
		if index > 0 {
			if str[index-1] != '\\' {
				return !inString
			}
		} else {
			return !inString
		}
	}
	return inString
}

func SplitParams(str string) []string {
	res := []string{}
	temp := []rune{}
	inString := false
	inArr := false
	parens := 0
	brackets := 0
	hasFuncName := false
	funcName := ""
	for i := 0; i < len(str); i++ {
		inString = StringToggle(str, i, inString)
		if !inString {
			if inArr {
				if str[i] == '[' {
					brackets++
				} else if str[i] == ']' {
					brackets--
				} else if str[i] == ' ' && temp[len(temp)-1] == ' ' {
					continue
				}
				temp = append(temp, rune(str[i]))
				if brackets == 0 {
					inArr = false
					temp = []rune("[" + strings.TrimSpace(string(temp[1:len(temp)-1])) + "]")
				}
			} else if str[i] == '(' {
				parens++
				hasFuncName = false
				funcName = ""
			} else if str[i] == ')' {
				parens--
				hasFuncName = false
				funcName = ""
			} else if str[i] == '[' {
				brackets++
				inArr = true
				temp = append(temp, rune(str[i]))
			} else if str[i] == ']' {
				inArr = false
				temp = []rune(strings.TrimSpace(string(temp)))
				temp = append(temp, rune(str[i]))
				res = append(res, string(temp))
				temp = []rune{}
			} else if str[i] == ' ' {
				if !hasFuncName {
					hasFuncName = true
					if funcName == "body" {
						start := i
						i += GetScopeEnd(str[start:])
						res = append(res, str[start-4:i])
						temp = []rune{}
					}
				}
				if len(temp) > 0 {
					res = append(res, string(temp))
				}
				temp = []rune{}
			} else {
				temp = append(temp, rune(str[i]))
				if !hasFuncName {
					funcName += string(str[i])
				}
			}
		} else {
			temp = append(temp, rune(str[i]))
		}
	}
	if len(temp) > 0 {
		res = append(res, string(temp))
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
		hasReturn, toReturn = HandleFunc(ds, scopes, flatBlock, parts...)
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
