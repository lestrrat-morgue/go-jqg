package lex

import (
	"context"
	"io"
	"strconv"
	"unicode"

	"github.com/lestrrat/go-jqg/query/token"
	"github.com/pkg/errors"
)

type lexFn func(context.Context, chan Item, *StringReader) lexFn

func New() *Lexer {
	return &Lexer{}
}

func (l *Lexer) Do(ctx context.Context, src string) *Iterator {
	ch := make(chan Item)
	rdr := NewStringReader(src)

	go l.run(ctx, ch, rdr)

	return &Iterator{ch: ch}
}

func (l *Lexer) run(ctx context.Context, ch chan Item, rdr *StringReader) {
	defer close(ch)

	var fn lexFn
	fn = lexExpr
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if fn = fn(ctx, ch, rdr); fn == nil {
				return
			}
		}
	}
}

func emit(ctx context.Context, ch chan Item, item Item) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- item:
		return true
	}
}

func lexExpr(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()

	r, _, err := rdr.PeekRune()
	if err != nil {
		if err == io.EOF {
			emit(ctx, ch, Item{Type: token.EOF})
			return nil
		}
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.Wrap(err, `lexExpr: PeekRune`)})
		return nil
	}

	switch r {
	case '.':
		return lexFilter
	}

	// If anything else, we have a problem
	return nil
}

func lexFilter(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()

	// Read first period
	r, _, err := rdr.ReadRune()
	if err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.Wrap(err, `lexFilter: ReadRune`)})
		return nil
	}

	if r != '.' {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.New(`lexFilter: expected '.'`)})
		return nil
	}

	if !emit(ctx, ch, Item{Type: token.PERIOD}) {
		return nil
	}

	rdr.SkipSpaces()
	r, _, err = rdr.PeekRune()
	if err != nil {
		if err == io.EOF {
			emit(ctx, ch, Item{Type: token.EOF})
		} else {
			emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.Wrap(err, `lexFilter: PeekRune`)})
		}
		return nil
	}

	switch {
	case r == '"':
		return lexQuotedStringField
	case unicode.IsLetter(r):
		return lexIdentField
	case r == '[':
		return lexIndex
	}

	if err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}
	return lexExpr
}

func lexConnectorOrFilter(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()

	r, _, err := rdr.PeekRune()
	if err != nil {
		if err == io.EOF {
			emit(ctx, ch, Item{Type: token.EOF})
		} else {
			emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		}
		return nil
	}

	switch r {
	case '|':
		rdr.ReadRune()
		if emit(ctx, ch, Item{Type: token.PIPE}) {
			return lexFilter
		}
		return nil
	case ',':
		rdr.ReadRune()
		if emit(ctx, ch, Item{Type: token.COMMA}) {
			return lexFilter
		}
		return nil
	case '.':
		return lexFilter
	}
	return nil
}

func lexIdentField(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	start := rdr.Offset()

	for {
		r, _, err := rdr.PeekRune()
		if err != nil {
			if err == io.EOF {
				emit(ctx, ch, Item{Type: token.EOF})
			} else {
				emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
			}
			return nil
		}
		if !unicode.IsLetter(r) {
			if !emit(ctx, ch, Item{
				Type:  token.STRING,
				Value: rdr.Slice(int(start), int(rdr.Offset())),
			}) {
				return nil
			}
			if err := maybeQuestion(ctx, ch, rdr); err != nil {
				emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
				return nil
			}
			return lexConnectorOrFilter
		}
		rdr.ReadRune()
	}
	return nil
}

func acceptQuotedString(ctx context.Context, ch chan Item, rdr *StringReader) error {
	r, _, err := rdr.ReadRune() // consume the quote
	if r != '"' {
		return errors.New(`expected double quote`)
	}

	start := rdr.Offset()
	for {
		r, _, err = rdr.PeekRune()
		if err != nil {
			return err
		}

		if r == '"' {
			defer rdr.ReadRune()
			emit(ctx, ch, Item{
				Type:  token.STRING,
				Value: rdr.Slice(int(start), int(rdr.Offset())),
			})
			return nil
		}
		rdr.ReadRune()
	}
	return errors.New(`not reachable`)
}

func lexQuotedStringField(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	if err := acceptQuotedString(ctx, ch, rdr); err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}
	if err := maybeQuestion(ctx, ch, rdr); err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}
	return lexConnectorOrFilter
}

func lexIndex(ctx context.Context, ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()
	r, _, err := rdr.ReadRune()
	if err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}

	if r != '[' {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.New(`expected '['`)})
		return nil
	}

	emit(ctx, ch, Item{Type: token.LBRACK})

	rdr.SkipSpaces()
	r, _, err = rdr.PeekRune()
	if err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}
	switch {
	case r == ']':
		// empty slice index is accepted
	case r == '"':
		if err := acceptQuotedString(ctx, ch, rdr); err != nil {
			emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
			return nil
		}
	default:
		if err := acceptNumberOrRange(ctx, ch, rdr); err != nil {
			emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
			return nil
		}
	}
	r, _, err = rdr.ReadRune()
	if r != ']' {
		// error
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: errors.New(`expected ']'`)})
		return nil
	}
	emit(ctx, ch, Item{Type: token.RBRACK})
	if err := maybeQuestion(ctx, ch, rdr); err != nil {
		emit(ctx, ch, Item{Type: token.ILLEGAL, Value: err})
		return nil
	}
	return lexConnectorOrFilter
}

func acceptNumberOrRange(ctx context.Context, ch chan Item, rdr *StringReader) error {
	if err := acceptNumber(ctx, ch, rdr); err != nil {
		return err
	}

	rdr.SkipSpaces()
	r, _, _ := rdr.PeekRune()
	if r == ':' {
		rdr.ReadRune()
		if !emit(ctx, ch, Item{Type: token.COLON}) {
			return errors.New(`failed to emit`)
		}

		rdr.SkipSpaces()
		if err := acceptNumber(ctx, ch, rdr); err != nil {
			return err
		}
	}
	return nil
}

func acceptNumber(ctx context.Context, ch chan Item, rdr *StringReader) error {
	start := rdr.Offset()
	r, _, err := rdr.ReadRune()
	if err != nil {
		return err
	}

	if !unicode.IsDigit(r) {
		return errors.New(`expected number`)
	}
	for {
		r, _, err := rdr.PeekRune()
		if err != nil {
			return err
		}

		if !unicode.IsDigit(r) {
			v, err := strconv.ParseInt(rdr.Slice(int(start), int(rdr.Offset()-1)), 10, 64)
			if err != nil {
				return err
			}
			emit(ctx, ch, Item{Type: token.INT, Value: v})
			return nil
		}
		rdr.ReadRune()
	}
	return nil
}

func maybeQuestion(ctx context.Context, ch chan Item, rdr *StringReader) error {
	r, _, err := rdr.PeekRune()
	if err != nil && err != io.EOF {
		return err
	}

	if r == '?' {
		emit(ctx, ch, Item{Type: token.QUESTION})
		rdr.ReadRune()
	}
	return nil
}
