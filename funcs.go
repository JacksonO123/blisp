package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func Print(ds *dataStore, params ...string) {
	res := []string{}
	for _, v := range params {
		if strings.Index(v, "\"") > -1 {
			res = append(res, v)
			continue
		}
		num, err := strconv.ParseFloat(v, 64)
		if err == nil {
			res = append(res, fmt.Sprint(num))
			continue
		} else {
			if val, ok := ds.vars[v]; ok {
				res = append(res, val.value)
			} else {
				log.Fatal("unknown value: " + v)
			}
		}
	}
	fmt.Println(strings.Join(res, ", "))
}

func GetFloat64FromStrings(ds *dataStore, strs ...string) []float64 {
	var nums []float64
	for _, v := range strs {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			if val, ok := ds.vars[v]; ok {
				if val.variableType == Int || val.variableType == Double {
					n, _ = strconv.ParseFloat(val.value, 64)
				} else {
					log.Fatal(err)
				}
			}
		}
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

func MakeVar(name string, val string, ds *dataStore) {
	reserved := []string{"print", "+", "-", "*", "/", "eval", "var"}
	nameChars := strings.Split(name, "")
	if StrArrIncludes(reserved, val) {
		log.Fatal("Variable name " + val + " is reserved")
		return
	}

	if len(val) == 0 {
		log.Fatal("Variable must have name")
		return
	}

	_, err := strconv.ParseFloat(nameChars[0], 64)
	if err == nil {
		log.Fatal("Variable name cannot start with a number")
		return
	}

	var variable variable
	variable.name = name
	if strings.Index(val, "\"") > 0 {
		variable.variableType = String
	} else if Eq(val, "true") || Eq(val, "false") {
		variable.variableType = Bool
	} else {
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			numStr := fmt.Sprint(num)
			if strings.Index(numStr, ".") > 0 {
				variable.variableType = Double
			} else {
				variable.variableType = Int
			}
		}
	}
	variable.value = val
	ds.vars[name] = variable
}
