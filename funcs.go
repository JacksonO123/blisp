package main

import (
	"fmt"
	"log"
	"math"
	"os"
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

func GetArr(tokens []token) (dataType, int) {
	res := []dataType{}
	index := 0
	for i := 1; i < len(tokens); i++ {
		if tokens[i].tokenType == OpenBracket {
			arr, indx := GetArr(tokens[i:])
			i += indx + 1
			index += 2
			res = append(res, dataType{dataType: List, value: arr})
		} else if tokens[i].tokenType == CloseBracket {
			return dataType{dataType: List, value: res}, index
		} else {
			res = append(res, GetDataTypeFromToken(tokens[i]))
		}
		index++
	}
	return dataType{dataType: Nil, value: nil}, 0
}

func GetArrStr(data dataType) dataType {
	var d dataType
	d.dataType = String
	res := "["
	arr := data.value.([]dataType)
	for i, v := range arr {
		if v.dataType == List {
			res += GetArrStr(v.value.(dataType)).value.(string)
		} else {
			res += fmt.Sprint(v.value)
		}
		if i < len(arr)-1 {
			res += " "
		}
	}
	res += "]"
	return dataType{dataType: String, value: res}
}

func PrintArr(data dataType) {
	arr := data.value.([]dataType)
	printAt := 224
	items := 0
	toPrint := "["
	for i, v := range arr {
		if v.dataType == List {
			if i == 0 {
				fmt.Print(toPrint)
			} else {
				fmt.Print(toPrint + " ")
			}
			toPrint = ""
			items = 0
			PrintArr(v.value.(dataType))
		} else {
			if i > 0 {
				toPrint += " "
			}
			toPrint += fmt.Sprint(v.value)
			items++
		}

		if items == printAt {
			fmt.Print(toPrint)
			toPrint = ""
			items = 0
		}
	}
	if len(toPrint) > 0 {
		fmt.Print(toPrint)
	}
	fmt.Print("]")
}

func Print(ds *dataStore, params ...dataType) {
	for i, v := range params {
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
			if v.dataType == Ident {
				log.Fatal("Unknown value: ", v.value.(string))
			}
		}
		if v.dataType == List {
			PrintArr(v)
		} else {
			os.Stdout.Write([]byte(fmt.Sprint(v.value)))
		}
		if i < len(params)-1 {
			os.Stdout.Write([]byte(", "))
		}
	}
	os.Stdout.Write([]byte("\n"))
}

// value of dataType passed in as a string
func GetStrValue(data dataType) string {
	res := ""
	if data.dataType == List {
		res = GetArrStr(data).value.(string)
	} else {
		res = fmt.Sprint(data.value)
	}
	return res
}

func Add(ds *dataStore, params ...dataType) dataType {
	var res float64 = 0
	for _, v := range params {
		info := v
		if info.dataType == Ident {
			info = GetDsValue(ds, info)
		}
		if info.dataType == Int {
			res += float64(info.value.(int))
		} else if info.dataType == Float {
			res += info.value.(float64)
		} else {
			log.Fatal("Cannot + type ", dataTypes[info.dataType])
		}
	}
	var d dataType
	if math.Floor(res) == res {
		d.dataType = Int
		d.value = int(res)
	} else {
		d.dataType = Float
		d.value = res
	}
	return d
}

func Sub(ds *dataStore, params ...dataType) dataType {
	var res float64 = 0
	firstVal := params[0]
	if firstVal.dataType == Ident {
		firstVal = GetDsValue(ds, firstVal)
	}
	if firstVal.dataType == Int {
		res = float64(firstVal.value.(int))
	} else if firstVal.dataType == Float {
		res = firstVal.value.(float64)
	} else {
		log.Fatal("Cannot - type ", dataTypes[firstVal.dataType])
	}
	if len(params) == 1 {
		res *= -1
	} else {
		for _, v := range params[1:] {
			info := v
			if info.dataType == Ident {
				info = GetDsValue(ds, info)
			}
			if info.dataType == Int {
				res -= float64(info.value.(int))
			} else if info.dataType == Float {
				res -= info.value.(float64)
			} else {
				log.Fatal("Cannot - type ", dataTypes[info.dataType])
			}
		}
	}
	var d dataType
	if math.Floor(res) == res {
		d.dataType = Int
		d.value = int(res)
	} else {
		d.dataType = Float
		d.value = res
	}
	return d
}

