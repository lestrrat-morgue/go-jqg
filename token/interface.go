package token

type Token int

const (
	ILLEGAL Token = iota
	EOF

	IDENT    // foo
	STRING   // "foo"
	INT      // 1, 2, 3, 4, 5 ...
	PERIOD   // .
	LBRACK   // [
	RBRACK   // ]
	COLON    // :
	QUESTION // ?

	// Operators
	COMMA // "," (could be slice operator, too)
	PIPE   // "|"
)

type Pos int
