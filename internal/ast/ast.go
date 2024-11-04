package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nobletk/json-parser/internal/token"
)

type Node interface {
	TokenLiteral() string
	String() string
	ToInterface() interface{}
}

type Element interface {
	Node
	elementNode()
}

type JSONFile struct {
	Elements []Element
}

func (jf *JSONFile) TokenLiteral() string {
	if len(jf.Elements) > 0 {
		return jf.Elements[0].TokenLiteral()
	}
	return ""
}
func (jf *JSONFile) String() string {
	var out bytes.Buffer

	for _, elem := range jf.Elements {
		out.WriteString(elem.String())
	}

	return out.String()
}
func (jf *JSONFile) ToInterface() interface{} {
	return jf.Elements[0].ToInterface()
}

type Object struct {
	Token token.Token
	Pairs map[Element]Element
}

func (o *Object) elementNode()         {}
func (o *Object) TokenLiteral() string { return o.Token.Literal }
func (o *Object) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range o.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
func (o *Object) ToInterface() interface{} {
	out := make(map[string]interface{})
	for k, v := range o.Pairs {
		keyStr, ok := k.(*StringLiteral)
		if !ok {
			continue
		}
		out[keyStr.Value] = v.ToInterface()
	}
	return out
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Element
}

func (al *ArrayLiteral) elementNode()         {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (al *ArrayLiteral) ToInterface() interface{} {
	elements := []interface{}{}
	for _, elm := range al.Elements {
		elements = append(elements, elm.ToInterface())
	}
	return elements
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) elementNode()         {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", sl.Value)
}
func (sl *StringLiteral) ToInterface() interface{} { return sl.Value }

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) elementNode()             {}
func (b *Boolean) TokenLiteral() string     { return b.Token.Literal }
func (b *Boolean) String() string           { return b.Token.Literal }
func (b *Boolean) ToInterface() interface{} { return b.Value }

type Null struct {
	Token token.Token
	Value string
}

func (n *Null) elementNode()             {}
func (n *Null) TokenLiteral() string     { return n.Token.Literal }
func (n *Null) String() string           { return n.Token.Literal }
func (n *Null) ToInterface() interface{} { return nil }

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (nl *NumberLiteral) elementNode()             {}
func (nl *NumberLiteral) TokenLiteral() string     { return nl.Token.Literal }
func (nl *NumberLiteral) String() string           { return nl.Token.Literal }
func (nl *NumberLiteral) ToInterface() interface{} { return nl.Value }

type CommaLiteral struct {
	Token token.Token
	Value string
}

func (cl *CommaLiteral) elementNode()         {}
func (cl *CommaLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CommaLiteral) String() string       { return cl.Token.Literal }
