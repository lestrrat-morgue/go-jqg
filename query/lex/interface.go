package lex

import "github.com/lestrrat/go-jqg/query/token"

type Lexer struct{}

type Item struct {
	Type  token.Token
	Value interface{}
}

type Iterator struct {
	ch   chan Item
	next Item
}
