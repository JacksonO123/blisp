package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/valyala/fastjson/fastfloat"
)

var reserved []string = []string{
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
	"or",
	"not",
	"func",
	"return",
}

func HandleFunc(ds *dataStore, scopes int, parts ...token) (bool, []token) {
	params := parts[1:]
	hasReturn := true
	toReturn := []token{}
	if val, ok := ds.builtins[parts[0].value]; ok {
		hasReturn, toReturn = val(ds, scopes, params)
	} else {
		hasReturn = false
		if _, ok := ds.funcs[parts[0].value]; ok {
			ds.inFunc = true
			return CallFunc(ds, scopes+1, parts...)
		} else {
			fmt.Println("default", "["+strings.Join(TokensToValue(parts), ", ")+"]")
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

func Print(ds *dataStore, params ...token) {
	res := []string{}
	for _, v := range params {
		if v.tokenType == StringToken {
			res = append(res, v.value)
			continue
		}

		info := GetValue(ds, v)
		res = append(res, info.value)
	}
	fmt.Println(FormatPrint(strings.Join(res, ", ")))
}

func GetFloat64FromToken(ds *dataStore, tk token) float64 {
	n, err := fastfloat.Parse(tk.value)
	if err != nil {
		if val, ok := ds.vars[tk.value]; ok {
			if val[len(val)-1].variableType == Int || val[len(val)-1].variableType == Float {
				n, _ = fastfloat.Parse(val[len(val)-1].value)
			} else {
				log.Fatal(err)
			}
		}
	}
	return n
}

func GetFloat64FromTokens(ds *dataStore, tokens ...token) []float64 {
	var nums []float64
	for _, v := range tokens {
		n := GetFloat64FromToken(ds, v)
		nums = append(nums, n)
	}
	return nums
}

func Add(ds *dataStore, params ...token) float64 {
	var res float64 = 0
	for _, v := range params {
		val := GetFloat64FromToken(ds, v)
		res += val
	}
	return res
}

func Sub(ds *dataStore, params ...token) float64 {
	nums := GetFloat64FromTokens(ds, params...)
	var res float64 = nums[0]
	if len(params) == 1 {
		res *= -1
	} else {
		for _, v := range nums[1:] {
			res -= v
		}
	}
	return res
}

func Mult(ds *dataStore, params ...token) float64 {
	nums := GetFloat64FromTokens(ds, params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res *= v
	}
	return res
}

func Divide(ds *dataStore, params ...token) float64 {
	nums := GetFloat64FromTokens(ds, params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res /= v
	}
	return res
}

func Exp(ds *dataStore, base token, exp token) float64 {
	num1 := GetFloat64FromToken(ds, base)
	num2 := GetFloat64FromToken(ds, exp)
	return math.Pow(num1, num2)
}

func Mod(ds *dataStore, num1 token, num2 token) int {
	val1 := GetFloat64FromToken(ds, num1)
	val2 := GetFloat64FromToken(ds, num2)
	if math.Floor(val1) != val1 {
		log.Fatal(val1, " is not an int")
	} else if math.Floor(val2) != val2 {
		log.Fatal(val2, " is not an int")
	}
	return int(val1) % int(val2)
}

func MakeVar(ds *dataStore, scopes int, name token, val token) {
	if scopes < len(ds.scopedVars) && StrArrIncludes(ds.scopedVars[scopes], name.value) {
		log.Fatal("Variable already initialized: ", name.value)
		return
	}

	if StrArrIncludes(reserved, name.value) {
		log.Fatal("Variable name \"", name.value, "\" is reserved")
		return
	}

	if len(val.value) == 0 {
		log.Fatal("Variable must have name")
		return
	}

	if name.tokenType == NumberToken {
		log.Fatal("Variable named ", name.value, " cannot be a number")
		return
	}

	ds.vars[name.value] = append(ds.vars[name.value], GetVariableInfo(name.value, val))
	for len(ds.scopedVars) < scopes {
		ds.scopedVars = append(ds.scopedVars, []string{})
	}
	for len(ds.scopedRedef) < scopes {
		ds.scopedRedef = append(ds.scopedRedef, []string{})
	}
	if _, ok := ds.vars[name.value]; ok {
		ds.scopedRedef[scopes-1] = append(ds.scopedRedef[scopes-1], name.value)
	} else {
		ds.scopedVars[scopes-1] = append(ds.scopedVars[scopes-1], name.value)
	}
}

func SetVar(ds *dataStore, name token, val token) {
	if _, ok := ds.vars[name.value]; !ok {
		log.Fatal("Variable not initialized: ", name.value)
		return
	}
	ds.vars[name.value][len(ds.vars[name.value])-1] = GetValue(ds, val)
}

func FreeVar(ds *dataStore, name token) {
	if _, ok := ds.vars[name.value]; !ok {
		log.Fatal("Unable to free, variable not initialized: ", name)
		return
	}
	delete(ds.vars, name.value)
}

func FreeFunc(ds *dataStore, name string) {
	if _, ok := ds.funcs[name]; !ok {
		log.Fatal("Unable to free, variable not initialized: " + name)
		return
	}
	delete(ds.funcs, name)
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

func GetValue(ds *dataStore, val token) variable {
	if v, ok := ds.vars[val.value]; ok {
		return v[len(v)-1]
	} else if v, ok := ds.funcs[val.value]; ok {
		f := v[len(v)-1]
		res := GetVariableInfo("", GetToken("\"("+f.params.value+" ("+f.body.value+"))\""))
		return res
	}
	res := GetVariableInfo("", val)
	return res
}

func GetValueType(ds *dataStore, val token) string {
	if v, ok := ds.vars[val.value]; ok {
		return GetType(v[len(v)-1].variableType)
	}
	tempVar := GetVariableInfo("", val)
	return GetType(tempVar.variableType)
}

func GetValueFromList(ds *dataStore, list token, index token) token {
	if v, ok := ds.vars[list.value]; ok {
		parts := SplitList(v[len(v)-1].value)
		intIndex, err := strconv.Atoi(GetValue(ds, index).value)
		if err != nil {
			log.Fatal(err)
		}
		return parts[intIndex]
	} else {
		tempVar := GetVariableInfo("", list)
		if tempVar.variableType != List {
			log.Fatal("Error getting ", index.value, " from ", list.value, ", ", list.value, " is not a list")
		} else {
			parts := SplitList(tempVar.value)
			intIndex, err := strconv.Atoi(index.value)
			if err != nil {
				log.Fatal(err)
			}
			return parts[intIndex]
		}
	}
	return token{}
}

func LoopListIterator(ds *dataStore, scopes int, list token, iteratorName token, body []token) {
	listVar := GetValue(ds, list)
	parts := SplitList(listVar.value)
	made := false
	for _, v := range parts {
		if !made {
			MakeVar(ds, scopes+1, iteratorName, v)
		} else {
			SetVar(ds, iteratorName, v)
		}
		hasReturn, val := Eval(ds, body, scopes, false)
		if hasReturn {
			if val[0] == GetToken("break") {
				break
			}
		}
	}
}

func LoopListIndexIterator(ds *dataStore, scopes int, list token, indexIterator token, iteratorName token, body []token) {
	listVar := GetValue(ds, list)
	parts := SplitList(listVar.value)
	made := false
	for i, v := range parts {
		if !made {
			MakeVar(ds, scopes+1, iteratorName, v)
			MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
		} else {
			SetVar(ds, iteratorName, v)
			SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
		}
		hasReturn, val := Eval(ds, body, scopes, false)
		if hasReturn {
			if val[0] == GetToken("break") {
				break
			}
		}
	}
}

func LoopTo(ds *dataStore, scopes int, max token, indexIterator token, body []token) {
	maxNum := int(GetFloat64FromToken(ds, max))
	made := false
	for i := 0; i < maxNum; i++ {
		if !made {
			MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
		} else {
			SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
		}
		hasReturn, val := Eval(ds, body, scopes, false)
		if hasReturn {
			if val[0] == GetToken("break") {
				break
			}
		}
	}
}

func LoopFromTo(ds *dataStore, scopes int, start token, max token, indexIterator token, body []token) {
	startNum := int(GetFloat64FromToken(ds, start))
	maxNum := int(GetFloat64FromToken(ds, max))
	made := false
	i := startNum
	next := func() {
		if startNum <= maxNum {
			i++
		} else {
			i--
		}
	}
	comp := func() bool {
		if startNum <= maxNum {
			return i < maxNum
		} else {
			return i > maxNum
		}
	}
	for ; comp(); next() {
		if !made {
			MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
		} else {
			SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
		}
		hasReturn, val := Eval(ds, body, scopes, false)
		if hasReturn {
			if val[0] == GetToken("break") {
				break
			}
		}
	}
}

func Eq(ds *dataStore, params ...token) bool {
	eq := true
	for i := 0; i < len(params)-1; i++ {
		if GetValue(ds, params[i]).value != GetValue(ds, params[i+1]).value {
			eq = false
		}
	}
	return eq
}

func If(ds *dataStore, scopes int, params ...token) (bool, []token) {
	hasReturn := true
	toReturn := []token{}
	info := GetValue(ds, params[0])
	if info.variableType == Bool {
		if val, err := strconv.ParseBool(info.value); err == nil && val {
			hasReturn, toReturn = Eval(ds, Tokenize(params[1].value), scopes, false)
		} else if len(params) == 3 {
			hasReturn, toReturn = Eval(ds, Tokenize(params[2].value), scopes, false)
		}
	} else {
		log.Fatal("Error in \"if\", expected type: \"Bool\" found ", info.variableType)
	}
	return hasReturn, toReturn
}

func AppendToList(list token, toAppend token) token {
	val := list.value
	tokens := Tokenize(val[1 : len(val)-1])
	res := append(tokens, toAppend)
	return GetToken(JoinList(res))
}

func PrependToList(list token, toPrepend token) token {
	val := list.value
	tokens := Tokenize(val[1 : len(val)-1])
	res := append([]token{toPrepend}, tokens...)
	return GetToken(JoinList(res))
}

func ListFunc(ds *dataStore, f func(list token, val token) token, params ...token) token {
	info := GetValue(ds, params[0])
	if info.variableType == List {
		list := GetToken(info.value)
		for _, v := range params[1:] {
			toAppend := GetValue(ds, v)
			if toAppend.variableType == List {
				parts := SplitList(toAppend.value)
				for i := 0; i < len(parts); i++ {
					list = f(list, parts[i])
				}
			} else {
				list = f(list, GetToken(toAppend.value))
			}
		}
		return list
	} else {
		log.Fatal("Error in \"append\", expected type \"List\" found ", info.variableType)
	}
	return GetToken("[]")
}

func Concat(ds *dataStore, params ...token) string {
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

func Pop(ds *dataStore, list token) token {
	val := GetValue(ds, list)
	if val.variableType != List {
		log.Fatal("Error in \"pop\" expected \"List\" found ", val.variableType)
	}
	listItems := SplitList(val.value)
	if len(listItems) > 0 {
		lastItem := listItems[len(listItems)-1]
		listItems = listItems[:len(listItems)-1]
		if _, ok := ds.vars[list.value]; ok {
			SetVar(ds, list, GetToken(JoinList(listItems)))
		}
		return lastItem
	} else {
		return GetToken("")
	}
}

func Remove(ds *dataStore, list token, index token) token {
	val := GetValue(ds, list)
	listIndex := int(GetFloat64FromToken(ds, index))
	if val.variableType != List {
		log.Fatal("Error in \"remove\" expected \"List\" found ", val.variableType)
	}
	listItems := SplitList(val.value)
	if len(listItems) > 0 {
		item := listItems[listIndex]
		listItems = append(listItems[:listIndex], listItems[listIndex+1:]...)
		if _, ok := ds.vars[list.value]; ok {
			SetVar(ds, list, GetToken(JoinList(listItems)))
		}
		return item
	} else {
		return GetToken("")
	}
}

func Len(ds *dataStore, list token) int {
	parts := SplitList(GetValue(ds, list).value)
	return len(parts)
}

func And(ds *dataStore, params ...token) bool {
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

func SetIndex(ds *dataStore, list token, index token, val token) token {
	info := GetValue(ds, list)
	if info.variableType != List {
		log.Fatal("Error in \"set\", expected list found ", info.variableType)
	}
	parts := SplitList(info.value)
	listIndex := int(GetFloat64FromToken(ds, index))
	parts[listIndex] = GetToken(GetValue(ds, val).value)
	newList := JoinList(parts)
	var listToken token
	listToken.tokenType = ListToken
	listToken.value = newList
	if _, ok := ds.vars[list.value]; ok {
		SetVar(ds, list, listToken)
	}
	return listToken
}

func Or(ds *dataStore, params ...token) bool {
	res := false
	for i, v := range params {
		info := GetValue(ds, v)
		if info.variableType != Bool {
			log.Fatal("Error in \"or\", expected \"Bool\" found ", info.variableType)
		}
		val, err := strconv.ParseBool(info.value)
		if i == 0 {
			if err == nil {
				res = val
			}
		} else {
			res = res || val
		}
	}
	return res
}

func Not(ds *dataStore, val token) bool {
	info := GetValue(ds, val)
	if info.variableType != Bool {
		log.Fatal("Error in \"not\", expected \"Bool\" found ", info.variableType)
	}
	v, err := strconv.ParseBool(info.value)
	if err == nil {
		return !v
	}
	return true
}

func GetFunctionBody(params ...token) (token, int) {
	for i, v := range params {
		if v.tokenType == UnTokenized {
			return v, i + 1
		}
	}
	return GetToken(""), -1
}

func MakeFunction(ds *dataStore, scopes int, params ...token) {
	var f function
	body, index := GetFunctionBody(params[1:]...)
	f.body = body
	f.name = params[0].value
	paramList := GetVariableInfo("", GetToken("["+strings.Join(TokensToValue(params[1:index]), " ")+"]"))
	f.params = paramList

	if scopes < len(ds.scopedFuncs) && StrArrIncludes(ds.scopedFuncs[scopes], f.name) {
		log.Fatal("Function already initialized: " + f.name)
		return
	}

	if StrArrIncludes(reserved, f.name) {
		log.Fatal("Function name \"" + f.name + "\" is reserved")
		return
	}

	_, err := fastfloat.Parse(f.name)
	if err == nil {
		log.Fatal("Function named " + f.name + " cannot be a number")
		return
	}

	ds.funcs[f.name] = append(ds.funcs[f.name], f)
	for len(ds.scopedFuncs) < scopes {
		ds.scopedFuncs = append(ds.scopedFuncs, []string{})
	}
	for len(ds.scopedRedefFuncs) < scopes {
		ds.scopedRedefFuncs = append(ds.scopedRedefFuncs, []string{})
	}
	if _, ok := ds.funcs[f.name]; ok {
		ds.scopedRedefFuncs[scopes-1] = append(ds.scopedRedefFuncs[scopes-1], f.name)
	} else {
		ds.scopedFuncs[scopes-1] = append(ds.scopedFuncs[scopes-1], f.name)
	}
}

func CallFunc(ds *dataStore, scopes int, params ...token) (bool, []token) {
	name := params[0]
	inputs := params[1:]
	f := ds.funcs[name.value][len(ds.funcs[name.value])-1]
	funcParams := SplitList(f.params.value)
	for i, v := range funcParams {
		MakeVar(ds, scopes, v, GetToken(GetValue(ds, inputs[i]).value))
	}
	hasReturn, toReturn := Eval(ds, Tokenize(f.body.value), scopes, false)
	ds.inFunc = false
	return hasReturn, toReturn
}
