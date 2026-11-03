package parser

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/extensions"
	"github.com/cedar-policy/cedar-go/internal/scanner"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func rustMarshalCedar(p *Policy) []byte {
	var buf bytes.Buffer
	annotations := slices.Clone(p.Annotations)
	slices.SortStableFunc(annotations, func(a, b ast.AnnotationType) int {
		return strings.Compare(string(a.Key), string(b.Key))
	})
	for _, a := range annotations {
		buf.WriteRune('@')
		buf.WriteString(string(a.Key))
		buf.WriteRune('(')
		buf.Write(a.Value.MarshalCedar())
		buf.WriteString(")\n")
	}
	if p.Effect == ast.EffectPermit {
		buf.WriteString("permit(")
	} else {
		buf.WriteString("forbid(")
	}
	rustMarshalScope(consts.Principal, p.Principal, &buf)
	buf.WriteString(", ")
	rustMarshalScope(consts.Action, p.Action, &buf)
	buf.WriteString(", ")
	rustMarshalScope(consts.Resource, p.Resource, &buf)
	buf.WriteRune(')')
	if cond, ok := rustFoldConditions(p.Conditions); ok {
		buf.WriteString(" when { ")
		rustMarshalExpr(cond, &buf)
		buf.WriteString(" }")
	}
	buf.WriteRune(';')
	return buf.Bytes()
}

func rustFoldConditions(conds []ast.ConditionType) (ast.IsNode, bool) {
	if len(conds) == 0 {
		return nil, false
	}
	var acc ast.IsNode
	for i := len(conds) - 1; i >= 0; i-- {
		n := conds[i].Body
		if conds[i].Condition == ast.ConditionUnless {
			n = rustNot(n)
		}
		if acc == nil {
			acc = n
		} else {
			acc = ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{Left: n, Right: acc}}
		}
	}
	return acc, true
}

func rustMarshalScope(name string, scope ast.IsScopeNode, buf *bytes.Buffer) {
	buf.WriteString(name)
	switch t := scope.(type) {
	case ast.ScopeTypeAll:
	case ast.ScopeTypeEq:
		buf.WriteString(" == ")
		buf.Write(t.Entity.MarshalCedar())
	case ast.ScopeTypeIn:
		buf.WriteString(" in ")
		buf.Write(t.Entity.MarshalCedar())
	case ast.ScopeTypeInSet:
		if len(t.Entities) == 1 {
			buf.WriteString(" in ")
			buf.Write(t.Entities[0].MarshalCedar())
			return
		}
		buf.WriteString(" in [")
		for i, e := range t.Entities {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.Write(e.MarshalCedar())
		}
		buf.WriteRune(']')
	case ast.ScopeTypeIs:
		buf.WriteString(" is ")
		buf.WriteString(string(t.Type))
	case ast.ScopeTypeIsIn:
		buf.WriteString(" is ")
		buf.WriteString(string(t.Type))
		buf.WriteString(" in ")
		buf.Write(t.Entity.MarshalCedar())
	default:
		panic(fmt.Sprintf("unknown scope type %T", t))
	}
}

func rustMarshalValue(v types.Value, buf *bytes.Buffer) {
	switch t := v.(type) {
	case types.Long:
		if t < 0 {
			buf.WriteRune('(')
			buf.WriteString(strconv.FormatInt(int64(t), 10))
			buf.WriteRune(')')
		} else {
			buf.WriteString(strconv.FormatInt(int64(t), 10))
		}
	case types.Set:
		buf.WriteRune('[')
		i := 0
		for e := range t.All() {
			if i > 0 {
				buf.WriteString(", ")
			}
			rustMarshalValue(e, buf)
			i++
		}
		buf.WriteRune(']')
	case types.Record:
		buf.WriteRune('{')
		for i, k := range slices.Sorted(t.Keys()) {
			if i > 0 {
				buf.WriteString(", ")
			}
			rustMarshalAttr(k, buf)
			buf.WriteString(": ")
			e, _ := t.Get(k)
			rustMarshalValue(e, buf)
		}
		buf.WriteRune('}')
	default:
		buf.Write(v.MarshalCedar())
	}
}

func rustMarshalAttr(attr types.String, buf *bytes.Buffer) {
	if scanner.IsIdent(string(attr)) {
		buf.WriteString(string(attr))
	} else {
		buf.Write(attr.MarshalCedar())
	}
}

