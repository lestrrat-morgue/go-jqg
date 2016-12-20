package ast

import (
	"context"
	"errors"

	"github.com/lestrrat/go-jqg/query/lex"
	"github.com/lestrrat/go-jqg/query/token"
)

func NewBuilder() *Builder {
	return &Builder{}
}

type buildCtx struct {
	context.Context
	iter *lex.Iterator
	tokens [3]lex.Item
	peekCount int
	root *AST
	current Node
}

func (c *buildCtx) Peek() lex.Item {
	if c.peekCount >0 {
		return c.tokens[c.peekCount-1]
	}

	if !c.iter.Next(c) {
		panic("fudge")
	}

	c.peekCount = 1
	c.tokens[0] = c.iter.Item()
	return c.tokens[0]
}

func (c *buildCtx) Next() lex.Item {
	if c.peekCount > 0 {
		c.peekCount--
	} else {
		if !c.iter.Next(c) {
			panic("hello")
		}
		c.tokens[0] = c.iter.Item()
	}
	return c.tokens[c.peekCount]
}

func (b *Builder) Run(ctx context.Context, txt string) (*AST, error) {

	iter := lex.New().Do(ctx, txt)

	bctx := buildCtx{
		Context: ctx,
		iter:    iter,
		root:    &AST{},
	}
	bctx.current = bctx.root

	err := b.ParseStart(&bctx)
	if err != nil {
		return nil, err
	}

	return bctx.root, nil
}

func (b *Builder) ParseStart(ctx *buildCtx) error{
	n, err := b.ParseFilter(ctx)
	if err != nil {
		return err
	}

	appender, ok := ctx.current.(Appender)
	if !ok {
		return errors.New(`current node cannot have children`)
	}
	appender.Append(n)
	return nil
}

func (b *Builder) ParseFilter(ctx *buildCtx) (Node, error) {
	if ctx.Next().Type != token.PERIOD {
		return nil, errors.New(`expected filter expression`)
	}

	tok := ctx.Peek()
	switch tok.Type {
	case token.STRING:
		return NewFieldLookupNode(tok.Value.(string)), nil
	case token.EOF:
		return NewSelfNode(), nil
	}
	return nil, errors.New(`unimplemented`)
}