func Mult(ds *dataStore, params ...dataType) dataType {
	var res float64 = 0
	firstVal := params[0]
	if firstVal.dataType == Ident {
		firstVal = GetDsValue(ds, firstVal)
	}
	if firstVal.dataType == Int {
		res = float64(firstVal.value.(int))
	} else if firstVal.dataType == Float {
		res = firstVal.value.(float64)
	} else {
		log.Fatal("Cannot * type ", dataTypes[firstVal.dataType])
	}
	for _, v := range params[1:] {
		if v.dataType == Int {
			res *= float64(v.value.(int))
		} else if v.dataType == Float {
			res *= v.value.(float64)
		} else {
			log.Fatal("Cannot * type ", dataTypes[v.dataType])
		}
	}
	var d dataType
	if math.Floor(res) == res {
		d.dataType = Int
		d.value = int(res)
	} else {
		d.dataType = Float
		d.value = res
	}
	return d
}

func Divide(ds *dataStore, params ...dataType) dataType {
	var res float64 = 0
	firstVal := params[0]
	if firstVal.dataType == Ident {
		firstVal = GetDsValue(ds, firstVal)
	}
	if firstVal.dataType == Int {
		res = float64(firstVal.value.(int))
	} else if firstVal.dataType == Float {
		res = firstVal.value.(float64)
	} else {
		log.Fatal("Cannot / type ", dataTypes[firstVal.dataType])
	}
	for _, v := range params[1:] {
		if v.dataType == Int {
			res /= float64(v.value.(int))
		} else if v.dataType == Float {
			res /= v.value.(float64)
		} else {
			log.Fatal("Cannot / type ", dataTypes[v.dataType])
		}
	}
	var d dataType
	if math.Floor(res) == res {
		d.dataType = Int
		d.value = int(res)
	} else {
		d.dataType = Float
		d.value = res
	}
	return d
}

func Exp(ds *dataStore, base dataType, exp dataType) dataType {
	var num1 float64 = 0
	if base.dataType == Ident {
		base = GetDsValue(ds, base)
	}
	if base.dataType == Float {
		num1 = base.value.(float64)
	} else if base.dataType == Int {
		num1 = float64(base.value.(int))
	} else {
		log.Fatal("Cannot ^ type ", dataTypes[base.dataType])
	}
	var num2 float64 = 0
	if exp.dataType == Ident {
		exp = GetDsValue(ds, exp)
	}
	if exp.dataType == Float {
		num2 = exp.value.(float64)
	} else if exp.dataType == Int {
		num2 = float64(exp.value.(int))
	} else {
		log.Fatal("Cannot ^ type ", dataTypes[exp.dataType])
	}
	res := math.Pow(num1, num2)
	var d dataType
	if math.Floor(res) == res {
		d.dataType = Int
		d.value = int(res)
	} else {
		d.dataType = Float
		d.value = res
	}
	return d
}

func Mod(ds *dataStore, num1 dataType, num2 dataType) dataType {
	val1 := 0
	if num1.dataType == Ident {
		num1 = GetDsValue(ds, num1)
	}
	if num1.dataType == Int {
		val1 = num1.value.(int)
	} else {
		log.Fatal("Cannot % type ", dataTypes[num1.dataType])
	}
	val2 := 0
	if num2.dataType == Ident {
		num2 = GetDsValue(ds, num2)
	}
	if num2.dataType == Float {
	} else if num2.dataType == Int {
		val2 = num2.value.(int)
	} else {
		log.Fatal("Cannot % type", dataTypes[num2.dataType])
	}
	return dataType{dataType: Int, value: val1 % val2}
}