func rustBoolLiteral(n ast.IsNode) (bool, bool) {
	if v, ok := n.(ast.NodeValue); ok {
		b, ok := v.Value.(types.Boolean)
		return bool(b), ok
	}
	if t, ok := n.(ast.NodeTypeAnd); ok {
		l, lok := rustBoolLiteral(t.Left)
		r, rok := rustBoolLiteral(t.Right)
		return l && r, lok && rok
	}
	if t, ok := n.(ast.NodeTypeOr); ok {
		l, lok := rustBoolLiteral(t.Left)
		r, rok := rustBoolLiteral(t.Right)
		return l || r, lok && rok
	}
	return false, false
}

func rustNot(n ast.IsNode) ast.IsNode {
	return ast.NodeTypeNot{UnaryNode: ast.UnaryNode{Arg: n}}
}

func rustMarshalOperand(n ast.IsNode, buf *bytes.Buffer) {
	if b, ok := rustBoolLiteral(n); ok {
		rustMarshalValue(types.Boolean(b), buf)
		return
	}
	switch n.(type) {
	case ast.NodeValue, ast.NodeTypeVariable, ast.NodeTypeSet, ast.NodeTypeRecord:
		rustMarshalExpr(n, buf)
	case ast.NodeTypeNot, ast.NodeTypeNegate,
		ast.NodeTypeEquals, ast.NodeTypeNotEquals, ast.NodeTypeIn,
		ast.NodeTypeLessThan, ast.NodeTypeLessThanOrEqual, ast.NodeTypeGreaterThan, ast.NodeTypeGreaterThanOrEqual,
		ast.NodeTypeAnd, ast.NodeTypeOr, ast.NodeTypeAdd, ast.NodeTypeSub, ast.NodeTypeMult,
		ast.NodeTypeContains, ast.NodeTypeContainsAll, ast.NodeTypeContainsAny, ast.NodeTypeIsEmpty,
		ast.NodeTypeAccess, ast.NodeTypeHas, ast.NodeTypeGetTag, ast.NodeTypeHasTag,
		ast.NodeTypeLike, ast.NodeTypeIs, ast.NodeTypeIsIn, ast.NodeTypeIfThenElse, ast.NodeTypeExtensionCall:
		buf.WriteRune('(')
		rustMarshalExpr(n, buf)
		buf.WriteRune(')')
	}
}

func rustMarshalInfix(n ast.BinaryNode, op string, buf *bytes.Buffer) {
	rustMarshalOperand(n.Left, buf)
	buf.WriteRune(' ')
	buf.WriteString(op)
	buf.WriteRune(' ')
	rustMarshalOperand(n.Right, buf)
}

func rustMarshalMethod(n ast.BinaryNode, method string, buf *bytes.Buffer) {
	rustMarshalOperand(n.Left, buf)
	buf.WriteRune('.')
	buf.WriteString(method)
	buf.WriteRune('(')
	rustMarshalExpr(n.Right, buf)
	buf.WriteRune(')')
}

func rustMarshalList(nodes []ast.IsNode, buf *bytes.Buffer) {
	for i, n := range nodes {
		if i > 0 {
			buf.WriteString(", ")
		}
		rustMarshalExpr(n, buf)
	}
}

