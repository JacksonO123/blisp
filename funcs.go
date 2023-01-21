package main

import (
	"fmt"
	"log"
	"math"
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

var tokenTypes []string = []string{
	"Identifier",
	"OpenParen",
	"CloseParen",
	"OpenBracket",
	"CloseBracket",
	"StringToken",
	"BoolToken",
	"IntToken",
	"FloatToken",
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

func Print(ds *dataStore, params ...dataType) {
	for i, v := range params {
		if v.dataType == Ident {
			v = GetDsValue(ds, v)
			if v.dataType == Ident {
				log.Fatal("Unknown value: ", v.value.(string))
			}
		}
		if v.dataType == List {
			fmt.Print(GetDsValue(ds, GetArrStr(v)).value.(string))
		} else {
			fmt.Print(fmt.Sprint(v.value))
		}
		if i < len(params)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()
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
		if v.dataType == Int {
			res += float64(v.value.(int))
		} else if v.dataType == Float {
			res += v.value.(float64)
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
	if params[0].dataType == Int {
		res = float64(params[0].value.(int))
	} else if params[0].dataType == Float {
		res = params[0].value.(float64)
	}
	if len(params) == 1 {
		res *= -1
	} else {
		for _, v := range params[1:] {
			if v.dataType == Int {
				res -= float64(v.value.(int))
			} else if v.dataType == Float {
				res -= v.value.(float64)
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
	if params[0].dataType == Int {
		res = float64(params[0].value.(int))
	} else if params[0].dataType == Float {
		res = params[0].value.(float64)
	}
	for _, v := range params[1:] {
		if v.dataType == Int {
			res *= float64(v.value.(int))
		} else if v.dataType == Float {
			res *= v.value.(float64)
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
	if params[0].dataType == Int {
		res = float64(params[0].value.(int))
	} else if params[0].dataType == Float {
		res = params[0].value.(float64)
	}
	for _, v := range params[1:] {
		if v.dataType == Int {
			res /= float64(v.value.(int))
		} else if v.dataType == Float {
			res /= v.value.(float64)
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
	if base.dataType == Float {
		num1 = base.value.(float64)
	} else {
		num1 = float64(base.value.(int))
	}
	var num2 float64 = 0
	if base.dataType == Float {
		num2 = base.value.(float64)
	} else {
		num2 = float64(exp.value.(int))
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
	if num1.dataType == Float {
		log.Fatal("Unable to apply operator % on type \"Float\"")
	} else {
		val1 = num1.value.(int)
	}
	val2 := 0
	if num2.dataType == Float {
		log.Fatal("Unable to apply operator % on type \"Float\"")
	} else {
		val2 = num2.value.(int)
	}
	return dataType{dataType: Int, value: val1 % val2}
}

func MakeVar(ds *dataStore, scopes int, name string, data dataType) {
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

	if data.dataType == List {
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, data))
	} else if data.dataType == Ident {
		val := GetDsValue(ds, data)
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, val))
	} else {
		ds.vars[name] = append(ds.vars[name], GetVariableFrom(name, data))
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
	if _, ok := ds.vars[name]; !ok {
		log.Fatal("Variable not initialized: ", name)
		return
	}
	if data.dataType == Ident {
		val := GetDsValue(ds, data)
		ds.vars[name][len(ds.vars[name])-1] = GetVariableFrom(name, val)
	} else {
		ds.vars[name][len(ds.vars[name])-1] = GetVariableFrom(name, data)
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

func LoopListIterator(ds *dataStore, scopes int, list []token, iteratorName token, body []token) {
	// var arr []token
	// if len(list) == 1 {
	// 	arr = GetDsValue(ds, list[0])
	// } else {
	// 	arr = list
	// }
	// made := false
	// for _, v := range arr {
	// 	if !made {
	// 		MakeVar(ds, scopes+1, iteratorName, v)
	// 	} else {
	// 		SetVar(ds, iteratorName, v)
	// 	}
	// 	hasReturn, val := Eval(ds, body, scopes, false)
	// 	if hasReturn {
	// 		if val[0] == GetToken("break") {
	// 			break
	// 		}
	// 	}
	// }
}

func LoopListIndexIterator(ds *dataStore, scopes int, list []token, indexIterator token, iteratorName token, body []token) {
	// var arr []token
	// if len(list) == 1 {
	// 	arr = GetDsValue(ds, list[0])
	// } else {
	// 	arr = list
	// }
	// made := false
	// for i, v := range arr {
	// 	if !made {
	// 		MakeVar(ds, scopes+1, iteratorName, v)
	// 		MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
	// 	} else {
	// 		SetVar(ds, iteratorName, v)
	// 		SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
	// 	}
	// 	hasReturn, val := Eval(ds, body, scopes, false)
	// 	if hasReturn {
	// 		if val[0] == GetToken("break") {
	// 			break
	// 		}
	// 	}
	// }
}

func LoopTo(ds *dataStore, scopes int, max token, indexIterator token, body []token) {
	// maxNum := int(GetFloat64FromToken(max))
	// made := false
	// for i := 0; i < maxNum; i++ {
	// 	if !made {
	// 		MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
	// 	} else {
	// 		SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
	// 	}
	// 	hasReturn, val := Eval(ds, body, scopes, false)
	// 	if hasReturn {
	// 		if val[0] == GetToken("break") {
	// 			break
	// 		}
	// 	}
	// }
}

func LoopFromTo(ds *dataStore, scopes int, start token, max token, indexIterator token, body []token) {
	// startNum := int(GetFloat64FromToken(start))
	// maxNum := int(GetFloat64FromToken(max))
	// made := false
	// i := startNum
	// next := func() {
	// 	if startNum <= maxNum {
	// 		i++
	// 	} else {
	// 		i--
	// 	}
	// }
	// comp := func() bool {
	// 	if startNum <= maxNum {
	// 		return i < maxNum
	// 	} else {
	// 		return i > maxNum
	// 	}
	// }
	// for ; comp(); next() {
	// 	if !made {
	// 		MakeVar(ds, scopes+1, indexIterator, GetToken(fmt.Sprint(i)))
	// 	} else {
	// 		SetVar(ds, indexIterator, GetToken(fmt.Sprint(i)))
	// 	}
	// 	hasReturn, val := Eval(ds, body, scopes, false)
	// 	if hasReturn {
	// 		if val[0] == GetToken("break") {
	// 			break
	// 		}
	// 	}
	// }
}

func Eq(ds *dataStore, params ...token) bool {
	// eq := true
	// for i := 0; i < len(params)-1; i++ {
	// 	if GetDsValue(ds, params[i])[0].value != GetDsValue(ds, params[i+1])[0].value {
	// 		eq = false
	// 	}
	// }
	return false
}

func If(ds *dataStore, scopes int, params ...token) (bool, []token) {
	// hasReturn := true
	// toReturn := []token{}
	// info := GetDsValue(ds, params[0])
	// if info.variableType == Bool {
	// 	if val, err := strconv.ParseBool(info.value); err == nil && val {
	// 		hasReturn, toReturn = Eval(ds, Tokenize(params[1].value), scopes, false)
	// 	} else if len(params) == 3 {
	// 		hasReturn, toReturn = Eval(ds, Tokenize(params[2].value), scopes, false)
	// 	}
	// } else {
	// 	log.Fatal("Error in \"if\", expected type: \"Bool\" found ", info.variableType)
	// }
	// return hasReturn, toReturn
	return false, []token{}
}

func AppendToList(list token, toAppend token) token {
	// val := list.value
	// tokens := Tokenize(val[1 : len(val)-1])
	// res := append(tokens, toAppend)
	// return GetToken(JoinList(res))
	return token{}
}

func PrependToList(list token, toPrepend token) token {
	// val := list.value
	// tokens := Tokenize(val[1 : len(val)-1])
	// res := append([]token{toPrepend}, tokens...)
	// return GetToken(JoinList(res))
	return token{}
}

func ListFunc(ds *dataStore, f func(list token, val token) token, params ...token) token {
	// info := GetDsValue(ds, params[0])
	// if info.variableType == List {
	// 	list := GetToken(info.value)
	// 	for _, v := range params[1:] {
	// 		toAppend := GetDsValue(ds, v)
	// 		if toAppend.variableType == List {
	// 			parts := SplitList(toAppend.value)
	// 			for i := 0; i < len(parts); i++ {
	// 				list = f(list, parts[i])
	// 			}
	// 		} else {
	// 			list = f(list, GetToken(toAppend.value))
	// 		}
	// 	}
	// 	return list
	// } else {
	// 	log.Fatal("Error in \"append\", expected type \"List\" found ", info.variableType)
	// }
	// return GetToken("[]")
	return token{}
}

func Concat(ds *dataStore, params ...token) string {
	// res := ""
	// for _, v := range params {
	// 	info := GetDsValue(ds, v)
	// 	if info.variableType == String {
	// 		res += info.value[1 : len(info.value)-1]
	// 	} else {
	// 		res += info.value
	// 	}
	// }
	// return res
	return ""
}

func Pop(ds *dataStore, list token) token {
	// val := GetDsValue(ds, list)
	// if val.variableType != List {
	// 	log.Fatal("Error in \"pop\" expected \"List\" found ", val.variableType)
	// }
	// listItems := SplitList(val.value)
	// if len(listItems) > 0 {
	// 	lastItem := listItems[len(listItems)-1]
	// 	listItems = listItems[:len(listItems)-1]
	// 	if _, ok := ds.vars[list.value]; ok {
	// 		SetVar(ds, list, GetToken(JoinList(listItems)))
	// 	}
	// 	return lastItem
	// } else {
	// 	return GetToken("")
	// }
	return token{}
}

func Remove(ds *dataStore, list token, index token) token {
	// val := GetDsValue(ds, list)
	// listIndex := int(GetFloat64FromToken(ds, index))
	// if val.variableType != List {
	// 	log.Fatal("Error in \"remove\" expected \"List\" found ", val.variableType)
	// }
	// listItems := SplitList(val.value)
	// if len(listItems) > 0 {
	// 	item := listItems[listIndex]
	// 	listItems = append(listItems[:listIndex], listItems[listIndex+1:]...)
	// 	if _, ok := ds.vars[list.value]; ok {
	// 		SetVar(ds, list, GetToken(JoinList(listItems)))
	// 	}
	// 	return item
	// } else {
	// 	return GetToken("")
	// }
	return token{}
}

func Len(ds *dataStore, list token) int {
	// parts := SplitList(GetDsValue(ds, list).value)
	// return len(parts)
	return 0
}

func And(ds *dataStore, params ...token) bool {
	// res := false
	// for i, v := range params {
	// 	info := GetDsValue(ds, v)
	// 	if info.variableType != Bool {
	// 		log.Fatal("Error in \"and\", expected \"Bool\" found ", info.variableType)
	// 	}
	// 	val, err := strconv.ParseBool(info.value)
	// 	if err == nil {
	// 		if i == 0 {
	// 			res = val
	// 		} else {
	// 			if !res || !val {
	// 				res = false
	// 			}
	// 		}
	// 	}
	// }
	// return res
	return false
}

func SetIndex(ds *dataStore, list dataType, index int, value dataType) (bool, dataType) {
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

func Or(ds *dataStore, params ...token) bool {
	// res := false
	// for i, v := range params {
	// 	info := GetDsValue(ds, v)
	// 	if info.variableType != Bool {
	// 		log.Fatal("Error in \"or\", expected \"Bool\" found ", info.variableType)
	// 	}
	// 	val, err := strconv.ParseBool(info.value)
	// 	if i == 0 {
	// 		if err == nil {
	// 			res = val
	// 		}
	// 	} else {
	// 		res = res || val
	// 	}
	// }
	// return res
	return false
}

func Not(ds *dataStore, val token) bool {
	// info := GetDsValue(ds, val)
	// if info.variableType != Bool {
	// 	log.Fatal("Error in \"not\", expected \"Bool\" found ", info.variableType)
	// }
	// v, err := strconv.ParseBool(info.value)
	// if err == nil {
	// 	return !v
	// }
	// return true
	return false
}

func MakeFunction(ds *dataStore, scopes int, params ...token) {
	// make function var

	// if scopes < len(ds.scopedFuncs) && StrArrIncludes(ds.scopedFuncs[scopes], f.name) {
	// 	log.Fatal("Function already initialized: " + f.name)
	// 	return
	// }

	// if StrArrIncludes(reserved, f.name) {
	// 	log.Fatal("Function name \"" + f.name + "\" is reserved")
	// 	return
	// }

	// _, err := fastfloat.Parse(f.name)
	// if err == nil {
	// 	log.Fatal("Function named " + f.name + " cannot be a number")
	// 	return
	// }

	// ds.funcs[f.name] = append(ds.funcs[f.name], f)
	// for len(ds.scopedFuncs) < scopes {
	// 	ds.scopedFuncs = append(ds.scopedFuncs, []string{})
	// }
	// for len(ds.scopedRedefFuncs) < scopes {
	// 	ds.scopedRedefFuncs = append(ds.scopedRedefFuncs, []string{})
	// }
	// if _, ok := ds.funcs[f.name]; ok {
	// 	ds.scopedRedefFuncs[scopes-1] = append(ds.scopedRedefFuncs[scopes-1], f.name)
	// } else {
	// 	ds.scopedFuncs[scopes-1] = append(ds.scopedFuncs[scopes-1], f.name)
	// }
}

func CallFunc(ds *dataStore, scopes int, params ...token) (bool, []token) {
	// name := params[0]
	// inputs := params[1:]
	// f := ds.funcs[name.value][len(ds.funcs[name.value])-1]
	// funcParams := SplitList(f.params.value)
	// for i, v := range funcParams {
	// 	MakeVar(ds, scopes, v, GetToken(GetDsValue(ds, inputs[i]).value))
	// }
	// hasReturn, toReturn := Eval(ds, Tokenize(f.body.value), scopes, false)
	// ds.inFunc = false
	// return hasReturn, toReturn
	return false, []token{}
}
