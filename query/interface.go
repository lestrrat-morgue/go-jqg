package query

import "github.com/lestrrat/jqg/token"

type Query struct {
}

type Item struct {
	Type  token.Token
	Value interface{}
}
