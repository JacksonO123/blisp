package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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
	vars map[string]variable
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
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			numStr := fmt.Sprint(num)
			if strings.Index(numStr, ".") > 0 {
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
	temp := ""
	var blocks []string
	inString := false
	parenScopes := 0
	chars := strings.Split(code, "")
	for i, v := range chars {
		if Eq(v, "\"") {
			if i > 0 {
				if chars[i-1] != "\\" {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if Eq(v, "\t") {
			if len(temp) > 0 && !Eq(strings.Split(temp, "")[len(temp)-1], " ") {
				temp += " "
			}
			continue
		}
		if !inString {
			if Eq(v, "(") {
				parenScopes++
			} else if Eq(v, ")") {
				parenScopes--
			} else if Eq(v, "\n") {
				continue
			}
		}
		if parenScopes > 0 {
			if Eq(v, " ") && len(temp) > 0 && !Eq(strings.Split(temp, "")[len(temp)-1], " ") {
				temp += v
			} else if !Eq(v, " ") {
				temp += v
			}
		}
		if parenScopes == 0 && strings.TrimSpace(temp) != "" {
			temp += ")"
			temp = strings.Replace(temp, "\\\"", "\"", -1)
			blocks = append(blocks, temp)
			temp = ""
		}
	}
	return blocks
}

func Flatten(ds *dataStore, block string) string {
	res := block[1 : len(block)-1]
	inString := false
	var starts []int
	chars := strings.Split(res, "")
	for i := 0; i < len(res); i++ {
		if Eq(chars[i], "\"") {
			if i > 0 {
				if chars[i-1] != "\\" {
					inString = !inString
				}
			} else {
				inString = !inString
			}
		}
		if !inString {
			if Eq(chars[i], "(") {
				starts = append(starts, i)
			} else if Eq(chars[i], ")") {
				slice := res[starts[len(starts)-1] : i+1]
				hasReturn, val := Eval(ds, slice)
				if hasReturn {
					start := starts[len(starts)-1]
					res = res[:start] + val + res[i+1:]
					chars = strings.Split(res, "")
					diff := len(slice) - len(val)
					i -= diff
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
				if Eq(v, " ") && Eq(strings.Split(temp, "")[len(temp)-1], ",") {
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

func Eval(ds *dataStore, code string) (bool, string) {
	blocks := GetBlocks(code)
	hasReturn := true
	toReturn := ""
	if len(blocks) != 1 {
		for i, block := range blocks {
			_, blocks[i] = Eval(ds, block)
		}
	} else {
		flatBlock := Flatten(ds, blocks[0])
		parts := SplitParams(flatBlock)
		params := parts[1:]
		switch parts[0] {
		case "print":
			{
				toReturn = "\"(printing " + strings.Join(params, ", ") + ")\""
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
		case "%":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"%\". Expected 2 found " + fmt.Sprint(len(params)))
				}
				toReturn = fmt.Sprint(Mod(ds, params[0], params[1]))
			}
		case "eval":
			{
				if len(params) == 1 {
					hasReturn, toReturn = Eval(ds, params[0][1:len(params[0])-1])
					if !hasReturn {
						toReturn = "\"(evaluating " + QuoteToQuoteLiteral(params[0]) + ")\""
					}
				} else {
					toReturn = "\"(evaluating " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
					for _, v := range params {
						if len(v) > 0 {
							Eval(ds, v[1:len(v)-1])
						}
					}
				}
			}
		case "var":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"var\". Expected 2 found " + fmt.Sprint(len(params)))
				}
				toReturn = "\"(initializing " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
				MakeVar(ds, params[0], params[1])
			}
		case "set":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"set\". Expected 2 found " + fmt.Sprint(len(params)))
				}
				toReturn = "\"(setting " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
				SetVar(ds, params[0], params[1])
			}
		case "free":
			{
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"free\". Expected 1 found " + fmt.Sprint(len(params)))
				}
				toReturn = "\"(freeing " + QuoteToQuoteLiteral(params[0]) + ")\""
				FreeVar(ds, params[0])
			}
		case "type":
			{
				if len(params) != 1 {
					log.Fatal("Invalid number of parameters to \"type\". Expected 1 found " + fmt.Sprint(len(params)))
				}
				toReturn = "\"" + GetValueType(ds, params[0]) + "\""
			}
		case "get":
			{
				if len(params) != 2 {
					log.Fatal("Invalid number of parameters to \"get\". Expected 2 found " + fmt.Sprint(len(params)))
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
	if len(args) > 0 {
		fileName = args[0]
		if strings.Index(fileName, ".blisp") < 0 {
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
	dat, err := os.ReadFile(fileName)
	check(err)
	ds := new(dataStore)
	ds.vars = make(map[string]variable)
	start := time.Now()
	Eval(ds, string(dat))
	if benchmark {
		fmt.Println("Finished in " + fmt.Sprint(time.Now().Sub(start)))
	}
}
