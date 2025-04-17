package parser

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func (p *Policy) MarshalCedar(buf *bytes.Buffer) {
	for _, a := range p.Annotations {
		marshalAnnotation(a, buf)
		buf.WriteRune('\n')
	}
	marshalEffect(p.Effect, buf)
	buf.WriteRune(' ')
	p.marshalScope(buf)

	for _, c := range p.Conditions {
		buf.WriteRune('\n')
		marshalCondition(c, buf)
	}

	buf.WriteRune(';')
}

// scopeToNode is copied in from eval, with the expectation that
// eval will not be using it in the future.
func scopeToNode(varNode ast.NodeTypeVariable, in ast.IsScopeNode) ast.Node {
	switch t := in.(type) {
	case ast.ScopeTypeAll:
		return ast.True()
	case ast.ScopeTypeEq:
		return ast.NewNode(varNode).Equal(ast.Value(t.Entity))
	case ast.ScopeTypeIn:
		return ast.NewNode(varNode).In(ast.Value(t.Entity))
	case ast.ScopeTypeInSet:
		set := make([]ast.Node, len(t.Entities))
		for i, e := range t.Entities {
			set[i] = ast.Value(e)
		}
		return ast.NewNode(varNode).In(ast.Set(set...))
	case ast.ScopeTypeIs:
		return ast.NewNode(varNode).Is(t.Type)

	case ast.ScopeTypeIsIn:
		return ast.NewNode(varNode).IsIn(t.Type, ast.Value(t.Entity))
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}

func (p *Policy) marshalScope(buf *bytes.Buffer) {
	_, principalAll := p.Principal.(ast.ScopeTypeAll)
	_, actionAll := p.Action.(ast.ScopeTypeAll)
	_, resourceAll := p.Resource.(ast.ScopeTypeAll)
	if principalAll && actionAll && resourceAll {
		buf.WriteString("( " + consts.Principal + ", " + consts.Action + ", " + consts.Resource + " )")
		return
	}

	buf.WriteString("(\n    ")
	if principalAll {
		buf.WriteString(consts.Principal)
	} else {
		astNodeToMarshalNode(scopeToNode(ast.NewPrincipalNode(), p.Principal).AsIsNode()).marshalCedar(buf)
	}
	buf.WriteString(",\n    ")
	if actionAll {
		buf.WriteString(consts.Action)
	} else {
		astNodeToMarshalNode(scopeToNode(ast.NewActionNode(), p.Action).AsIsNode()).marshalCedar(buf)
	}
	buf.WriteString(",\n    ")
	if resourceAll {
		buf.WriteString(consts.Resource)
	} else {
		astNodeToMarshalNode(scopeToNode(ast.NewResourceNode(), p.Resource).AsIsNode()).marshalCedar(buf)
	}
	buf.WriteString("\n)")
}

func marshalAnnotation(n ast.AnnotationType, buf *bytes.Buffer) {
	buf.WriteRune('@')
	buf.WriteString(string(n.Key))
	buf.WriteRune('(')
	buf.Write(n.Value.MarshalCedar())
	buf.WriteString(")")
}

func marshalEffect(e ast.Effect, buf *bytes.Buffer) {
	if e == ast.EffectPermit {
		buf.WriteString("permit")
	} else {
		buf.WriteString("forbid")
	}
}

func (n NodeTypeVariable) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString(string(n.Name))
}

func marshalCondition(c ast.ConditionType, buf *bytes.Buffer) {
	if c.Condition == ast.ConditionWhen {
		buf.WriteString("when")
	} else {
		buf.WriteString("unless")
	}

	buf.WriteString(" { ")
	astNodeToMarshalNode(c.Body).marshalCedar(buf)
	buf.WriteString(" }")
}

func (n NodeValue) marshalCedar(buf *bytes.Buffer) {
	buf.Write(n.Value.MarshalCedar())
}

