package query

import "testing"

func TestParse(t *testing.T) {
	t.Run(`.`, func(t *testing.T) {
		Parse(`.`)
	})
	t.Run(`.foo`, func(t *testing.T) {
		Parse(`.foo`)
	})
	t.Run(`."foo$"`, func(t *testing.T) {
		Parse(`."foo$"`)
	})
	t.Run(`.foo?`, func(t *testing.T) {
		Parse(`.foo?`)
	})
	t.Run(`.["foo"]`, func(t *testing.T) {
		Parse(`.["foo"]`)
	})
	t.Run(`.[10]`, func(t *testing.T) {
		Parse(`.[10]`)
	})
	t.Run(`.[10:100]`, func(t *testing.T) {
		Parse(`.[10:100]`)
	})
	t.Run(`.[]`, func(t *testing.T) {
		Parse(`.[]`)
	})
	t.Run(`.[]?`, func(t *testing.T) {
		Parse(`.[]?`)
	})
	t.Run(`.foo, .bar`, func(t *testing.T) {
		Parse(`.foo, .bar`)
	})
	t.Run(`.foo | .bar`, func(t *testing.T) {
		Parse(`.foo | .bar`)
	})
}
