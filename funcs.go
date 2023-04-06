package main

import (
	"fmt"
	"log"
	"math"
	"os"
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
	"parse",
	"<",
	">",
	"<=",
	">=",
	"read",
	"write",
	"substr",
	"struct",
	"shift",
	".",
}

func GetArr(tokens []token) (dataType, int) {
	res := make([]dataType, 100)
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

func PrintStruct(ds *dataStore, val dataType) {
	if val.dataType == Struct {
		s := val.value.([]structAttr)
		fmt.Println("{")
		for _, attr := range s {
			fmt.Print("\t" + fmt.Sprint(attr.name) + ": ")
			Print(ds, *attr.attr)
		}
		fmt.Println("}")
	} else {
		log.Fatal("Unable to PrintStruct for type ", dataTypes[val.dataType])
	}
}

func Print(ds *dataStore, params ...dataType) {
	for i, v := range params {
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
			if v.dataType == Ident {
				log.Fatal("Unknown value: ", v.value)
			}
		}
		if v.dataType == List {
			PrintArr(v)
		} else if v.dataType == Struct {
			PrintStruct(ds, v)
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
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
		}
		if v.dataType == Int {
			res += float64(v.value.(int))
		} else if v.dataType == Float {
			res += v.value.(float64)
		} else {
			log.Fatal("Cannot + type ", dataTypes[v.dataType])
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
			if v.dataType == Ident {
				v = GetDsValue(ds, v)
			}
			if v.dataType == Int {
				res -= float64(v.value.(int))
			} else if v.dataType == Float {
				res -= v.value.(float64)
			} else {
				log.Fatal("Cannot - type ", dataTypes[v.dataType])
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
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
		}
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
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
		}
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
	if num2.dataType == Int {
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

	if name == "_" {
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
	if name == "_" {
		return
	}
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

func GetFromValue(ds *dataStore, val dataType, index dataType) dataType {
	if val.dataType == Ident {
		val = GetDsValue(ds, val)
	}
	if val.dataType != String && val.dataType != List && val.dataType != Struct {
		log.Fatal("Error in \"get\", expected \"String\", \"List\", or \"Struct\" found ", dataTypes[val.dataType])
	}

	if val.dataType != Struct {
		if index.dataType == Ident {
			index = GetDsValue(ds, index)
		}
		if index.dataType != Int {
			log.Fatal("Error in \"get\", expected \"Int\" found ", dataTypes[index.dataType])
		}
	} else {
		if index.dataType != Ident {
			log.Fatal("Unable to index \"Struct\" with type ", dataTypes[index.dataType])
		}
	}

	if val.dataType == String {
		parts := strings.Split(val.value.(string), "")
		var d dataType
		d.dataType = String
		d.value = parts[index.value.(int)]
		return d
	} else if val.dataType == List {
		parts := val.value.([]dataType)
		return parts[index.value.(int)]
	} else if val.dataType == Struct {
		parts := val.value.([]structAttr)
		for i := 0; i < len(parts); i++ {
			if parts[i].name == index.value.(string) {
				return *parts[i].attr
			}
		}
		return dataType{dataType: Nil, value: nil}
	}
	return dataType{dataType: Nil, value: nil}
}

func LoopListIterator(ds *dataStore, scopes int, list dataType, iteratorName dataType, body dataType) (bool, []dataType) {
	if list.dataType == Ident {
		list = GetDsValue(ds, list)
	}
	if list.dataType != List {
		log.Fatal("Error in \"loop\" expected \"List\" found ", dataTypes[list.dataType])
	}
	if iteratorName.dataType != Ident {
		log.Fatal("Error in \"loop\" expected \"Ident\" found ", dataTypes[list.dataType])
	}
	made := false
	for _, v := range list.value.([]dataType) {
		if !made {
			MakeVar(ds, scopes+1, iteratorName.value.(string), v, false)
		} else {
			SetVar(ds, iteratorName.value.(string), v)
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes)
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
		hasReturn, val := Eval(ds, body.value.([]token), scopes)
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
	if max.dataType == Ident {
		max = GetDsValue(ds, max)
	}
	var maxNum int
	if max.dataType == Int {
		maxNum = max.value.(int)
	} else {
		log.Fatal("Error in \"loop\", expected \"Int\" found ", dataTypes[max.dataType])
	}
	if indexIterator.dataType != Ident {
		log.Fatal("Error in \"loop\" expected \"Ident\" found ", dataTypes[indexIterator.dataType])
	}
	made := false
	for i := 0; i < maxNum; i++ {
		if !made {
			MakeVar(ds, scopes+1, indexIterator.value.(string), dataType{dataType: Int, value: i}, false)
		} else {
			SetVar(ds, indexIterator.value.(string), dataType{dataType: Int, value: i})
		}
		made = true
		hasReturn, val := Eval(ds, body.value.([]token), scopes)
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
	if start.dataType == Ident {
		start = GetDsValue(ds, start)
	}
	if max.dataType == Ident {
		max = GetDsValue(ds, max)
	}
	if start.dataType != Int {
		log.Fatal("Error in \"loop\", expected \"Int\" found ", dataTypes[start.dataType])
	}
	if max.dataType != Int {
		log.Fatal("Error in \"loop\", expected \"Int\" found ", dataTypes[max.dataType])
	}
	if indexIterator.dataType != Ident {
		log.Fatal("Error in \"loop\" expected \"Ident\" found ", dataTypes[indexIterator.dataType])
	}
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
		hasReturn, val := Eval(ds, body.value.([]token), scopes)
		if hasReturn && len(val) > 0 && (val[0].dataType == BreakVals || val[0].dataType == ReturnVals) {
			if val[0].dataType == ReturnVals {
				return true, []dataType{val[0]}
			}
			break
		}
	}
	return false, []dataType{}
}

func CompareLists(ds *dataStore, val1 dataType, val2 dataType) bool {
	list1 := val1.value.([]dataType)
	list2 := val2.value.([]dataType)
	if len(list1) != len(list2) {
		return false
	}
	for i := 0; i < len(list1); i++ {
		if !Eq(ds, list1[i], list2[i]) {
			return false
		}
	}
	return true
}

func CompareStructs(ds *dataStore, val1 dataType, val2 dataType) bool {
	s1 := val1.value.([]structAttr)
	s2 := val1.value.([]structAttr)
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		found := false
		for j := 0; j < len(s2); j++ {
			if s1[i].name == s2[j].name &&
				s1[i].attr.dataType == s2[j].attr.dataType &&
				Eq(ds, *s1[i].attr, *s2[j].attr) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func Eq(ds *dataStore, params ...dataType) bool {
	for i := 0; i < len(params)-1; i++ {
		val1 := params[i]
		val2 := params[i+1]
		if val1.dataType == Ident {
			val1 = GetDsValue(ds, val1)
		}
		if val2.dataType == Ident {
			val2 = GetDsValue(ds, val2)
		}
		if val1.dataType == Ident {
			log.Fatal("Cannot compare unknown value ", val1.value)
		}
		if val2.dataType == Ident {
			log.Fatal("Cannot compare unknown value ", val2.value)
		}
		if val1.dataType == List {
			if val2.dataType == List {
				if !CompareLists(ds, val1, val2) {
					return false
				}
			} else {
				return false
			}
		} else if val2.dataType == List {
			log.Fatal("Cannot compare types \"List\" and ", dataTypes[val1.dataType])
		} else if val1.dataType == Struct {
			if val2.dataType == Struct {
				if !CompareStructs(ds, val1, val2) {
					return false
				}
			} else {
				return false
			}
		} else if val2.dataType == Struct {
			return false
		} else if val1.value != val2.value {
			return false
		}
	}
	return true
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
			hasReturn, toReturn = Eval(ds, params[1].value.([]token), scopes)
		} else if len(params) == 3 {
			hasReturn, toReturn = Eval(ds, params[2].value.([]token), scopes)
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

// Used for mutating lists that return just the resulting list
// if list is from a variable, it updates the variable
// returns teh new list
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
		log.Fatal("Error mutating list, expected type \"List\" found ", dataTypes[list.dataType])
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
			res += info.value.(string)
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
		return dataType{dataType: List, value: []dataType{}}
	}
}

func Remove(ds *dataStore, val dataType, index dataType) dataType {
	isIdent := false
	name := ""
	if val.dataType == Ident {
		isIdent = true
		name = val.value.(string)
		val = GetDsValue(ds, val)
	}
	if val.dataType == List {
		listIndex := 0
		var listIndexData dataType
		if index.dataType == Ident {
			listIndexData = GetDsValue(ds, index)
		}
		if listIndexData.dataType == Int {
			listIndex = listIndexData.value.(int)
		} else {
			log.Fatal("Error in \"remove\" expected \"Int\" found ", dataTypes[val.dataType])
		}
		items := val.value.([]dataType)
		if len(items) > 0 {
			item := items[listIndex]
			items = append(items[:listIndex], items[listIndex+1:]...)
			val.value = items
			if isIdent {
				SetVar(ds, name, val)
			}
			return item
		} else {
			return dataType{dataType: Nil, value: nil}
		}
	} else if val.dataType == Struct {
		strct := val.value.([]structAttr)
		if index.dataType != Ident {
			log.Fatal("Error in \"remove\" expected \"Ident\" found ", dataTypes[val.dataType])
		}
		item := dataType{dataType: Nil, value: nil}
		for i := 0; i < len(strct); i++ {
			if strct[i].name == index.value.(string) {
				item = *strct[i].attr
				strct = append(strct[:i], strct[i+1:]...)
				val.value = strct
				if isIdent {
					SetVar(ds, name, val)
				}
				break
			}
		}
		return item
	} else {
		log.Fatal("Error in \"remove\" expected \"List\" or \"Struct\" found ", dataTypes[val.dataType])
	}
	return dataType{value: nil, dataType: Nil}
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

func SetValue(ds *dataStore, val dataType, index dataType, value dataType) {
	if value.dataType == Ident {
		value = GetDsValue(ds, value)
	}

	name := ""
	setVar := false

	if val.dataType == Ident {
		name = val.value.(string)
		setVar = true
	}

	if val.dataType == Ident {
		val = GetDsValue(ds, val)
	}
	if val.dataType != Struct && val.dataType != List {
		log.Fatal("Error in \"set\", expected \"Struct\" or \"List\" found ", dataTypes[val.dataType])
	}

	if val.dataType == Struct {
		if index.dataType != Ident {
			log.Fatal("Error in \"set\", expected \"Ident\" found ", dataTypes[index.dataType])
		}

		s := val.value.([]structAttr)
		for i := 0; i < len(s); i++ {
			if s[i].name == index.value.(string) {
				s[i].attr = &value
			}
		}

		val.value = s
	} else if val.dataType == List {
		list := val.value.([]dataType)
		if index.dataType == Ident {
			index = GetDsValue(ds, index)
		}
		if index.dataType != Int {
			log.Fatal("Error in \"set\", expected \"Int\" found ", dataTypes[index.dataType])
		}
		idx := index.value.(int)
		if idx >= len(list) {
			log.Fatal("Index out of bounds ", index, " on list of length ", len(list))
		}
		val.value.([]dataType)[idx] = value
	} else {
		log.Fatal("Error in \"set\", expected \"List\" or \"Struct\" found ", dataTypes[val.dataType])
	}

	if setVar {
		SetVar(ds, name, val)
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

func MakeFunction(ds *dataStore, scopes int, name dataType, data []dataType) (bool, *function) {
	if name.dataType != Ident {
		log.Fatal("Function named " + fmt.Sprint(name.value) + " must be an Ident")
		return false, nil
	}

	nameStr := name.value.(string)
	save := true
	if nameStr == "_" {
		save = false
	}

	if scopes < len(ds.scopedFuncs) && StrArrIncludes(ds.scopedFuncs[scopes], nameStr) {
		log.Fatal("Function already initialized: " + nameStr)
		return false, nil
	}

	if StrArrIncludes(reserved, nameStr) {
		log.Fatal("Function name \"" + nameStr + "\" is reserved")
		return false, nil
	}

	f := function{name: nameStr, body: data[len(data)-1].value.([]token), params: data[0 : len(data)-1]}

	if save {
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
		return false, nil
	}
	return true, &f
}

func CallFunc(ds *dataStore, scopes int, name dataType, params []dataType) (bool, []dataType) {
	highestVarIndex := 0
	varWithName := ds.vars[name.value.(string)]
	funcWithName := ds.funcs[name.value.(string)]
	ds.inFunc = true
	var f function
	for i := len(varWithName); i > 0; i-- {
		v := varWithName[i-1]
		if v.data.dataType == Func {
			highestVarIndex = i - 1
			break
		}
	}
	if highestVarIndex > int(math.Max(float64(len(funcWithName)-1), 0)) {
		f = varWithName[highestVarIndex].data.value.(function)
	} else {
		if len(funcWithName) == 0 {
			if name.dataType == Ident {
				newName := GetDsValue(ds, name)
				if newName.dataType == Func {
					return CallInlineFunc(ds, scopes, name.value.(string), newName.value.(function), params)
				} else if newName.dataType != String && newName.dataType != Ident {
					log.Fatal("Unknown function: \"", newName.value, "\"")
				}
				newName.dataType = Ident
				if f, ok := ds.builtins[newName.value.(string)]; ok {
					return f(ds, scopes, params)
				} else {
					return CallFunc(ds, scopes, newName, params)
				}
			}
			log.Fatal("Unknown function: \"", name, "\"")
		}
		f = funcWithName[len(funcWithName)-1]
	}

	return CallInlineFunc(ds, scopes, name.value.(string), f, params)
}

func CallInlineFunc(ds *dataStore, scopes int, name string, f function, params []dataType) (bool, []dataType) {
	if len(f.params) != len(params) {
		log.Fatal("Error in \"", name, "\", expected ", len(f.params), " params found ", len(params))
	}

	for i := 0; i < len(f.params); i++ {
		MakeVar(ds, scopes+1, f.params[i].value.(string), GetDsValue(ds, params[i]), false)
	}
	hasReturn, toReturn := Eval(ds, f.body, scopes)
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
		log.Fatal("Expected \"Int\" or \"Float\" found ", dataTypes[val1.dataType])
	}
	if val2.dataType == Int {
		num2 = float64(val2.value.(int))
	} else if val2.dataType == Float {
		num2 = val2.value.(float64)
	} else {
		log.Fatal("Expected \"Int\" or \"Float\" found ", dataTypes[val2.dataType])
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

func GetFile(ds *dataStore, file dataType) string {
	if file.dataType == Ident {
		file = GetDsValue(ds, file)
	}
	if file.dataType != String {
		log.Fatal("Error in \"read\", expected \"String\" found ", dataTypes[file.dataType])
	}
	val, err := os.ReadFile(file.value.(string))
	if err == nil {
		return string(val)
	} else {
		log.Fatal(err)
	}
	return ""
}

func WriteFile(ds *dataStore, file dataType, data dataType) {
	if file.dataType == Ident {
		file = GetDsValue(ds, file)
	}
	if data.dataType == Ident {
		data = GetDsValue(ds, data)
	}
	if file.dataType != String {
		log.Fatal("Error in \"read\", expected \"String\" found ", dataTypes[file.dataType])
	}
	if data.dataType != String {
		log.Fatal("Error in \"read\", expected \"String\" found ", dataTypes[file.dataType])
	}
	fileStr := file.value.(string)
	fileStr = fileStr[1 : len(fileStr)-1]
	dataStr := data.value.(string)
	dataStr = dataStr[1 : len(dataStr)-1]
	err := os.WriteFile(fileStr, []byte(dataStr), 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func SubstrEnd(ds *dataStore, str dataType, index dataType) string {
	if str.dataType == Ident {
		str = GetDsValue(ds, str)
	}
	if index.dataType == Ident {
		index = GetDsValue(ds, index)
	}
	if str.dataType != String {
		log.Fatal("Error in \"substr\", expected \"String\" found ", dataTypes[str.dataType])
	}
	if index.dataType != Int {
		log.Fatal("Error in \"substr\", expected \"Int\" found ", dataTypes[index.dataType])
	}
	return str.value.(string)[index.value.(int):]
}

func Substr(ds *dataStore, str dataType, startIndex dataType, endIndex dataType) string {
	if str.dataType == Ident {
		str = GetDsValue(ds, str)
	}
	if startIndex.dataType == Ident {
		startIndex = GetDsValue(ds, startIndex)
	}
	if endIndex.dataType == Ident {
		endIndex = GetDsValue(ds, endIndex)
	}
	if str.dataType != String {
		log.Fatal("Error in \"substr\", expected \"String\" found ", dataTypes[str.dataType])
	}
	if startIndex.dataType != Int {
		log.Fatal("Error in \"substr\", expected \"Int\" found ", dataTypes[startIndex.dataType])
	}
	if endIndex.dataType != Int {
		log.Fatal("Error in \"substr\", expected \"Int\" found ", dataTypes[endIndex.dataType])
	}
	return str.value.(string)[startIndex.value.(int):endIndex.value.(int)]
}

func GetType(ds *dataStore, val dataType) string {
	isIdent := false
	if val.dataType == Ident {
		isIdent = true
		val = GetDsValue(ds, val)
	}

	str := ""
	if isIdent {
		str += "Identifier: "
	}
	str += dataTypes[val.dataType]
	return str
}

func MakeStruct(ds *dataStore, params ...dataType) dataType {
	var d dataType
	d.dataType = Struct

	m := make([]structAttr, 20)
	key := dataType{dataType: Nil, value: nil}
	for i, v := range params {
		if i%2 == 0 {
			if v.dataType != Ident {
				log.Fatal("Expected \"Ident\" found ", dataTypes[v.dataType])
			}
			key = v
		} else {
			info := v
			if info.dataType == Ident {
				info = GetDsValue(ds, info)
				if info.dataType == Ident {
					log.Fatal("Unknown value: ", info.value.(string))
				}
			}
			if key.dataType == Ident {
				exists := false
				for i := 0; i < len(m); i++ {
					if m[i].name == key.value.(string) {
						exists = true
					}
				}
				if !exists {
					m = append(m, structAttr{name: key.value.(string), attr: &info})
				}
			}
		}
	}

	d.value = m
	return d
}

func Parse(ds *dataStore, str dataType) dataType {
	if str.dataType == Ident {
		str = GetDsValue(ds, str)
	}
	if str.dataType == String {
		val, err := fastfloat.Parse(str.value.(string))
		if err == nil {
			if math.Floor(val) == val {
				return dataType{
					value:    int(val),
					dataType: Int,
				}
			} else {
				return dataType{
					value:    val,
					dataType: Float,
				}
			}
		}
	}
	return dataType{value: nil, dataType: Nil}
}

func Shift(ds *dataStore, arr dataType) dataType {
	isIdent := false
	name := ""
	if arr.dataType == Ident {
		isIdent = true
		name = arr.value.(string)
		arr = GetDsValue(ds, arr)
	}
	if arr.dataType != List {
		log.Fatal("Error in \"shift\", expected \"List\" found ", dataTypes[arr.dataType])
	}
	list := arr.value.([]dataType)
	if len(list) == 0 {
		return dataType{dataType: List, value: []dataType{}}
	}
	val := list[0]
	list = list[1:]
	dt := dataType{
		dataType: List,
		value:    list,
	}
	if isIdent {
		SetVar(ds, name, dt)
	}
	return val
}

func CallProp(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
	obj := params[0]
	if obj.dataType == Ident {
		obj = GetDsValue(ds, obj)
	}
	if obj.dataType != Struct {
		log.Fatal("Error in \".\", expected \"Struct\" found ", dataTypes[obj.dataType])
	}
	key := params[1]
	if key.dataType != Ident {
		log.Fatal("Error in \".\", expected \"Ident\" found ", dataTypes[obj.dataType])
	}

	structAttrs := obj.value.([]structAttr)
	index := 0
	for i := 0; i < len(structAttrs); i++ {
		if structAttrs[i].name == key.value.(string) {
			index = i
			break
		}
	}
	fn := structAttrs[index].attr
	f := fn.value.(function)

	if len(f.params) != len(params)-1 {
		log.Fatal("Error in \".\", expected ", len(f.params)-1, " params found ", len(params)-2)
	}

	MakeVar(ds, scopes+1, f.params[0].value.(string), GetDsValue(ds, params[0]), false)
	for i := 2; i < len(params); i++ {
		MakeVar(ds, scopes+1, f.params[i-1].value.(string), GetDsValue(ds, params[i]), false)
	}
	ds.inFunc = false
	return Eval(ds, f.body, scopes+1)
}

func WhileLoop(ds *dataStore, scopes int, params []dataType) (bool, []dataType) {
	conditionTokens := params[0].value.([]token)

	hasReturn := false
	res := []dataType{{dataType: Nil, value: nil}}

	if conditionTokens[0].tokenType == Identifier {
		dt := dataType{dataType: Ident, value: conditionTokens[0].value.(string)}
		condition := GetDsValue(ds, dt)
		if condition.dataType != Bool {
			log.Fatal("Error in \"while\", expected \"Bool\" found ", dataTypes[dt.dataType])
		}
		for condition.value.(bool) {
			hasReturn, res = Eval(ds, params[1].value.([]token), scopes+1)
			if hasReturn && len(res) > 0 && (res[0].dataType == BreakVals || res[0].dataType == ReturnVals) {
				if res[0].dataType == ReturnVals {
					return true, []dataType{res[0]}
				}
				break
			}
			condition = GetDsValue(ds, dt)
		}
	} else if conditionTokens[0].tokenType == BoolToken {
		if conditionTokens[0].value.(bool) {
			for {
				hasReturn, res = Eval(ds, params[1].value.([]token), scopes+1)
				if hasReturn && len(res) > 0 && (res[0].dataType == BreakVals || res[0].dataType == ReturnVals) {
					if res[0].dataType == ReturnVals {
						return true, []dataType{res[0]}
					}
					break
				}
			}
		}
	} else if conditionTokens[0].tokenType == OpenParen {
		hasCondition, condition := Eval(ds, conditionTokens, scopes+1)
		if len(condition) != 1 {
			log.Fatal("Error in \"while\", expected one return from condition, found ", len(condition))
		}
		if condition[0].dataType != Bool {
			log.Fatal("Error in \"while\", expected \"Bool\" condition, found ", dataTypes[condition[0].dataType])
		}
		for hasCondition && condition[0].value.(bool) {
			hasReturn, res = Eval(ds, params[1].value.([]token), scopes+1)
			hasCondition, condition = Eval(ds, conditionTokens, scopes+1)
		}
		if !hasCondition {
			log.Fatal("Error in \"while\", expected return from condition, found none")
		}
	} else {
		log.Fatal("Unexpected token ", conditionTokens[0].value)
	}
	return hasReturn, res
}

func AddOne(ds *dataStore, num dataType) dataType {
	isIdent := false
	name := ""
	if num.dataType == Ident {
		name = num.value.(string)
		isIdent = true
		num = GetDsValue(ds, num)
	}
	if num.dataType == Int {
		num.value = num.value.(int) + 1
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else if num.dataType == Float {
		num.value = num.value.(float64) + 1
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else {
		log.Fatal("Error in \"++\", expected \"Int\" or \"Float\"")
	}
	return dataType{dataType: Nil, value: nil}
}

func AddMany(ds *dataStore, num dataType, amount dataType) dataType {
	isIdent := false
	name := ""
	if num.dataType == Ident {
		name = num.value.(string)
		isIdent = true
		num = GetDsValue(ds, num)
	}

	if amount.dataType == Ident {
		amount = GetDsValue(ds, amount)
	}

	var amountVal float64
	if amount.dataType == Int {
		amountVal = float64(amount.value.(int))
	} else if amount.dataType == Float {
		amountVal = amount.value.(float64)
	} else {
		log.Fatal("Error in \"+=\", expected \"Int\" or \"Float\"")
	}

	if num.dataType == Int {
		num.value = num.value.(int) + int(amountVal)
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else if num.dataType == Float {
		num.value = num.value.(float64) + amountVal
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else {
		log.Fatal("Error in \"+=\", expected \"Int\" or \"Float\"")
	}
	return dataType{dataType: Nil, value: nil}
}

func SubOne(ds *dataStore, num dataType) dataType {
	isIdent := false
	name := ""
	if num.dataType == Ident {
		name = num.value.(string)
		isIdent = true
		num = GetDsValue(ds, num)
	}
	if num.dataType == Int {
		num.value = num.value.(int) - 1
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else if num.dataType == Float {
		num.value = num.value.(float64) - 1
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else {
		log.Fatal("Error in \"++\", expected \"Int\" or \"Float\"")
	}
	return dataType{dataType: Nil, value: nil}
}

func SubMany(ds *dataStore, num dataType, amount dataType) dataType {
	isIdent := false
	name := ""
	if num.dataType == Ident {
		name = num.value.(string)
		isIdent = true
		num = GetDsValue(ds, num)
	}

	if amount.dataType == Ident {
		amount = GetDsValue(ds, amount)
	}

	var amountVal float64
	if amount.dataType == Int {
		amountVal = float64(amount.value.(int))
	} else if amount.dataType == Float {
		amountVal = amount.value.(float64)
	} else {
		log.Fatal("Error in \"+=\", expected \"Int\" or \"Float\"")
	}

	if num.dataType == Int {
		num.value = num.value.(int) - int(amountVal)
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else if num.dataType == Float {
		num.value = num.value.(float64) - amountVal
		if isIdent {
			SetVar(ds, name, num)
		}
		return num
	} else {
		log.Fatal("Error in \"+=\", expected \"Int\" or \"Float\"")
	}
	return dataType{dataType: Nil, value: nil}
}

