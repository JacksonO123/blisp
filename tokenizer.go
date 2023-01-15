package main

import "fmt"

type TokenType string

const (
	Identifier  TokenType = "Identifier"
	OpenParen   TokenType = "OpenParen"
	CloseParen  TokenType = "CloseParen"
	ListToken   TokenType = "ListToken"
	StringToken TokenType = "StringToken"
)

type token struct {
	tokenType TokenType
	value     string
}

func InitTokenMap(tm map[rune]func(*[]token, string, *int)) {
	tm['('] = func(res *[]token, code string, i *int) {
		var t token
		t.tokenType = OpenParen
		t.value = "("
		*res = append(*res, t)
	}
	tm[')'] = func(res *[]token, code string, i *int) {
		var t token
		t.tokenType = CloseParen
		t.value = ")"
		*res = append(*res, t)
	}
	tm['['] = func(res *[]token, code string, i *int) {
		var t token
		t.tokenType = ListToken
		arr, index := GetArr(code)
		*i += index - 1
		t.value = arr
		*res = append(*res, t)
	}
	tm['"'] = func(res *[]token, code string, i *int) {
		var t token
		t.tokenType = StringToken
		str, index := GetStrSlice(code)
		t.value = str
		*i += index
		*res = append(*res, t)
	}
}

func Tokenize(code string) []token {
	res := []token{}
	temp := []rune{}
	tokenMap := make(map[rune]func(*[]token, string, *int))
	InitTokenMap(tokenMap)
	for i := 0; i < len(code); i++ {
		if val, ok := tokenMap[rune(code[i])]; ok {
			if len(temp) > 0 {
				var t token
				t.tokenType = Identifier
				t.value = string(temp)
				res = append(res, t)
				temp = []rune{}
			}
			val(&res, code[i:], &i)
		} else if code[i] == ' ' {
			var t token
			t.tokenType = Identifier
			t.value = string(temp)
			res = append(res, t)
			temp = []rune{}
		} else {
			exclude := []string{"\n", "\t"}
			if !StrArrIncludes([]string{string(code[i])}, exclude...) {
				temp = append(temp, rune(code[i]))
			}
		}
	}
	for i, v := range res {
		if i > 0 {
			if res[i-1].tokenType == Identifier {
				fmt.Print(" ")
			}
		}
		fmt.Print(v.value)
	}
	return res
}
