package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/valyala/fastjson/fastfloat"
)

func HandleFunc(ds *dataStore, scopes int, flatBlock string, parts ...string) (bool, string) {
	params := parts[1:]
	hasReturn := true
	toReturn := ""
	switch parts[0] {
	case "print":
		{
			toReturn = "\"(printing " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
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
				log.Fatal("Invalid number of parameters to \"^\". Expected 2 found ", len(params))
			}
			toReturn = fmt.Sprint(Exp(ds, params[0], params[1]))
		}
	case "%":
		{
			if len(params) != 2 {
				log.Fatal("Invalid number of parameters to \"%\". Expected 2 found ", len(params))
			}
			toReturn = fmt.Sprint(Mod(ds, params[0], params[1]))
		}
	case "eval":
		{
			if len(params) == 1 {
				hasReturn, toReturn = Eval(ds, params[0][1:len(params[0])-1], scopes, false)
				if !hasReturn {
					toReturn = "\"(evaluating " + QuoteToQuoteLiteral(params[0]) + ")\""
				}
			} else {
				toReturn = "\"(evaluating " + QuoteToQuoteLiteral(strings.Join(params, ", ")) + ")\""
				for _, v := range params {
					if len(v) > 0 {
						Eval(ds, v[1:len(v)-1], scopes, false)
					}
				}
			}
		}
	case "var":
		{
			if len(params) != 2 {
				log.Fatal("Invalid number of parameters to \"var\". Expected 2 found ", len(params))
			}
			toReturn = "\"(initializing " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
			MakeVar(ds, scopes, params[0], params[1])
		}
	case "set":
		{
			if len(params) == 2 {
				toReturn = "\"(setting " + QuoteToQuoteLiteral(params[0]) + " to " + QuoteToQuoteLiteral(params[1]) + ")\""
				SetVar(ds, params[0], params[1])
			} else if len(params) == 3 {
			} else {
				log.Fatal("Invalid number of parameters to \"set\". Expected 2 or 3 found ", len(params))
			}
		}
	case "free":
		{
			if len(params) != 1 {
				log.Fatal("Invalid number of parameters to \"free\". Expected 1 found ", len(params))
			}
			toReturn = "\"(freeing " + QuoteToQuoteLiteral(params[0]) + ")\""
			FreeVar(ds, params[0])
		}
	case "type":
		{
			if len(params) != 1 {
				log.Fatal("Invalid number of parameters to \"type\". Expected 1 found ", len(params))
			}
			toReturn = "\"" + GetValueType(ds, params[0]) + "\""
		}
	case "get":
		{
			if len(params) != 2 {
				log.Fatal("Invalid number of parameters to \"get\". Expected 2 found ", len(params))
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
			if len(params) == 2 || len(params) == 3 {
				hasReturn, toReturn = If(ds, scopes, params...)
			} else {
				log.Fatal("Invalid number of parameters to \"if\". Expected 2 found ", len(params))
			}
		}
	case "eq":
		{
			if len(params) > 0 {
				toReturn = fmt.Sprint(Eq(ds, params...))
			} else {
				log.Fatal("Invalid number of parameters to \"eq\". Expected 1 or more found ", len(params))
			}
		}
	case "body":
		{
			hasReturn, toReturn = Eval(ds, flatBlock[6:len(flatBlock)-1], scopes, false)
		}
	case "append":
		{
			if len(params) < 2 {
				log.Fatal("Invalid number of parameters to \"append\". Expected 2 or more found ", len(params))
			}
			res := ListFunc(ds, AppendToList, params...)
			if _, ok := ds.vars[params[0]]; ok {
				SetVar(ds, params[0], res)
				toReturn = "\"(appending [" + QuoteToQuoteLiteral(strings.Join(params[1:], ",")) + "] to " + params[0] + ")\""
			} else {
				toReturn = res
			}
		}
	case "prepend":
		{
			if len(params) < 2 {
				log.Fatal("Invalid number of parameters to \"prepend\". Expected 2 or more found ", len(params))
			}
			res := ListFunc(ds, PrependToList, params...)
			if _, ok := ds.vars[params[0]]; ok {
				SetVar(ds, params[0], res)
				toReturn = "\"(prepending [" + QuoteToQuoteLiteral(strings.Join(params[1:], " ")) + "] to " + params[0] + ")\""
			} else {
				toReturn = res
			}
		}
	case "concat":
		{
			toReturn = "\"" + Concat(ds, params...) + "\""
		}
	case "exit":
		{
			if len(params) != 0 {
				log.Fatal("Invalid number of parameters to \"exit\". Expected 0 found ", len(params))
			}
			os.Exit(0)
		}
	case "break":
		{
			toReturn = "(break)"
		}
	case "pop":
		{
			if len(params) != 1 {
				log.Fatal("Invalid number of parameters to \"pop\". Expected 1 found ", len(params))
			}
			toReturn = Pop(ds, params[0])
		}
	case "remove":
		{
			if len(params) != 2 {
				log.Fatal("Invalid number of parameters to \"remove\". Expected 2 found ", len(params))
			}
			toReturn = Remove(ds, params[0], params[1])
		}
	case "len":
		{
			if len(params) != 1 {
				log.Fatal("Invalid number of parameters to \"len\". Expected 1 found ", len(params))
			}
			toReturn = fmt.Sprint(Len(ds, params[0]))
		}
	case "and":
		{
			if len(params) == 0 {
				log.Fatal("Invalid number of parameters to \"and\". Expected 1 or more found ", len(params))
			}
			toReturn = fmt.Sprint(And(ds, params...))
		}
	default:
		{
			hasReturn = false
			fmt.Println("default", parts)
		}
	}
	return hasReturn, toReturn
}

