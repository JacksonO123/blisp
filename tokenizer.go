package main

import (
	"math"
	"strings"

	"github.com/valyala/fastjson/fastfloat"
)

type TokenType int

const (
	Identifier TokenType = iota
	OpenParen
	CloseParen
	OpenBracket
	CloseBracket
	StringToken
	BoolToken
	IntToken
	FloatToken
	NilToken
)

type token struct {
	tokenType TokenType
	value     any
}

func GetToken(val string) token {
	val = strings.TrimSpace(val)
	var t token
	if len(val) == 1 {
		switch val[0] {
		case '(':
			t.tokenType = OpenParen
		case ')':
			t.tokenType = CloseParen
		case '[':
			t.tokenType = OpenBracket
		case ']':
			t.tokenType = CloseBracket
		default:
			{
				num, err := fastfloat.Parse(val)
				if err == nil {
					if math.Floor(num) == num {
						t.tokenType = IntToken
						t.value = int(num)
					} else {
						t.tokenType = FloatToken
						t.value = num
					}
					return t
				} else {
					t.tokenType = Identifier
				}
			}
		}
	} else {
		if val == "true" {
			t.tokenType = BoolToken
			t.value = true
			return t
		} else if val == "false" {
			t.tokenType = BoolToken
			t.value = false
			return t
		} else if val == "nil" {
			t.tokenType = NilToken
			t.value = nil
		} else {
			num, err := fastfloat.Parse(val)
			if err == nil {
				if math.Floor(num) == num {
					t.tokenType = IntToken
					t.value = int(num)
				} else {
					t.tokenType = FloatToken
					t.value = num
				}
				return t
			} else {
				t.tokenType = Identifier
			}
			t.tokenType = Identifier
		}
	}
	t.value = val
	return t
}

func GetString(str string) (string, int) {
	for i, v := range str {
		if v == '"' && ((i > 0 && str[i-1] != '\\') || (i > 1 && str[i-1] == '\\' && str[i-2] == '\\')) {
			i++
			return str[:i], i
		}
	}
	return "", 0
}

func GetComment(str string) int {
	for i, v := range str {
		if v == '\n' {
			return i
		}
	}
	return 0
}

func Tokenize(code string) []token {
	res := []token{}
	temp := make([]rune, 0, len(code)/6)
	for i := 0; i < len(code); i++ {
		if code[i] == '#' {
			i += GetComment(code[i:])
		} else if code[i] == '"' {
			var t token
			t.tokenType = StringToken
			str, index := GetString(code[i:])
			t.value = str
			res = append(res, t)
			i += index - 1
		} else if code[i] == ' ' || code[i] == '\n' || code[i] == '\t' {
			if len(temp) > 0 {
				t := GetToken(string(temp))
				res = append(res, t)
			}
			temp = []rune{}
		} else {
			var t token
			switch code[i] {
			case '(':
				t.tokenType = OpenParen
			case ')':
				t.tokenType = CloseParen
			case '[':
				t.tokenType = OpenBracket
			case ']':
				t.tokenType = CloseBracket
			default:
				{
					temp = append(temp, rune(code[i]))
					continue
				}
			}
			t.value = string(code[i])
			if len(temp) > 0 {
				t := GetToken(string(temp))
				res = append(res, t)
			}
			temp = []rune{}
			res = append(res, GetToken(string(code[i])))
		}
	}
	if len(string(temp)) > 0 {
		t := GetToken(string(temp))
		res = append(res, t)
	}
	return res
}