func MakeVar(ds *dataStore, scopes int, name string, data dataType, isConst bool) {
	if scopes < len(ds.scopedVars) && StrArrIncludes(ds.scopedVars[scopes], name) {
		log.Fatal("Variable already initialized: ", name)
		return
	}

	if StrArrIncludes(reserved, name) {
		log.Fatal("Variable name \"", name, "\" is reserved")
		return
	}

	if len(name) == 0 {
		log.Fatal("Variable must have name")
		return
	}

	temp := ds.vars[name]
	if len(temp) > 0 && temp[len(temp)-1].isConst {
		log.Fatal("Variable is constant, unable to redefine value")
	}

	if data.dataType == List {
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, data, isConst))
	} else if data.dataType == Ident {
		val := GetDsValue(ds, data)
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, val, isConst))
	} else {
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, data, isConst))
	}
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

func SetVar(ds *dataStore, name string, data dataType) {
	if v, ok := ds.vars[name]; !ok {
		log.Fatal("Variable not initialized: ", name)
		return
	} else {
		if v[len(v)-1].isConst {
			log.Fatal("Variable is constant, unable to set value")
		}
	}
	if data.dataType == Ident {
		val := GetDsValue(ds, data)
		ds.vars[name][len(ds.vars[name])-1] = GetVariableFrom(name, val, false)
	} else {
		ds.vars[name][len(ds.vars[name])-1] = GetVariableFrom(name, data, false)
	}
}

func FreeVar(ds *dataStore, name string) {
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Unable to free, variable not initialized: ", name)
		return
	}
	delete(ds.vars, name)
}

func FreeFunc(ds *dataStore, name string) {
	if _, ok := ds.funcs[name]; !ok {
		log.Fatal("Unable to free, variable not initialized: " + name)
		return
	}
	delete(ds.funcs, name)
}

// gets value of token from ds
// h -> searches ds for var h, returns
// | -> searches ds for func h, returns
// ---> returns the input
func GetDsValue(ds *dataStore, val dataType) dataType {
	if val.dataType == Ident {
		if v, ok := ds.vars[val.value.(string)]; ok {
			return v[len(v)-1].data
		} else if v, ok := ds.funcs[val.value.(string)]; ok {
			f := v[len(v)-1]
			return dataType{dataType: Func, value: f}
		}
	}
	return val
}

func GetValueFromList(ds *dataStore, list dataType, index dataType) dataType {
	var arr []dataType
	if list.dataType == Ident {
		arr = GetDsValue(ds, list).value.([]dataType)
	} else {
		arr = list.value.([]dataType)
	}
	var indx int = 0
	if index.dataType == Int {
		indx = index.value.(int)
	} else {
		log.Fatal("Expected \"Int\" found ", dataTypes[index.dataType])
	}
	return arr[indx]
}

func LoopListIterator(ds *dataStore, scopes int, list dataType, iteratorName dataType, body dataType) (bool, []dataType) {
	arr := list
	if arr.dataType == Ident {
		arr = GetDsValue(ds, list)
	}
	made := false
	for _, v := range arr.value.([]dataType) {
		if !made {
			MakeVar(ds, scopes+1, iteratorName.value.(string), v, false)
		} else {
			SetVar(ds, iteratorName.value.(string), v)
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes, false)
		if hasReturn && len(val) > 0 && (val[0].dataType == BreakVals || val[0].dataType == ReturnVals) {
			if val[0].dataType == ReturnVals {
				return true, []dataType{val[0]}
			}
			break
		}
	}
	return false, []dataType{}
}