func FormatPrint(str string) string {
	if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	return QuoteLiteralToQuote(str)
}

func Print(ds *dataStore, params ...string) {
	res := []string{}
	for _, v := range params {
		if strings.Contains(v, "\"") {
			res = append(res, v)
			continue
		}
		_, err := fastfloat.Parse(v)
		if err == nil {
			res = append(res, v)
		} else {
			if v == "true" || v == "false" {
				res = append(res, v)
			} else if val, ok := ds.vars[v]; ok {
				res = append(res, val[len(val)-1].value)
			} else {
				log.Fatal("Unknown value: " + v)
			}
		}
	}
	fmt.Println(FormatPrint(strings.Join(res, ", ")))
}

func GetFloat64FromString(ds *dataStore, str string) float64 {
	n, err := fastfloat.Parse(str)
	if err != nil {
		if val, ok := ds.vars[str]; ok {
			if val[len(val)-1].variableType == Int || val[len(val)-1].variableType == Float {
				n, _ = fastfloat.Parse(val[len(val)-1].value)
			} else {
				log.Fatal(err)
			}
		}
	}
	return n
}

func GetFloat64FromStrings(ds *dataStore, strs ...string) []float64 {
	var nums []float64
	for _, v := range strs {
		n := GetFloat64FromString(ds, v)
		nums = append(nums, n)
	}
	return nums
}

func Add(ds *dataStore, params ...string) float64 {
	nums := GetFloat64FromStrings(ds, params...)
	var res float64 = 0
	for _, v := range nums {
		res += v
	}
	return res
}

func Sub(ds *dataStore, params ...string) float64 {
	nums := GetFloat64FromStrings(ds, params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res -= v
	}
	return res
}

func Mult(ds *dataStore, params ...string) float64 {
	nums := GetFloat64FromStrings(ds, params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res *= v
	}
	return res
}

func Divide(ds *dataStore, params ...string) float64 {
	nums := GetFloat64FromStrings(ds, params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res /= v
	}
	return res
}

func Exp(ds *dataStore, base string, exp string) float64 {
	num1 := GetFloat64FromString(ds, base)
	num2 := GetFloat64FromString(ds, exp)
	return math.Pow(num1, num2)
}

func Mod(ds *dataStore, num1 string, num2 string) int {
	val1 := GetFloat64FromString(ds, num1)
	val2 := GetFloat64FromString(ds, num2)
	val1Str := fmt.Sprint(val1)
	val2Str := fmt.Sprint(val2)
	v1 := GetVariableInfo("", val1Str)
	v2 := GetVariableInfo("", val2Str)
	if v1.variableType != Int {
		log.Fatal(val1Str + " is not an int")
	} else if v2.variableType != Int {
		log.Fatal(val2Str + " is not an int")
	}
	return int(val1) % int(val2)
}