func rustMarshalExpr(n ast.IsNode, buf *bytes.Buffer) {
	switch t := n.(type) {
	case ast.NodeValue:
		rustMarshalValue(t.Value, buf)
	case ast.NodeTypeVariable:
		buf.WriteString(string(t.Name))
	case ast.NodeTypeNot:
		buf.WriteRune('!')
		rustMarshalOperand(t.Arg, buf)
	case ast.NodeTypeNegate:
		buf.WriteString("-(")
		rustMarshalExpr(t.Arg, buf)
		buf.WriteRune(')')
	case ast.NodeTypeEquals:
		rustMarshalInfix(t.BinaryNode, "==", buf)
	case ast.NodeTypeNotEquals:
		rustMarshalExpr(rustNot(ast.NodeTypeEquals(t)), buf)
	case ast.NodeTypeIn:
		rustMarshalInfix(t.BinaryNode, "in", buf)
	case ast.NodeTypeLessThan:
		rustMarshalInfix(t.BinaryNode, "<", buf)
	case ast.NodeTypeLessThanOrEqual:
		rustMarshalInfix(t.BinaryNode, "<=", buf)
	case ast.NodeTypeGreaterThan:
		rustMarshalExpr(rustNot(ast.NodeTypeLessThanOrEqual(t)), buf)
	case ast.NodeTypeGreaterThanOrEqual:
		rustMarshalExpr(rustNot(ast.NodeTypeLessThan(t)), buf)
	case ast.NodeTypeAnd:
		if b, ok := rustBoolLiteral(t); ok {
			rustMarshalValue(types.Boolean(b), buf)
			return
		}
		rustMarshalInfix(t.BinaryNode, "&&", buf)
	case ast.NodeTypeOr:
		if b, ok := rustBoolLiteral(t); ok {
			rustMarshalValue(types.Boolean(b), buf)
			return
		}
		rustMarshalInfix(t.BinaryNode, "||", buf)
	case ast.NodeTypeAdd:
		rustMarshalInfix(t.BinaryNode, "+", buf)
	case ast.NodeTypeSub:
		rustMarshalInfix(t.BinaryNode, "-", buf)
	case ast.NodeTypeMult:
		rustMarshalInfix(t.BinaryNode, "*", buf)
	case ast.NodeTypeContains:
		rustMarshalMethod(t.BinaryNode, "contains", buf)
	case ast.NodeTypeContainsAll:
		rustMarshalMethod(t.BinaryNode, "containsAll", buf)
	case ast.NodeTypeContainsAny:
		rustMarshalMethod(t.BinaryNode, "containsAny", buf)
	case ast.NodeTypeGetTag:
		rustMarshalMethod(t.BinaryNode, "getTag", buf)
	case ast.NodeTypeHasTag:
		rustMarshalMethod(t.BinaryNode, "hasTag", buf)
	case ast.NodeTypeIsEmpty:
		rustMarshalOperand(t.Arg, buf)
		buf.WriteString(".isEmpty()")
	case ast.NodeTypeAccess:
		rustMarshalOperand(t.Arg, buf)
		if scanner.IsIdent(string(t.Value)) {
			buf.WriteRune('.')
			buf.WriteString(string(t.Value))
		} else {
			buf.WriteRune('[')
			buf.Write(t.Value.MarshalCedar())
			buf.WriteRune(']')
		}
	case ast.NodeTypeHas:
		rustMarshalOperand(t.Arg, buf)
		buf.WriteString(" has ")
		rustMarshalAttr(t.Value, buf)
	case ast.NodeTypeLike:
		rustMarshalOperand(t.Arg, buf)
		buf.WriteString(" like ")
		buf.Write(t.Value.MarshalCedar())
	case ast.NodeTypeIs:
		rustMarshalOperand(t.Left, buf)
		buf.WriteString(" is ")
		buf.WriteString(string(t.EntityType))
	case ast.NodeTypeIsIn:
		rustMarshalExpr(ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeTypeIs{Left: t.Left, EntityType: t.EntityType},
			Right: ast.NodeTypeIn{BinaryNode: ast.BinaryNode{Left: t.Left, Right: t.Entity}},
		}}, buf)
	case ast.NodeTypeIfThenElse:
		buf.WriteString("if ")
		rustMarshalOperand(t.If, buf)
		buf.WriteString(" then ")
		rustMarshalOperand(t.Then, buf)
		buf.WriteString(" else ")
		rustMarshalOperand(t.Else, buf)
	case ast.NodeTypeSet:
		buf.WriteRune('[')
		rustMarshalList(t.Elements, buf)
		buf.WriteRune(']')
	case ast.NodeTypeRecord:
		elements := slices.Clone(t.Elements)
		slices.SortStableFunc(elements, func(a, b ast.RecordElementNode) int {
			return strings.Compare(string(a.Key), string(b.Key))
		})
		buf.WriteRune('{')
		for i, e := range elements {
			if i > 0 {
				buf.WriteString(", ")
			}
			rustMarshalAttr(e.Key, buf)
			buf.WriteString(": ")
			rustMarshalExpr(e.Value, buf)
		}
		buf.WriteRune('}')
	case ast.NodeTypeExtensionCall:
		if extensions.ExtMap[t.Name].IsMethod && len(t.Args) > 0 {
			rustMarshalOperand(t.Args[0], buf)
			buf.WriteRune('.')
			buf.WriteString(string(t.Name))
			buf.WriteRune('(')
			rustMarshalList(t.Args[1:], buf)
			buf.WriteRune(')')
		} else {
			buf.WriteString(string(t.Name))
			buf.WriteRune('(')
			rustMarshalList(t.Args, buf)
			buf.WriteRune(')')
		}
	default:
		panic(fmt.Sprintf("unknown node type %T", t))
	}
}
