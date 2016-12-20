package lex

import (
	"fmt"
	"strings"
	"unicode"
)

type StringReader struct {
	lastrune int
	offset   int64
	str      string
	*strings.Reader
}

func NewStringReader(s string) *StringReader {
	return &StringReader{
		str:    s,
		Reader: strings.NewReader(s),
	}
}

func (r *StringReader) Slice(i, j int) string {
	fmt.Printf("i = %d, j = %d\n", i, j)
	return r.str[i:j]
}

func (r *StringReader) Offset() int64 {
	return r.offset
}

func (r *StringReader) PeekRune() (rune, int, error) {
	defer r.UnreadRune()
	return r.ReadRune()
}

func (r *StringReader) UnreadRune() error {
	r.offset -= int64(r.lastrune)
	r.lastrune = 0
	return r.Reader.UnreadRune()
}

func (r *StringReader) ReadRune() (rune, int, error) {
	v, n, err := r.Reader.ReadRune()
	r.offset += int64(n)
	r.lastrune = n
	return v, n, err
}

// Skips spaces. Does not report IO errors
func (r *StringReader) SkipSpaces() error {
	for {
		v, _, err := r.ReadRune()
		if err != nil {
			r.UnreadRune()
			return nil
		}

		if !unicode.IsSpace(v) {
			r.UnreadRune()
			return nil
		}
	}
	return nil // can't get here
}