func marshalChildNode(thisNodePrecedence nodePrecedenceLevel, childAstNode ast.IsNode, buf *bytes.Buffer) {
	childNode := astNodeToMarshalNode(childAstNode)
	if thisNodePrecedence > childNode.precedenceLevel() {
		buf.WriteRune('(')
		childNode.marshalCedar(buf)
		buf.WriteRune(')')
	} else {
		childNode.marshalCedar(buf)
	}
}

func (n NodeTypeNot) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('!')
	marshalChildNode(n.precedenceLevel(), n.NodeTypeNot.Arg, buf)
}

func (n NodeTypeNegate) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('-')
	marshalChildNode(n.precedenceLevel(), n.NodeTypeNegate.Arg, buf)
}

func canMarshalAsIdent(s string) bool {
	for i, r := range s {
		if !isIdentRune(r, i == 0) {
			return false
		}
	}
	return true
}

func (n NodeTypeAccess) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)

	if canMarshalAsIdent(string(n.Value)) {
		buf.WriteRune('.')
		buf.WriteString(string(n.Value))
	} else {
		buf.WriteRune('[')
		buf.Write(n.Value.MarshalCedar())
		buf.WriteRune(']')
	}
}

func (n NodeTypeExtensionCall) marshalCedar(buf *bytes.Buffer) {
	var args []ast.IsNode
	info := extensions.ExtMap[n.Name]
	if info.IsMethod {
		marshalChildNode(n.precedenceLevel(), n.Args[0], buf)
		buf.WriteRune('.')
		args = n.Args[1:]
	} else {
		args = n.Args
	}
	buf.WriteString(string(n.Name))
	buf.WriteRune('(')
	for i := range args {
		marshalChildNode(n.precedenceLevel(), args[i], buf)
		if i != len(args)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(')')
}

func (n NodeTypeContains) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".contains(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n NodeTypeContainsAll) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".containsAll(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n NodeTypeContainsAny) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".containsAny(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n NodeTypeIsEmpty) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(".isEmpty()")
}

func (n NodeTypeGetTag) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".getTag(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n NodeTypeHasTag) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".hasTag(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n NodeTypeSet) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('[')
	for i := range n.Elements {
		marshalChildNode(n.precedenceLevel(), n.Elements[i], buf)
		if i != len(n.Elements)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
}

func (n NodeTypeRecord) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('{')
	for i := range n.Elements {
		buf.Write(n.Elements[i].Key.MarshalCedar())
		buf.WriteString(":")
		marshalChildNode(n.precedenceLevel(), n.NodeTypeRecord.Elements[i].Value, buf)
		if i != len(n.Elements)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune('}')
}

func marshalInfixBinaryOp(n ast.BinaryNode, precedence nodePrecedenceLevel, op string, buf *bytes.Buffer) {
	marshalChildNode(precedence, n.Left, buf)
	buf.WriteRune(' ')
	buf.WriteString(op)
	buf.WriteRune(' ')
	marshalChildNode(precedence, n.Right, buf)
}

func (n NodeTypeMult) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "*", buf)
}

func (n NodeTypeAdd) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "+", buf)
}

func (n NodeTypeSub) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "-", buf)
}

func (n NodeTypeLessThan) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "<", buf)
}

func (n NodeTypeLessThanOrEqual) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "<=", buf)
}

func (n NodeTypeGreaterThan) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), ">", buf)
}

func (n NodeTypeGreaterThanOrEqual) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), ">=", buf)
}

func (n NodeTypeEquals) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "==", buf)
}

func (n NodeTypeNotEquals) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "!=", buf)
}

func (n NodeTypeIn) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "in", buf)
}

func (n NodeTypeAnd) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "&&", buf)
}

func (n NodeTypeOr) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.BinaryNode, n.precedenceLevel(), "||", buf)
}

func (n NodeTypeHas) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(" has ")
	if canMarshalAsIdent(string(n.Value)) {
		buf.WriteString(string(n.Value))
	} else {
		buf.Write(n.Value.MarshalCedar())
	}
}

func (n NodeTypeIs) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.EntityType))
}

func (n NodeTypeIsIn) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.EntityType))
	buf.WriteString(" in ")
	marshalChildNode(n.precedenceLevel(), n.Entity, buf)
}

