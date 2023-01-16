package main

import (
	"strings"

	"github.com/valyala/fastjson/fastfloat"
)

type TokenType string

const (
	Identifier  TokenType = "Identifier"
	OpenParen   TokenType = "OpenParen"
	CloseParen  TokenType = "CloseParen"
	ListToken   TokenType = "ListToken"
	StringToken TokenType = "StringToken"
	BoolToken   TokenType = "BoolToken"
	NumberToken TokenType = "NumberToken"
	UnTokenized TokenType = "UnTokenized"
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

func GetToken(val string) token {
	var t token
	if len(val) == 0 {
		t.tokenType = Identifier
	} else if strings.Index(val, "\"") == 0 {
		t.tokenType = StringToken
	} else if strings.Index(val, "[") == 0 {
		t.tokenType = ListToken
	} else if val == "true" || val == "false" {
		t.tokenType = BoolToken
	} else {
		_, err := fastfloat.Parse(val)
		if err == nil {
			t.tokenType = NumberToken
		} else {
			t.tokenType = Identifier
		}
	}
	t.value = val
	return t
}

func Tokenize(code string) []token {
	res := []token{}
	temp := []rune{}
	tokenMap := make(map[rune]func(*[]token, string, *int))
	InitTokenMap(tokenMap)
	for i := 0; i < len(code); i++ {
		if val, ok := tokenMap[rune(code[i])]; ok {
			if len(temp) > 0 {
				t := GetToken(string(temp))
				res = append(res, t)
				temp = []rune{}
			}
			val(&res, code[i:], &i)
		} else if code[i] == ' ' {
			t := GetToken(string(temp))
			res = append(res, t)
			temp = []rune{}
		} else {
			exclude := []string{"\n", "\t"}
			if !StrArrIncludes([]string{string(code[i])}, exclude...) {
				temp = append(temp, rune(code[i]))
			}
		}
	}
	return res
}
