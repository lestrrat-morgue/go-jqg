package query

import "errors"

func Parse(s string) (*Query, error) {
	if len(s) == 0 {
		return nil, errors.New(`empty query`)
	}

	return nil, errors.New(`unimplemented`)
}
