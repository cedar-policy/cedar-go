package ast

import (
	"bytes"
)

// TODO: Add errors to all of this!
func (p *Policy) MarshalCedar(buf *bytes.Buffer) {
	for _, a := range p.annotations {
		a.MarshalCedar(buf)
		buf.WriteRune('\n')
	}
	p.effect.MarshalCedar(buf)
	buf.WriteRune(' ')
	p.marshalScope(buf)

	for _, c := range p.conditions {
		buf.WriteRune('\n')
		c.MarshalCedar(buf)
	}

	buf.WriteRune(';')
}

func (p *Policy) marshalScope(buf *bytes.Buffer) {
	_, principalAll := p.principal.(scopeTypeAll)
	_, actionAll := p.action.(scopeTypeAll)
	_, resourceAll := p.resource.(scopeTypeAll)
	if principalAll && actionAll && resourceAll {
		buf.WriteString("( principal, action, resource )")
		return
	}

	buf.WriteString("(\n    ")
	p.principal.MarshalCedar(buf)
	buf.WriteString(",\n    ")
	p.action.MarshalCedar(buf)
	buf.WriteString(",\n    ")
	p.resource.MarshalCedar(buf)
	buf.WriteString("\n)")
}

func (n annotationType) MarshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('@')
	buf.WriteString(string(n.Key))
	buf.WriteRune('(')
	buf.WriteString(n.Value.Cedar())
	buf.WriteString(")")
}

func (e effect) MarshalCedar(buf *bytes.Buffer) {
	if e == effectPermit {
		buf.WriteString("permit")
	} else {
		buf.WriteString("forbid")
	}
}

func (n nodeTypeVariable) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString(string(n.Name))
}

func (n scopeTypeAll) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
}

func (n scopeTypeEq) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" == ")
	buf.WriteString(n.Entity.Cedar())
}

func (n scopeTypeIn) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" in ")
	buf.WriteString(n.Entity.Cedar())
}

func (n scopeTypeInSet) MarshalCedar(buf *bytes.Buffer) {
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

func (n scopeTypeIs) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.Type))
}

func (n scopeTypeIsIn) MarshalCedar(buf *bytes.Buffer) {
	n.Variable.marshalCedar(buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.Type))
	buf.WriteString(" in ")
	buf.WriteString(n.Entity.Cedar())
}

func (c conditionType) MarshalCedar(buf *bytes.Buffer) {
	if c.Condition == conditionWhen {
		buf.WriteString("when")
	} else {
		buf.WriteString("unless")
	}

	buf.WriteString(" { ")
	c.Body.marshalCedar(buf)
	buf.WriteString(" }")
}

func (n nodeValue) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString(n.Value.Cedar())
}

func marshalChildNode(thisNodePrecedence nodePrecedenceLevel, childNode node, buf *bytes.Buffer) {
	if thisNodePrecedence > childNode.precedenceLevel() {
		buf.WriteRune('(')
		childNode.marshalCedar(buf)
		buf.WriteRune(')')
	} else {
		childNode.marshalCedar(buf)
	}
}

func (n nodeTypeNot) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('!')
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
}

func (n nodeTypeNegate) marshalCedar(buf *bytes.Buffer) {
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

func (n nodeTypeAccess) marshalCedar(buf *bytes.Buffer) {
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

func (n nodeTypeExtensionCall) marshalCedar(buf *bytes.Buffer) {
	var args []node
	if n.Name != "ip" && n.Name != "decimal" {
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

func (n nodeTypeContains) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".contains(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n nodeTypeContainsAll) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".containsAll(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n nodeTypeContainsAny) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(".containsAny(")
	marshalChildNode(n.precedenceLevel(), n.Right, buf)
	buf.WriteRune(')')
}

func (n nodeTypeSet) marshalCedar(buf *bytes.Buffer) {
	buf.WriteRune('[')
	for i := range n.Elements {
		marshalChildNode(n.precedenceLevel(), n.Elements[i], buf)
		if i != len(n.Elements)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(']')
}

func marshalInfixBinaryOp(n binaryNode, precedence nodePrecedenceLevel, op string, buf *bytes.Buffer) {
	marshalChildNode(precedence, n.Left, buf)
	buf.WriteRune(' ')
	buf.WriteString(op)
	buf.WriteRune(' ')
	marshalChildNode(precedence, n.Right, buf)
}

func (n nodeTypeMult) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "*", buf)
}

func (n nodeTypeAdd) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "+", buf)
}

func (n nodeTypeSub) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "-", buf)
}

func (n nodeTypeLessThan) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "<", buf)
}

func (n nodeTypeLessThanOrEqual) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "<=", buf)
}

func (n nodeTypeGreaterThan) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), ">", buf)
}

func (n nodeTypeGreaterThanOrEqual) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), ">=", buf)
}

func (n nodeTypeEquals) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "==", buf)
}

func (n nodeTypeNotEquals) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "!=", buf)
}

func (n nodeTypeIn) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "in", buf)
}

func (n nodeTypeAnd) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "&&", buf)
}

func (n nodeTypeOr) marshalCedar(buf *bytes.Buffer) {
	marshalInfixBinaryOp(n.binaryNode, n.precedenceLevel(), "||", buf)
}

func (n nodeTypeHas) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(" has ")
	if canMarshalAsIdent(string(n.Value)) {
		buf.WriteString(string(n.Value))
	} else {
		buf.WriteString(n.Value.Cedar())
	}
}

func (n nodeTypeIs) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.EntityType))
}

func (n nodeTypeIsIn) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Left, buf)
	buf.WriteString(" is ")
	buf.WriteString(string(n.EntityType))
	buf.WriteString(" in ")
	n.Entity.marshalCedar(buf)
}

func (n nodeTypeLike) marshalCedar(buf *bytes.Buffer) {
	marshalChildNode(n.precedenceLevel(), n.Arg, buf)
	buf.WriteString(" like ")
	n.Value.MarshalCedar(buf)
}

func (n nodeTypeIf) marshalCedar(buf *bytes.Buffer) {
	buf.WriteString("if ")
	marshalChildNode(n.precedenceLevel(), n.If, buf)
	buf.WriteString(" then ")
	marshalChildNode(n.precedenceLevel(), n.Then, buf)
	buf.WriteString(" else ")
	marshalChildNode(n.precedenceLevel(), n.Else, buf)
}
