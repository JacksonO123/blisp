package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func FormatPrint(str string) string {
	return QuoteLiteralToQuote(str)
}

func Print(ds *dataStore, params ...string) {
	res := []string{}
	for _, v := range params {
		if strings.Index(v, "\"") > -1 {
			res = append(res, v)
			continue
		}
		_, err := strconv.ParseFloat(v, 64)
		if err == nil {
			res = append(res, v)
		} else {
			if val, ok := ds.vars[v]; ok {
				res = append(res, val.value)
			} else {
				log.Fatal("Unknown value: " + v)
			}
		}
	}
	fmt.Println(FormatPrint(strings.Join(res, ", ")))
}

func GetFloat64FromString(ds *dataStore, str string) float64 {
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		if val, ok := ds.vars[str]; ok {
			if val.variableType == Int || val.variableType == Float {
				n, _ = strconv.ParseFloat(val.value, 64)
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

func MakeVar(ds *dataStore, name string, val string) {
	if _, ok := ds.vars[name]; ok {
		log.Fatal("Variable already initialized: " + name)
		return
	}
	reserved := []string{"print", "+", "-", "*", "/", "%", "eval", "var", "set", "free", "type", "get"}
	if StrArrIncludes(reserved, val) {
		log.Fatal("Variable name " + val + " is reserved")
		return
	}

	if len(val) == 0 {
		log.Fatal("Variable must have name")
		return
	}

	_, err := strconv.ParseFloat(name, 64)
	if err == nil {
		log.Fatal("Variable named " + name + " cannot be a number")
		return
	}

	ds.vars[name] = GetVariableInfo(name, val)
}

func SetVar(ds *dataStore, name string, val string) {
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Variable not initialized: " + name)
		return
	}
	ds.vars[name] = GetVariableInfo(name, val)
}

func FreeVar(ds *dataStore, name string) {
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Unable to free, variable not initialized: " + name)
		return
	}
	delete(ds.vars, name)
}

func GetType(val VariableType) string {
	res := ""
	switch val {
	case Int:
		{
			res = "Int"
		}
	case String:
		{
			res = "String"
		}
	case Float:
		{
			res = "Float"
		}
	case Bool:
		{
			res = "Bool"
		}
	case List:
		{
			res = "List"
		}
	default:
		{
			res = "Invalid type"
		}
	}
	return res
}

func GetValueType(ds *dataStore, val string) string {
	if v, ok := ds.vars[val]; ok {
		return GetType(v.variableType)
	}
	tempVar := GetVariableInfo("", val)
	return GetType(tempVar.variableType)
}

func GetValueFromList(ds *dataStore, index string, list string) string {
	if v, ok := ds.vars[list]; ok {
		parts := SplitList(v.value)
		intIndex, err := strconv.Atoi(index)
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