func MakeVar(ds *dataStore, scopes int, name string, val string) {
	if scopes < len(ds.scopedVars) && StrArrIncludes(ds.scopedVars[scopes], name) {
		log.Fatal("Variable already initialized: " + name)
		return
	}
	reserved := []string{
		"print",
		"+",
		"-",
		"*",
		"/",
		"%",
		"eval",
		"var",
		"set",
		"free",
		"type",
		"get",
		"true",
		"false",
		"loop",
		"body",
		"scan-line",
		"if",
		"eq",
		"append",
		"prepend",
		"concat",
		"exit",
		"break",
		"pop",
		"remove",
		"len",
		"and",
	}
	if StrArrIncludes(reserved, name) {
		log.Fatal("Variable name \"" + name + "\" is reserved")
		return
	}

	if len(val) == 0 {
		log.Fatal("Variable must have name")
		return
	}

	_, err := fastfloat.Parse(name)
	if err == nil {
		log.Fatal("Variable named " + name + " cannot be a number")
		return
	}

	ds.vars[name] = append(ds.vars[name], GetVariableInfo(name, val))
	for len(ds.scopedVars) < scopes {
		ds.scopedVars = append(ds.scopedVars, []string{})
	}
	for len(ds.scopedRedef) < scopes {
		ds.scopedRedef = append(ds.scopedRedef, []string{})
	}
	if _, ok := ds.vars[name]; ok {
		ds.scopedRedef[scopes-1] = append(ds.scopedRedef[scopes-1], name)
	} else {
		ds.scopedVars[scopes-1] = append(ds.scopedVars[scopes-1], name)
	}
}

func SetVar(ds *dataStore, name string, val string) {
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Variable not initialized: " + name)
		return
	}
	ds.vars[name][len(ds.vars[name])-1] = GetValue(ds, val)
}

func FreeVar(ds *dataStore, name string) {
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Unable to free, variable not initialized: " + name)
		return
	}
	delete(ds.vars, name)
}

func GetType(val VariableType) string {
	res := "Invalid Type"
	switch val {
	case Int:
		res = "Int"
	case String:
		res = "String"
	case Float:
		res = "Float"
	case Bool:
		res = "Bool"
	case List:
		res = "List"
	}
	return res
}

func GetValue(ds *dataStore, val string) variable {
	if v, ok := ds.vars[val]; ok {
		return v[len(v)-1]
	}
	return GetVariableInfo("", val)
}

func GetValueType(ds *dataStore, val string) string {
	if v, ok := ds.vars[val]; ok {
		return GetType(v[len(v)-1].variableType)
	}
	tempVar := GetVariableInfo("", val)
	return GetType(tempVar.variableType)
}

func GetValueFromList(ds *dataStore, list string, index string) string {
	if v, ok := ds.vars[list]; ok {
		parts := SplitList(v[len(v)-1].value)
		intIndex, err := strconv.Atoi(GetValue(ds, index).value)
		if err != nil {
			log.Fatal(err)
		}
		return parts[intIndex]
	} else {
		tempVar := GetVariableInfo("", list)
		if tempVar.variableType != List {
			log.Fatal("Error getting " + index + " from " + list + ", " + list + " is not a list")
		} else {
			parts := SplitList(tempVar.value)
			intIndex, err := strconv.Atoi(index)
			if err != nil {
				log.Fatal(err)
			}
			return parts[intIndex]
		}
	}
	return ""
}

func LoopListIterator(ds *dataStore, scopes int, list string, iteratorName string, body string) {
	listVar := GetValue(ds, list)
	parts := SplitList(listVar.value)
	made := false
	for _, v := range parts {
		if !made {
			MakeVar(ds, scopes+1, iteratorName, v)
		} else {
			SetVar(ds, iteratorName, v)
		}
		Eval(ds, body, scopes, false)
	}
}

func LoopListIndexIterator(ds *dataStore, scopes int, list string, indexIterator string, iteratorName string, body string) {
	listVar := GetValue(ds, list)
	parts := SplitList(listVar.value)
	made := false
	for i, v := range parts {
		if !made {
			MakeVar(ds, scopes+1, iteratorName, v)
			MakeVar(ds, scopes+1, indexIterator, fmt.Sprint(i))
		} else {
			SetVar(ds, iteratorName, v)
			SetVar(ds, indexIterator, fmt.Sprint(i))
		}
		Eval(ds, body, scopes, false)
	}
}

func LoopTo(ds *dataStore, scopes int, max string, indexIterator string, body string) {
	maxNum := int(GetFloat64FromString(ds, max))
	made := false
	for i := 0; i < maxNum; i++ {
		if !made {
			MakeVar(ds, scopes+1, indexIterator, fmt.Sprint(i))
		} else {
			SetVar(ds, indexIterator, fmt.Sprint(i))
		}
		hasReturn, val := Eval(ds, body, scopes, false)
		if hasReturn {
			if val == "(break)" {
				break
			}
		}
	}
}