func LoopListIndexIterator(ds *dataStore, scopes int, list dataType, indexIterator dataType, iteratorName dataType, body dataType) (bool, []dataType) {
	arr := list
	if arr.dataType == Ident {
		arr = GetDsValue(ds, list)
	}
	made := false
	for i, v := range arr.value.([]dataType) {
		if !made {
			MakeVar(ds, scopes+1, iteratorName.value.(string), v, false)
			MakeVar(ds, scopes+1, indexIterator.value.(string), dataType{dataType: Int, value: i}, false)
		} else {
			SetVar(ds, iteratorName.value.(string), v)
			SetVar(ds, indexIterator.value.(string), dataType{dataType: Int, value: i})
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes, false)
		if hasReturn && len(val) > 0 && (val[0].dataType == BreakVals || val[0].dataType == ReturnVals) {
			if val[0].dataType == ReturnVals {
				return true, []dataType{val[0]}
			}
			break
		}
	}
	return false, []dataType{}
}

func LoopTo(ds *dataStore, scopes int, max dataType, indexIterator dataType, body dataType) (bool, []dataType) {
	maxNum := max.value.(int)
	made := false
	for i := 0; i < maxNum; i++ {
		if !made {
			MakeVar(ds, scopes+1, indexIterator.value.(string), dataType{dataType: Int, value: i}, false)
		} else {
			SetVar(ds, indexIterator.value.(string), dataType{dataType: Int, value: i})
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes, false)
		if hasReturn && len(val) > 0 && (val[0].dataType == BreakVals || val[0].dataType == ReturnVals) {
			if val[0].dataType == ReturnVals {
				return true, []dataType{val[0]}
			}
			break
		}
	}
	return false, []dataType{}
}

func LoopFromTo(ds *dataStore, scopes int, start dataType, max dataType, indexIterator dataType, body dataType) (bool, []dataType) {
	startNum := start.value.(int)
	maxNum := max.value.(int)
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
			MakeVar(ds, scopes+1, indexIterator.value.(string), dataType{dataType: Int, value: i}, false)
		} else {
			SetVar(ds, indexIterator.value.(string), dataType{dataType: Int, value: i})
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes, false)
		if hasReturn && len(val) > 0 && (val[0].dataType == BreakVals || val[0].dataType == ReturnVals) {
			if val[0].dataType == ReturnVals {
				return true, []dataType{val[0]}
			}
			break
		}
	}
	return false, []dataType{}
}

func Eq(ds *dataStore, params ...dataType) bool {
	eq := true
	for i := 0; i < len(params)-1; i++ {
		val1 := params[i]
		val2 := params[i+1]
		if val1.dataType == Ident {
			val1 = GetDsValue(ds, val1)
		}
		if val2.dataType == Ident {
			val2 = GetDsValue(ds, val2)
		}
		if val1.value != val2.value {
			eq = false
			break
		}
	}
	return eq
}

func If(ds *dataStore, scopes int, params ...dataType) (bool, []dataType) {
	hasReturn := true
	toReturn := []dataType{}
	info := params[0]
	if info.dataType == Ident {
		info = GetDsValue(ds, params[0])
	}
	if info.dataType == Bool {
		val := info.value.(bool)
		if val {
			hasReturn, toReturn = Eval(ds, params[1].value.([]token), scopes, false)
		} else if len(params) == 3 {
			hasReturn, toReturn = Eval(ds, params[2].value.([]token), scopes, false)
		}
	} else {
		log.Fatal("Error in \"if\", expected type: \"Bool\" found ", dataTypes[info.dataType])
	}
	return hasReturn, toReturn
}

func AppendToList(ds *dataStore, list []dataType, data ...dataType) []dataType {
	for _, v := range data {
		if v.dataType == Ident {
			list = append(list, GetDsValue(ds, v))
		} else {
			list = append(list, v)
		}
	}
	return list
}

func PrependToList(ds *dataStore, list []dataType, data ...dataType) []dataType {
	for _, v := range data {
		if v.dataType == Ident {
			list = append([]dataType{GetDsValue(ds, v)}, list...)
		} else {
			list = append([]dataType{v}, list...)
		}
	}
	return list
}

