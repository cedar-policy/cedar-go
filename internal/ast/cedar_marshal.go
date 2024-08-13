package ast

import (
	"bytes"
)

// TODO: Add errors to all of this! TODO: review this ask, I'm not sure any real errors are possible.  All buf errors are panics.
func (p *Policy) MarshalCedar(buf *bytes.Buffer) {
	for _, a := range p.Annotations {
		a.MarshalCedar(buf)
		buf.WriteRune('\n')
	}
	p.Effect.MarshalCedar(buf)
	buf.WriteRune(' ')
	p.marshalScope(buf)

	for _, c := range p.Conditions {
		buf.WriteRune('\n')
		c.MarshalCedar(buf)
	}

	buf.WriteRune(';')
}

func (p *Policy) marshalScope(buf *bytes.Buffer) {
	_, principalAll := p.Principal.(ScopeTypeAll)
	_, actionAll := p.Action.(ScopeTypeAll)
	_, resourceAll := p.Resource.(ScopeTypeAll)
	if principalAll && actionAll && resourceAll {
		buf.WriteString("( principal, action, resource )")
		return
	}

	buf.WriteString("(\n    ")
	p.Principal.MarshalCedar(buf)
	buf.WriteString(",\n    ")
	p.Action.MarshalCedar(buf)
	buf.WriteString(",\n    ")
	p.Resource.MarshalCedar(buf)
	buf.WriteString("\n)")
}

func (n AnnotationType) MarshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('@')
	buf.WriteString(string(n.Key))
	buf.WriteRune('(')
	buf.WriteString(n.Value.Cedar())
	buf.WriteString(")")
}

func (e Effect) MarshalCedar(buf *bytes.Buffer) {
	if e == EffectPermit {
		buf.WriteString("permit")
	} else {
		buf.WriteString("forbid")
	}
}

func (n NodeTypeVariable) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString(string(n.Name))
}

func (n ScopeTypeAll) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
}

func (n ScopeTypeEq) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" == ")
	buf.WriteString(n.Entity.Cedar())
}

func (n ScopeTypeIn) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" in ")
	buf.WriteString(n.Entity.Cedar())
}

func (n ScopeTypeInSet) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" in ")
	buf.WriteRune('[')
	for i := range n.Entities {
		buf.WriteString(n.Entities[i].Cedar())
		if i != len(n.Entities)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
}

func (n ScopeTypeIs) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.Type))
}

func (n ScopeTypeIsIn) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.Type))
	buf.WriteString(" in ")
	buf.WriteString(n.Entity.Cedar())
}

func (c ConditionType) MarshalCedar(buf *bytes.Buffer) {
	if c.Condition == ConditionWhen {
		buf.WriteString("when")
	} else {
		buf.WriteString("unless")
	}

	buf.WriteString(" { ")
	c.Body.marshalCedar(buf)
	buf.WriteString(" }")
}

func (n NodeValue) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString(n.Value.Cedar())
}

func marshalChildNode(thisNodePrecedence nodePrecedenceLevel, childNode IsNode, buf *bytes.Buffer) {
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
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
}

func (n NodeTypeNegate) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('-')
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
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
		buf.WriteString(n.Value.Cedar())
		buf.WriteRune(']')
	}
}

func (n NodeTypeExtensionCall) marshalCedar(buf *bytes.Buffer) {
	var args []IsNode
	info := ExtMap[n.Name]
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
		marshalChildNode(n.precedenceLevel(), n.Args[i], buf)
		if i != len(n.Args)-1 {
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
		buf.WriteString(n.Elements[i].Key.Cedar())
		buf.WriteString(":")
		marshalChildNode(n.precedenceLevel(), n.Elements[i].Value, buf)
		if i != len(n.Elements)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune('}')
}

func marshalInfixBinaryOp(n BinaryNode, precedence nodePrecedenceLevel, op string, buf *bytes.Buffer) {
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
		buf.WriteString(n.Value.Cedar())
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
	n.Entity.marshalCedar(buf)
}

func (n NodeTypeLike) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(" like ")
	buf.WriteString(n.Value.Cedar())
}

func (n NodeTypeIf) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString("if ")
	marshalChildNode(n.precedenceLevel(), n.If, buf)
	buf.WriteString(" then ")
	marshalChildNode(n.precedenceLevel(), n.Then, buf)
	buf.WriteString(" else ")
	marshalChildNode(n.precedenceLevel(), n.Else, buf)
}
