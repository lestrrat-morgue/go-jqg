package ast

type Builder struct {

}

type Node interface {
}

type Expr interface {}

// "|" operator combines two filters by feeding the output(s) of the
// Src to Dst
// If Src produces multiple results, the one on the right will be run
// for each of those results
type PipeExpr struct {
	Src Expr
	Dst Expr
}

// "." filter takes an input and produces it unchanged as output
type SelfFilterExpr struct {

}

// ".foo", "."foo$"", "
type FieldLookupExpr struct {
}

type SliceLookupExpr struct {
}

type AST struct {
	children nodeList
}

type Appender interface {
	Append(Node) error
}