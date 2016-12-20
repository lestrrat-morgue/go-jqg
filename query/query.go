package query

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/lestrrat/go-jqg/token"
)

type filter interface {
	Apply(interface{}) (interface{}, error)
}

type filterFn func(interface{}) (interface{}, error)

func (fn filterFn) Apply(v interface{}) (interface{}, error) {
	return fn(v)
}

func Parse(s string) (*Query, error) {
	var q Query

	if len(s) == 0 {
		return nil, errors.New(`empty query`)
	}

	ch := make(chan Item)
	go lex(ch, s)

	for i := range ch {
		fmt.Printf("%s\n", i)
	}

	return &q, nil
}

func lex(ch chan Item, s string) {
	r := NewStringReader(s)

	var fn lexFn
	for fn = lexExpr; fn != nil; {
		fn = fn(ch, r)
	}

	close(ch)
}

type lexFn func(chan Item, *StringReader) lexFn

func lexExpr(ch chan Item, rdr *StringReader) lexFn {
	fmt.Println("lexStart")
	rdr.SkipSpaces()

	r, _, err := rdr.PeekRune()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	switch r {
	case '.':
		return lexFilter
	}

	// If anything else, we have a problem
	fmt.Println("expected filter")
	return nil
}

func lexFilter(ch chan Item, rdr *StringReader) lexFn {
	fmt.Println("lexFilter")
	rdr.SkipSpaces()

	// Read first period
	r, _, err := rdr.ReadRune()
	if err != nil {
		fmt.Println("read rune failed")
		return nil
	}

	if r != '.' {
		fmt.Printf("rune is not '.' %c\n", r)
		return nil
	}
	ch <- Item{Type: token.PERIOD}

	rdr.SkipSpaces()
	r, _, err = rdr.PeekRune()
	switch {
	case r == '"':
		return lexQuotedStringField
	case unicode.IsLetter(r):
		return lexIdentField
	case r == '[':
		return lexIndex
	}
	// error
	return nil
}

func lexConnectorOrFilter(ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()

	r, _, err := rdr.PeekRune()
	if err != nil {
		// error
		return nil
	}

	switch r {
	case '|':
		rdr.ReadRune()
		ch <- Item{Type: token.PIPE}
		return lexFilter
	case ',':
		rdr.ReadRune()
		ch <- Item{Type: token.COMMA}
		return lexFilter
	case '.':
		return lexFilter
	}
	return nil
}

func lexIdentField(ch chan Item, rdr *StringReader) lexFn {
	fmt.Println("lexIdent")
	start := rdr.Offset()
	fmt.Printf("start -> %d\n", start)

	for {
		r, _, err := rdr.PeekRune()
		if err != nil && err != io.EOF {
			fmt.Printf("peek rune failed %s\n", err)
			return nil
		}
		if !unicode.IsLetter(r) {
			ch <- Item{
				Type:  token.STRING,
				Value: rdr.Slice(int(start), int(rdr.Offset())),
			}
			if err := maybeQuestion(ch, rdr); err != nil {
				return nil
			}
			return lexConnectorOrFilter
		}
		rdr.ReadRune()
	}
	return nil
}

func acceptQuotedString(ch chan Item, rdr *StringReader) error {
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
			ch <- Item{
				Type:  token.STRING,
				Value: rdr.Slice(int(start), int(rdr.Offset())),
			}
		}
		rdr.ReadRune()
	}
	return errors.New(`not reachable`)
}

func lexQuotedStringField(ch chan Item, rdr *StringReader) lexFn {
	fmt.Println("lexQuotedString")
	if err := acceptQuotedString(ch, rdr); err != nil {
		return nil
	}
	if err := maybeQuestion(ch, rdr); err != nil {
		return nil
	}
	return lexConnectorOrFilter
}

func lexIndex(ch chan Item, rdr *StringReader) lexFn {
	rdr.SkipSpaces()
	r, _, err := rdr.ReadRune()
	if err != nil {
		return nil
	}

	if r != '[' {
		return nil
	}

	ch <- Item{Type: token.LBRACK}

	rdr.SkipSpaces()
	r, _, err = rdr.PeekRune()
	switch {
	case r == ']':
		// empty slice index is accepted
	case r == '"':
		if err := acceptQuotedString(ch, rdr); err != nil {
			return nil
		}
	default:
		if err := acceptNumberOrRange(ch, rdr); err != nil {
			return nil
		}
	}
	r, _, err = rdr.ReadRune()
	if r != ']' {
		// error
		return nil
	}
	ch <- Item{Type: token.RBRACK}
	if err := maybeQuestion(ch, rdr); err != nil {
		return nil
	}
	return lexConnectorOrFilter
}

func acceptNumberOrRange(ch chan Item, rdr *StringReader) error {
	if err := acceptNumber(ch, rdr); err != nil {
		return err
	}

	rdr.SkipSpaces()
	r, _, _ := rdr.PeekRune()
	if r == ':' {
		ch <- Item{Type: token.COLON}
		rdr.ReadRune()
		rdr.SkipSpaces()
		if err := acceptNumber(ch, rdr); err != nil {
			return err
		}
	}
	return nil
}

func acceptNumber(ch chan Item, rdr *StringReader) error {
	start := rdr.Offset() - 1
	fmt.Println("acceptNumber")
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
			ch <- Item{
				Type:  token.INT,
				Value: v,
			}
			return nil
		}
		rdr.ReadRune()
	}
	return nil
}

func maybeQuestion(ch chan Item, rdr *StringReader) error {
	r, _, err := rdr.PeekRune()
	if err != nil {
		return err
	}

	if r == '?' {
		ch <- Item{Type: token.QUESTION}
		rdr.ReadRune()
	}
	return nil
}
