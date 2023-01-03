package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Eq(s1 string, s2 string) bool {
	return strings.Compare(s1, s2) == 0
}

func StringArrMap(arr []string, f func(string) string) []string {
	newArr := arr
	for i := range arr {
		newArr[i] = f(newArr[i])
	}
	return newArr
}

func GetBlocks(code string) []string {
	temp := ""
	var blocks []string
	inString := false
	parenScopes := 0
	chars := strings.Split(code, "")
	for _, v := range chars {
		if Eq(v, "\"") {
			inString = !inString
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
			blocks = append(blocks, temp)
			temp = ""
		}
	}
	return blocks
}

func Flatten(block string) string {
	res := block[1 : len(block)-1]
	inString := false
	var starts []int
	chars := strings.Split(res, "")
	for i := 0; i < len(res); i++ {
		if Eq(chars[i], "\"") {
			inString = !inString
		}
		if !inString {
			if Eq(chars[i], "(") {
				starts = append(starts, i)
			} else if Eq(chars[i], ")") {
				slice := res[starts[len(starts)-1] : i+1]
				hasReturn, val := Eval(slice)
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
	for _, v := range strings.Split(str, "") {
		if Eq(v, "\"") {
			inString = !inString
		}
		if !inString {
			if Eq(v, " ") {
				res = append(res, temp)
				temp = ""
			} else {
				temp += v
			}
		} else {
			temp += v
		}
	}
	res = append(res, temp)
	return res
}

func Eval(code string) (bool, string) {
	blocks := GetBlocks(code)
	hasReturn := false
	toReturn := ""
	if len(blocks) != 1 {
		for i, block := range blocks {
			_, blocks[i] = Eval(block)
		}
	} else {
		flatBlock := Flatten(blocks[0])
		parts := SplitParams(flatBlock)
		params := parts[1:]
		switch parts[0] {
		case "print":
			{
				hasReturn = true
				toReturn = "\"(printing " + strings.Join(params, ", ") + ")\""
				Print(parts[1:]...)
			}
		case "+":
			{
				hasReturn = true
				toReturn = fmt.Sprint(Add(params...))
			}
		case "-":
			{
				hasReturn = true
				toReturn = fmt.Sprint(Sub(params...))
			}
		case "*":
			{
				hasReturn = true
				toReturn = fmt.Sprint(Mult(params...))
			}
		case "/":
			{
				hasReturn = true
				toReturn = fmt.Sprint(Divide(params...))
			}
		default:
			{
				fmt.Println("default", parts)
			}
		}
	}
	return hasReturn, toReturn
}

func main() {
	args := os.Args[1:]
	var fileName string
	if len(args) > 0 {
		fileName = args[0]
		if strings.Index(fileName, ".blisp") < 0 {
			fileName += ".blisp"
		}
	} else {
		fileName = "main.blisp"
	}
	dat, err := os.ReadFile(fileName)
	check(err)
	Eval(string(dat))
}
