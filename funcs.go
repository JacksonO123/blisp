package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/valyala/fastjson/fastfloat"
)

func FormatPrint(str string) string {
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

func GetValueFromList(ds *dataStore, index string, list string) string {
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
			log.Fatal(list + " is not a list")
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
		Eval(ds, body, scopes)
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
		Eval(ds, body, scopes)
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
		Eval(ds, body, scopes)
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
		Eval(ds, body, scopes)
	}
}
