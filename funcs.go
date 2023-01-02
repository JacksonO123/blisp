package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func Print(params ...string) {
	fmt.Println(strings.Join(params, ", "))
}

func GetFloat64FromStrings(strs ...string) []float64 {
	var nums []float64
	for _, v := range strs {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Fatal(err)
		}
		nums = append(nums, n)
	}
	return nums
}

func Add(params ...string) float64 {
	nums := GetFloat64FromStrings(params...)
	var res float64 = 0
	for _, v := range nums {
		res += v
	}
	return res
}

func Sub(params ...string) float64 {
	nums := GetFloat64FromStrings(params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res -= v
	}
	return res
}

func Mult(params ...string) float64 {
	nums := GetFloat64FromStrings(params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res *= v
	}
	return res
}

func Divide(params ...string) float64 {
	nums := GetFloat64FromStrings(params...)
	var res float64 = nums[0]
	for _, v := range nums[1:] {
		res /= v
	}
	return res
}