func ListFunc(ds *dataStore, f func(ds *dataStore, list []dataType, val ...dataType) []dataType, params ...dataType) dataType {
	isIdent := false
	name := ""
	list := params[0]
	if list.dataType == Ident {
		isIdent = true
		name = list.value.(string)
		list = GetDsValue(ds, list)
	}
	if list.dataType == List {
		val := list.value.([]dataType)
		val = f(ds, val, params[1:]...)
		list.value = val
		if isIdent {
			SetVar(ds, name, list)
		}
		return list
	} else {
		log.Fatal("Error in \"append\", expected type \"List\" found ", dataTypes[list.dataType])
	}
	return dataType{dataType: Nil, value: nil}
}

func Concat(ds *dataStore, params ...dataType) string {
	res := ""
	for _, v := range params {
		info := v
		if info.dataType == Ident {
			info = GetDsValue(ds, info)
		}
		if info.dataType == String {
			val := info.value.(string)
			res += val[1 : len(val)-1]
		} else {
			res += GetStrValue(info)
		}
	}
	return res
}

func Pop(ds *dataStore, list dataType) dataType {
	isIdent := false
	name := ""
	if list.dataType == Ident {
		isIdent = true
		name = list.value.(string)
		list = GetDsValue(ds, list)
	}
	if list.dataType != List {
		log.Fatal("Error in \"pop\" expected \"List\" found ", dataTypes[list.dataType])
	}
	items := list.value.([]dataType)
	if len(items) > 0 {
		val := items[len(items)-1]
		items = items[:len(items)-1]
		list.value = items
		if isIdent {
			SetVar(ds, name, list)
		}
		return val
	} else {
		return dataType{dataType: Nil, value: nil}
	}
}

func Remove(ds *dataStore, list dataType, index dataType) dataType {
	isIdent := false
	name := ""
	if list.dataType == Ident {
		isIdent = true
		name = list.value.(string)
		list = GetDsValue(ds, list)
	}
	listIndex := index.value.(int)
	if list.dataType != List {
		log.Fatal("Error in \"remove\" expected \"List\" found ", dataTypes[list.dataType])
	}
	items := list.value.([]dataType)
	if len(items) > 0 {
		item := items[listIndex]
		items = append(items[:listIndex], items[listIndex+1:]...)
		list.value = items
		if isIdent {
			SetVar(ds, name, list)
		}
		return item
	} else {
		return dataType{dataType: Nil, value: nil}
	}
}

func Len(ds *dataStore, list dataType) int {
	if list.dataType == Ident {
		list = GetDsValue(ds, list)
	}
	return len(list.value.([]dataType))
}

func And(ds *dataStore, params ...dataType) bool {
	for _, v := range params {
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
		}
		if v.dataType != Bool {
			log.Fatal("Error in \"and\", expected \"Bool\" found ", dataTypes[v.dataType])
		}
		if !v.value.(bool) {
			return false
		}
	}
	return true
}

func SetIndex(ds *dataStore, list dataType, index int, value dataType) (bool, dataType) {
	if value.dataType == Ident {
		value = GetDsValue(ds, value)
	}
	if list.dataType == Ident {
		name := list.value.(string)
		list = GetDsValue(ds, list)
		if index >= len(list.value.([]dataType)) {
			log.Fatal("Index out of bounds ", index, " on ", GetArrStr(list))
		}
		list.value.([]dataType)[index] = value
		SetVar(ds, name, list)
		return false, dataType{dataType: Nil, value: nil}
	} else {
		list = GetDsValue(ds, list)
		if index >= len(list.value.([]dataType)) {
			log.Fatal("Index out of bounds ", index, " on ", GetArrStr(list))
		}
		list.value.([]dataType)[index] = value
		return true, list
	}
}

