package ast

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := NewBuilder()
	t.Run(`.`, func(t *testing.T) {
		n, err := b.Run(ctx, `.`)
		if !assert.NoError(t, err, `Run should succeed`) {
			return
		}

		t.Logf("%#v", n)
	})
}
