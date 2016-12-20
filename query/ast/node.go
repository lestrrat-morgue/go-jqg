package ast

type nodeList []Node

func (nl *nodeList) Append(n Node) error {
	*nl = append(*nl, n)
	return nil
}

func (ast *AST) Append(n Node) error {
	return ast.children.Append(n)
}

type SelfNode struct{}

func NewSelfNode() *SelfNode {
	return &SelfNode{}
}

type FieldLookupNode struct{}

func NewFieldLookupNode(name string) *FieldLookupNode {
	return &FieldLookupNode{}
}