func Or(ds *dataStore, params ...dataType) bool {
	for _, v := range params {
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
		}
		if v.dataType != Bool {
			log.Fatal("Error in \"or\", expected \"Bool\" found ", dataTypes[v.dataType])
		}
		if v.value.(bool) {
			return true
		}
	}
	return false
}

func Not(ds *dataStore, val dataType) bool {
	if val.dataType == Ident {
		val = GetDsValue(ds, val)
	}
	if val.dataType != Bool {
		log.Fatal("Error in \"not\", expected \"Bool\" found ", dataTypes[val.dataType])
	}
	return !val.value.(bool)
}

func MakeFunction(ds *dataStore, scopes int, name dataType, data []dataType) {
	if name.dataType != Ident {
		log.Fatal("Function named " + fmt.Sprint(name.value) + " must be a variable name")
		return
	}

	nameStr := name.value.(string)

	if scopes < len(ds.scopedFuncs) && StrArrIncludes(ds.scopedFuncs[scopes], nameStr) {
		log.Fatal("Function already initialized: " + nameStr)
		return
	}

	if StrArrIncludes(reserved, nameStr) {
		log.Fatal("Function name \"" + nameStr + "\" is reserved")
		return
	}

	f := function{name: nameStr, body: data[len(data)-1].value.([]token), params: data[0 : len(data)-1]}

	ds.funcs[nameStr] = append(ds.funcs[nameStr], f)
	for len(ds.scopedFuncs) < scopes {
		ds.scopedFuncs = append(ds.scopedFuncs, []string{})
	}
	for len(ds.scopedRedefFuncs) < scopes {
		ds.scopedRedefFuncs = append(ds.scopedRedefFuncs, []string{})
	}
	if _, ok := ds.funcs[nameStr]; ok {
		ds.scopedRedefFuncs[scopes-1] = append(ds.scopedRedefFuncs[scopes-1], nameStr)
	} else {
		ds.scopedFuncs[scopes-1] = append(ds.scopedFuncs[scopes-1], nameStr)
	}
}

func CallFunc(ds *dataStore, scopes int, name string, params []dataType) (bool, []dataType) {
	f := ds.funcs[name][len(ds.funcs[name])-1]
	for i, v := range f.params {
		MakeVar(ds, scopes, v.value.(string), GetDsValue(ds, params[i]), false)
	}
	hasReturn, toReturn := Eval(ds, f.body, scopes, false)
	ds.inFunc = false
	return hasReturn, toReturn
}

func GetAndCompareNumbers(ds *dataStore, val1 dataType, val2 dataType, f func(num1 float64, num2 float64) bool) bool {
	if val1.dataType == Ident {
		val1 = GetDsValue(ds, val1)
	}
	if val2.dataType == Ident {
		val2 = GetDsValue(ds, val2)
	}
	var num1 float64 = 0
	var num2 float64 = 0
	if val1.dataType == Int {
		num1 = float64(val1.value.(int))
	} else if val1.dataType == Float {
		num1 = val1.value.(float64)
	} else {
		log.Fatal("Expected \"In\" or \"Float\" found ", dataTypes[val1.dataType])
	}
	if val2.dataType == Int {
		num2 = float64(val2.value.(int))
	} else if val2.dataType == Float {
		num2 = val2.value.(float64)
	} else {
		log.Fatal("Expected \"In\" or \"Float\" found ", dataTypes[val2.dataType])
	}
	return f(num1, num2)
}

func LessThan(ds *dataStore, val1 dataType, val2 dataType) bool {
	comp := func(num1 float64, num2 float64) bool {
		return num1 < num2
	}
	return GetAndCompareNumbers(ds, val1, val2, comp)
}

func LessThanOrEqualTo(ds *dataStore, val1 dataType, val2 dataType) bool {
	comp := func(num1 float64, num2 float64) bool {
		return num1 <= num2
	}
	return GetAndCompareNumbers(ds, val1, val2, comp)
}