func LoopFromTo(ds *dataStore, scopes int, start string, max string, indexIterator string, body string) {
	startNum := int(GetFloat64FromString(ds, start))
	maxNum := int(GetFloat64FromString(ds, max))
	made := false
	for i := startNum; i < maxNum; i++ {
		if !made {
			MakeVar(ds, scopes+1, indexIterator, fmt.Sprint(i))
		} else {
			SetVar(ds, indexIterator, fmt.Sprint(i))
		}
		Eval(ds, body, scopes, false)
	}
}

func Eq(ds *dataStore, params ...string) bool {
	eq := true
	for i := 0; i < len(params)-1; i++ {
		if GetValue(ds, params[i]).value != GetValue(ds, params[i+1]).value {
			eq = false
		}
	}
	return eq
}

func If(ds *dataStore, scopes int, params ...string) (bool, string) {
	hasReturn := true
	toReturn := ""
	info := GetValue(ds, params[0])
	if info.variableType == Bool {
		if val, err := strconv.ParseBool(info.value); err == nil && val {
			hasReturn, toReturn = Eval(ds, params[1][1:len(params[1])-1], scopes, false)
		} else if len(params) == 3 {
			hasReturn, toReturn = Eval(ds, params[2][1:len(params[2])-1], scopes, false)
		}
	} else {
		log.Fatal("Error in \"if\", expected type: \"Bool\" found ", info.variableType)
	}
	return hasReturn, toReturn
}

func AppendToList(list string, toAppend string) string {
	res := list
	appendTo := res[:len(res)-1]
	if len(appendTo) == 1 {
		res = res[:len(res)-1] + toAppend + "]"
	} else {
		res = res[:len(res)-1] + " " + toAppend + "]"
	}
	return res
}

func PrependToList(list string, toPrepend string) string {
	res := list
	res = "[" + toPrepend + " " + res[1:]
	return res
}

func ListFunc(ds *dataStore, f func(list string, val string) string, params ...string) string {
	info := GetValue(ds, params[0])
	if info.variableType == List {
		list := info.value
		for _, v := range params[1:] {
			toAppend := GetValue(ds, v)
			if toAppend.variableType == List {
				parts := SplitList(toAppend.value)
				for i := 0; i < len(parts); i++ {
					list = f(list, parts[i])
				}
			} else {
				list = f(list, toAppend.value)
			}
		}
		return list
	} else {
		log.Fatal("Error in \"append\", expected type \"List\" found ", info.variableType)
	}
	return "[]"
}

func Concat(ds *dataStore, params ...string) string {
	res := ""
	for _, v := range params {
		info := GetValue(ds, v)
		if info.variableType == String {
			res += info.value[1 : len(info.value)-1]
		} else {
			res += info.value
		}
	}
	return res
}

func Pop(ds *dataStore, list string) string {
	val := GetValue(ds, list)
	if val.variableType != List {
		log.Fatal("Error in \"pop\" expected \"List\" found ", val.variableType)
	}
	listItems := SplitList(val.value)
	if len(listItems) > 0 {
		lastItem := listItems[len(listItems)-1]
		listItems = listItems[:len(listItems)-1]
		if _, ok := ds.vars[list]; ok {
			SetVar(ds, list, strings.Join(listItems, " "))
		}
		return lastItem
	} else {
		return ""
	}
}

func Remove(ds *dataStore, list string, index string) string {
	val := GetValue(ds, list)
	listIndex := int(GetFloat64FromString(ds, index))
	if val.variableType != List {
		log.Fatal("Error in \"remove\" expected \"List\" found ", val.variableType)
	}
	listItems := SplitList(val.value)
	if len(listItems) > 0 {
		item := listItems[listIndex]
		listItems = append(listItems[:listIndex], listItems[listIndex+1:]...)
		if _, ok := ds.vars[list]; ok {
			SetVar(ds, list, strings.Join(listItems, " "))
		}
		return item
	} else {
		return ""
	}
}

func Len(ds *dataStore, list string) int {
	parts := SplitList(GetValue(ds, list).value)
	return len(parts)
}

func And(ds *dataStore, params ...string) bool {
	res := false
	for i, v := range params {
		info := GetValue(ds, v)
		if info.variableType != Bool {
			log.Fatal("Error in \"and\", expected \"Bool\" found ", info.variableType)
		}
		val, err := strconv.ParseBool(info.value)
		if err == nil {
			if i == 0 {
				res = val
			} else {
				if !res || !val {
					res = false
				}
			}
		}
	}
	return res
}
