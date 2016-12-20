package lex

import (
	"context"
	"testing"

	"github.com/lestrrat/go-jqg/query/token"
	"github.com/stretchr/testify/assert"
)

func TestLex(t *testing.T) {
	assertNoError := func(t *testing.T, item Item) bool {
		if !assert.False(t, item.Type == token.ILLEGAL, "received error") {
			t.Logf("%s", item.Value)
			return false
		}
		t.Logf("%#v", item)
		return true
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := New()
	t.Run(`.`, func(t *testing.T) {
		for iter := l.Do(ctx, `.`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
	t.Run(`.foo`, func(t *testing.T) {
		for iter := l.Do(ctx, `.foo`); iter.Next(ctx); {
			t.Logf("%#v", iter.Item())
		}
	})
	t.Run(`."foo$"`, func(t *testing.T) {
		for iter := l.Do(ctx, `."foo$"`); iter.Next(ctx); {
			t.Logf("%#v", iter.Item())
		}
	})
	t.Run(`.foo?`, func(t *testing.T) {
		for iter := l.Do(ctx, `.foo?`); iter.Next(ctx); {
			t.Logf("%#v", iter.Item())
		}
	})
	t.Run(`.["foo"]`, func(t *testing.T) {
		for iter := l.Do(ctx, `.["foo"]`); iter.Next(ctx); {
			t.Logf("%#v", iter.Item())
		}
	})
	t.Run(`.[10]`, func(t *testing.T) {
		for iter := l.Do(ctx, `.[10]`); iter.Next(ctx); {
			t.Logf("%#v", iter.Item())
		}
	})
	t.Run(`.[10:100]`, func(t *testing.T) {
		for iter := l.Do(ctx, `.[10:100]`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
	t.Run(`.[]`, func(t *testing.T) {
		for iter := l.Do(ctx, `.[]`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
	t.Run(`.[]?`, func(t *testing.T) {
		for iter := l.Do(ctx, `.[]?`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
	t.Run(`.foo, .bar`, func(t *testing.T) {
		for iter := l.Do(ctx, `.foo, .bar`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
	t.Run(`.foo | .bar`, func(t *testing.T) {
		for iter := l.Do(ctx, `.foo | .bar`); iter.Next(ctx); {
			if !assertNoError(t, iter.Item()) {
				return
			}
		}
	})
}
