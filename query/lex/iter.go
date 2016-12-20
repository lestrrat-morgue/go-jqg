package lex

import "context"

func (iter *Iterator) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case c, ok := <-iter.ch:
		if !ok {
			return false
		}
		iter.next = c
		return true
	}
}

func (iter *Iterator) Item() Item {
	return iter.next
}