func (n NodeTypeLike) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(" like ")
	buf.Write(n.Value.MarshalCedar())
}

func (n NodeTypeIf) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString("if ")
	marshalChildNode(n.precedenceLevel(), n.If, buf)
	buf.WriteString(" then ")
	marshalChildNode(n.precedenceLevel(), n.Then, buf)
	buf.WriteString(" else ")
	marshalChildNode(n.precedenceLevel(), n.Else, buf)
}

func astNodeToMarshalNode(astNode ast.IsNode) IsNode {
	switch v := astNode.(type) {
	case ast.NodeTypeIfThenElse:
		return NodeTypeIf{v}
	case ast.NodeTypeOr:
		return NodeTypeOr{v}
	case ast.NodeTypeAnd:
		return NodeTypeAnd{v}
	case ast.NodeTypeLessThan:
		return NodeTypeLessThan{v, relationPrecedenceNode{}}
	case ast.NodeTypeLessThanOrEqual:
		return NodeTypeLessThanOrEqual{v, relationPrecedenceNode{}}
	case ast.NodeTypeGreaterThan:
		return NodeTypeGreaterThan{v, relationPrecedenceNode{}}
	case ast.NodeTypeGreaterThanOrEqual:
		return NodeTypeGreaterThanOrEqual{v, relationPrecedenceNode{}}
	case ast.NodeTypeNotEquals:
		return NodeTypeNotEquals{v, relationPrecedenceNode{}}
	case ast.NodeTypeEquals:
		return NodeTypeEquals{v, relationPrecedenceNode{}}
	case ast.NodeTypeIn:
		return NodeTypeIn{v, relationPrecedenceNode{}}
	case ast.NodeTypeHas:
		return NodeTypeHas{v, relationPrecedenceNode{}}
	case ast.NodeTypeHasTag:
		return NodeTypeHasTag{v, accessPrecedenceNode{}}
	case ast.NodeTypeLike:
		return NodeTypeLike{v, relationPrecedenceNode{}}
	case ast.NodeTypeIs:
		return NodeTypeIs{v, relationPrecedenceNode{}}
	case ast.NodeTypeIsIn:
		return NodeTypeIsIn{v, relationPrecedenceNode{}}
	case ast.NodeTypeSub:
		return NodeTypeSub{v, addPrecedenceNode{}}
	case ast.NodeTypeAdd:
		return NodeTypeAdd{v, addPrecedenceNode{}}
	case ast.NodeTypeMult:
		return NodeTypeMult{v}
	case ast.NodeTypeNegate:
		return NodeTypeNegate{v, UnaryNode{}}
	case ast.NodeTypeNot:
		return NodeTypeNot{v, UnaryNode{}}
	case ast.NodeTypeAccess:
		return NodeTypeAccess{v, accessPrecedenceNode{}}
	case ast.NodeTypeGetTag:
		return NodeTypeGetTag{v, accessPrecedenceNode{}}
	case ast.NodeTypeExtensionCall:
		return NodeTypeExtensionCall{v, accessPrecedenceNode{}}
	case ast.NodeTypeContains:
		return NodeTypeContains{v, accessPrecedenceNode{}}
	case ast.NodeTypeContainsAll:
		return NodeTypeContainsAll{v, accessPrecedenceNode{}}
	case ast.NodeTypeContainsAny:
		return NodeTypeContainsAny{v, accessPrecedenceNode{}}
	case ast.NodeTypeIsEmpty:
		return NodeTypeIsEmpty{v, accessPrecedenceNode{}}
	case ast.NodeValue:
		return NodeValue{v, primaryPrecedenceNode{}}
	case ast.NodeTypeRecord:
		return NodeTypeRecord{v, primaryPrecedenceNode{}}
	case ast.NodeTypeSet:
		return NodeTypeSet{v, primaryPrecedenceNode{}}
	case ast.NodeTypeVariable:
		return NodeTypeVariable{v, primaryPrecedenceNode{}}
	default:
		panic(fmt.Sprintf("unknown node type %T", v))
	}
}
