// Package uel is the uncommon expression language. this is like CEL but instead of using
// protobufs it lets you use your own types, kind of like userdata in lua.
package uel

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	ErrParser       = errors.New("parser error")
	ErrUnboundVar   = errors.New("unbound variable")
	ErrUnknownOp    = errors.New("unknown op")
	ErrInvalidOp    = errors.New("invalid op")
	ErrTypeMismatch = errors.New("type mismatch")
)

func setToMap(chars string) map[rune]bool {
	rv := map[rune]bool{}
	for _, char := range chars {
		rv[char] = true
	}
	return rv
}

type Evaluable interface {
	Run(env map[any]any) (any, error)
}

var (
	identChars                   = setToMap("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789")
	disallowedStartingIdentChars = setToMap("0123456789.")
	numberChars                  = setToMap("0123456789_.")
	durationSuffixes             = []string{"ns", "us", "µs", "ms", "s", "m", "h"}
)

type Parser struct {
	source         []rune
	pos, line, col int
	currentChar    rune
}

func NewParser(source string) *Parser {
	p := &Parser{
		source:      []rune(source),
		pos:         0,
		line:        1,
		col:         1,
		currentChar: -1,
	}
	if len(p.source) > 0 {
		p.currentChar = p.source[0]
	}
	return p
}

func (p *Parser) advance(distance int) error {
	for i := 0; i < distance; i++ {
		if p.eof() {
			return errors.New("unexpected eof")
		}
		if p.currentChar == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
		if p.pos >= len(p.source) {
			p.currentChar = -1
		} else {
			p.currentChar = p.source[p.pos]
		}
	}
	return nil
}

func (p *Parser) checkpoint() (pos, col, line int) {
	return p.pos, p.col, p.line
}

func (p *Parser) restore(pos, col, line int) {
	p.pos, p.col, p.line = pos, col, line
	if p.pos >= len(p.source) {
		p.currentChar = -1
	} else {
		p.currentChar = p.source[p.pos]
	}
}

func (p *Parser) sourceRef(pos, col, line int) (_line, _col int) {
	return line, col
}

func (p *Parser) sourceError(messagef string, args ...any) error {
	message := fmt.Sprintf(messagef, args...)
	return fmt.Errorf("%w: line %d, column %d: %#v\n%s", ErrParser, p.line, p.col, message, string(debug.Stack()))
}

func (p *Parser) eof() bool {
	return p.pos >= len(p.source)
}

func (p *Parser) char(lookahead int) rune {
	if p.pos+lookahead >= len(p.source) || p.pos+lookahead < 0 {
		return -1
	}
	return p.source[p.pos+lookahead]
}

func (p *Parser) string(width int) string {
	remaining := p.source[p.pos:]
	if len(remaining) < width {
		width = len(remaining)
	}
	return string(remaining[:width])
}

func (p *Parser) skipComment() (bool, error) {
	if p.currentChar != '#' {
		return false, nil
	}
	if err := p.advance(1); err != nil {
		return false, err
	}
	for {
		if p.eof() {
			return true, nil
		}
		if err := p.advance(1); err != nil {
			return false, err
		}
		if p.currentChar == '\n' {
			return true, nil
		}
	}
}

func (p *Parser) skipWhitespace() (bool, error) {
	if p.eof() {
		return false, nil
	}
	skipped, err := p.skipComment()
	if err != nil {
		return false, err
	}
	if skipped {
		return true, nil
	}
	switch p.currentChar {
	case ' ', '\t', '\r', '\n':
		return true, p.advance(1)
	}
	return false, nil
}

func (p *Parser) skipAllWhitespace() (bool, error) {
	anySkipped := false
	for {
		skipped, err := p.skipWhitespace()
		if err != nil {
			return false, err
		}
		if !skipped {
			return anySkipped, nil
		}
		anySkipped = true
	}
}

func (p *Parser) parseIdentifier() (Evaluable, error) {
	if disallowedStartingIdentChars[p.currentChar] {
		return nil, nil
	}
	chars, err := p.parseChars(identChars)
	if err != nil {
		return nil, err
	}
	if _, err = p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	if chars == "" {
		return nil, nil
	}
	return &Ident{Name: chars}, nil
}

func (p *Parser) parseDurationSuffix() (string, error) {
	for _, suffix := range durationSuffixes {
		if p.string(len(suffix)) == suffix {
			return suffix, p.advance(len(suffix))
		}
	}
	return "", nil
}

func (p *Parser) parseValue() (Evaluable, error) {
	num, err := p.parseChars(numberChars)
	if err != nil {
		return nil, err
	}

	suffix, err := p.parseDurationSuffix()
	if err != nil {
		return nil, err
	}

	if _, err = p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	if num == "" {
		return nil, nil
	}

	if suffix != "" {
		dur, err := time.ParseDuration(num + suffix)
		if err != nil {
			return nil, err
		}
		return &Value[time.Duration]{Val: dur}, nil
	}

	if strings.Contains(num, ".") {
		val, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return nil, err
		}
		return &Value[float64]{Val: val}, nil
	}
	val, err := strconv.ParseInt(num, 0, 64)
	if err != nil {
		return nil, err
	}
	return &Value[int64]{Val: val}, nil
}

func (p *Parser) parseChars(allowed map[rune]bool) (string, error) {
	if !allowed[p.currentChar] {
		return "", nil
	}
	chars := string(p.currentChar)
	if err := p.advance(1); err != nil {
		return "", err
	}
	for allowed[p.currentChar] {
		chars += string(p.currentChar)
		if err := p.advance(1); err != nil {
			return "", err
		}
	}
	return chars, nil
}

func (p *Parser) parseString() (Evaluable, error) {
	if p.char(0) != '"' {
		return nil, nil
	}
	if err := p.advance(1); err != nil {
		return nil, err
	}
	var val []rune
	for {
		r := p.char(0)
		if err := p.advance(1); err != nil {
			return nil, err
		}
		switch r {
		case '\\':
			r = p.char(0)
			if err := p.advance(1); err != nil {
				return nil, err
			}
			switch r {
			case '\\', '"':
				val = append(val, r)
			case 'n':
				val = append(val, '\n')
			case 't':
				val = append(val, '\t')
			default:
				return nil, p.sourceError("unexpected escape code: %#v", r)
			}
		case '"':
			_, err := p.skipAllWhitespace()
			return &Value[string]{Val: string(val)}, err
		case '\n':
			return nil, p.sourceError("unexpected end of line")
		default:
			val = append(val, r)
		}
	}
}

func (p *Parser) parseLiteral() (Evaluable, error) {
	str, err := p.parseString()
	if err != nil {
		return nil, err
	}
	if str != nil {
		return str, nil
	}
	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	if ident != nil {
		return ident, nil
	}
	return p.parseValue()
}

func (p *Parser) parseFunctionCall() (Evaluable, error) {
	val, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}
	for {
		if p.eof() {
			return val, nil
		}
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		if args == nil {
			_, err = p.skipAllWhitespace()
			return val, err
		}
		val = &Call{
			Func: val,
			Args: args,
		}
	}
}

func (p *Parser) parseArgs() ([]Evaluable, error) {
	if p.char(0) != '(' {
		return nil, nil
	}
	if err := p.advance(1); err != nil {
		return nil, err
	}
	if _, err := p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	args := []Evaluable{}
	if p.char(0) == ')' {
		if err := p.advance(1); err != nil {
			return nil, err
		}
		return args, nil
	}
	arg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)
	for {
		if _, err = p.skipAllWhitespace(); err != nil {
			return nil, err
		}
		if p.char(0) == ')' {
			if err := p.advance(1); err != nil {
				return nil, err
			}
			return args, nil
		}
		if p.char(0) != ',' {
			return nil, p.sourceError("unexpected character %#v", p.char(0))
		}
		if err := p.advance(1); err != nil {
			return nil, err
		}
		if _, err = p.skipAllWhitespace(); err != nil {
			return nil, err
		}
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
}

func (p *Parser) parseSubexpression() (Evaluable, error) {
	if p.char(0) != '(' {
		return p.parseFunctionCall()
	}
	if err := p.advance(1); err != nil {
		return nil, err
	}
	if _, err := p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if _, err = p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	if p.char(0) != ')' {
		return nil, p.sourceError("subexpression ended unexpectedly, found %#v", p.char(0))
	}
	if err = p.advance(1); err != nil {
		return nil, err
	}
	_, err = p.skipAllWhitespace()
	return &Subexpression{Expr: expr}, err
}

func (p *Parser) parseExponentiation() (Evaluable, error) {
	return p.parseOperation(
		p.parseSubexpression,
		map[OpType][]string{
			OpExp: []string{"^"},
		},
	)
}

func (p *Parser) parseValNegation() (Evaluable, error) {
	return p.parseModifier(
		p.parseExponentiation,
		map[ModType][]string{
			ModNeg: []string{"-"},
		},
	)
}

func (p *Parser) parseMultiplicationDivision() (Evaluable, error) {
	return p.parseOperation(
		p.parseValNegation,
		map[OpType][]string{
			OpMul: []string{"*"},
			OpDiv: []string{"/"},
		},
	)
}

func (p *Parser) parseAdditionSubtraction() (Evaluable, error) {
	return p.parseOperation(
		p.parseMultiplicationDivision,
		map[OpType][]string{
			OpAdd: []string{"+"},
			OpSub: []string{"-"},
		},
	)
}

func (p *Parser) parseComparison() (Evaluable, error) {
	return p.parseOperation(
		p.parseAdditionSubtraction,
		map[OpType][]string{
			OpLess:         []string{"<"},
			OpLessEqual:    []string{"<="},
			OpEqual:        []string{"=="},
			OpNotEqual:     []string{"!=", "~=", "<>"},
			OpGreater:      []string{">"},
			OpGreaterEqual: []string{">="},
		},
	)
}

func (p *Parser) parseBoolNegation() (Evaluable, error) {
	return p.parseModifier(
		p.parseComparison,
		map[ModType][]string{
			ModNot: []string{"!", "not"},
		},
	)
}

func (p *Parser) parseConjunction() (Evaluable, error) {
	return p.parseOperation(
		p.parseBoolNegation,
		map[OpType][]string{
			OpAnd: []string{"&&", "and"},
		},
	)
}

func (p *Parser) parseDisjunction() (Evaluable, error) {
	return p.parseOperation(
		p.parseConjunction,
		map[OpType][]string{
			OpOr: []string{"||", "or"},
		},
	)
}

func (p *Parser) parseOperation(valueParse func() (Evaluable, error),
	opMap map[OpType][]string) (Evaluable, error) {
	val, err := valueParse()
	if err != nil {
		return nil, err
	}
	for {
		if p.eof() {
			return val, nil
		}
		cls, rhs, err := parseOpAndRHS(p, valueParse, opMap)
		if err != nil {
			return nil, err
		}
		if cls == OpOrModNil {
			return val, nil
		}
		val = &Operation{
			Type:  cls,
			Left:  val,
			Right: rhs,
		}
	}
}

func (p *Parser) parseModifier(valueParse func() (Evaluable, error),
	modMap map[ModType][]string) (Evaluable, error) {
	cls, val, err := parseOpAndRHS(p, valueParse, modMap)
	if err != nil {
		return nil, err
	}
	if cls != OpOrModNil {
		return &Modifier{
			Type: cls,
			Val:  val,
		}, nil
	}
	return valueParse()
}

func (p *Parser) isBoundary(char1, char2 rune) bool {
	return !identChars[char1] || !identChars[char2]
}

func parseOpAndRHS[T OpType | ModType](p *Parser, valueParse func() (Evaluable, error),
	opMap map[T][]string) (T, Evaluable, error) {
	cpos, ccol, cline := p.checkpoint()
	for cls, operators := range opMap {
		for _, op := range operators {
			if strings.ToLower(p.string(len(op))) == op && p.isBoundary(p.char(len(op)-1), p.char(len(op))) {
				if err := p.advance(len(op)); err != nil {
					return OpOrModNil, nil, err
				}
				if _, err := p.skipAllWhitespace(); err != nil {
					return OpOrModNil, nil, err
				}
				rhs, err := valueParse()
				if err != nil {
					return OpOrModNil, nil, err
				}
				if rhs != nil {
					return cls, rhs, nil
				}
				p.restore(cpos, ccol, cline)
			}
		}
	}
	return OpOrModNil, nil, nil
}

func (p *Parser) Parse() (Evaluable, error) {
	if _, err := p.skipAllWhitespace(); err != nil {
		return nil, err
	}
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if !p.eof() {
		return nil, p.sourceError("unparsed input")
	}
	return val, nil
}

func (p *Parser) parseExpression() (Evaluable, error) {
	return p.parseDisjunction()
}

type Subexpression struct {
	Expr Evaluable
}

func (s *Subexpression) Run(env map[any]any) (any, error) {
	return s.Expr.Run(env)
}

type Call struct {
	Func Evaluable
	Args []Evaluable
}

func (c *Call) Run(env map[any]any) (rv any, err error) {
	f, err := c.Func.Run(env)
	if err != nil {
		return nil, err
	}
	args := make([]reflect.Value, 0, len(c.Args))
	for _, arg := range c.Args {
		res, err := arg.Run(env)
		if err != nil {
			return nil, err
		}
		args = append(args, reflect.ValueOf(res))
	}
	defer func() {
		if recv := recover(); recv != nil {
			err = fmt.Errorf("%w: %v", ErrTypeMismatch, recv)
		}
	}()
	result := reflect.ValueOf(f).Call(args)
	switch len(result) {
	case 1:
		return result[0].Interface(), nil
	case 2:
		errUncasted := result[1].Interface()
		if errUncasted == nil {
			return result[0].Interface(), nil
		}
		err, ok := result[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("%w: unexpected error return value: %#v", ErrTypeMismatch, result[1].Interface())
		}
		return result[0].Interface(), err
	default:
		return nil, fmt.Errorf("%w: unexpected return values", ErrTypeMismatch)
	}
}

type Operation struct {
	Type  OpType
	Left  Evaluable
	Right Evaluable
}

func (o *Operation) Run(env map[any]any) (any, error) {
	callableUncast, ok := env[o.Type]
	if !ok {
		callableUncast, ok = defaultEnv[o.Type]
		if !ok {
			return nil, fmt.Errorf("%w: %#v", ErrUnknownOp, o.Type)
		}
	}
	callable, ok := callableUncast.(func(env map[any]any, a, b any) (any, error))
	if !ok {
		return nil, fmt.Errorf("%w: %#v", ErrInvalidOp, o.Type)
	}
	lhs, err := o.Left.Run(env)
	if err != nil {
		return nil, err
	}
	rhs, err := o.Right.Run(env)
	if err != nil {
		return nil, err
	}
	return callable(env, lhs, rhs)
}

type OpType string

const (
	OpOrModNil            = ""
	OpExp          OpType = "^"
	OpMul          OpType = "*"
	OpDiv          OpType = "/"
	OpAdd          OpType = "+"
	OpSub          OpType = "-"
	OpLess         OpType = "<"
	OpLessEqual    OpType = "<="
	OpEqual        OpType = "=="
	OpNotEqual     OpType = "!="
	OpGreater      OpType = ">"
	OpGreaterEqual OpType = ">="
	OpAnd          OpType = "&&"
	OpOr           OpType = "||"
)

type Modifier struct {
	Type ModType
	Val  Evaluable
}

func (m *Modifier) Run(env map[any]any) (any, error) {
	callableUncast, ok := env[m.Type]
	if !ok {
		callableUncast, ok = defaultEnv[m.Type]
		if !ok {
			return nil, fmt.Errorf("%w: %#v", ErrUnknownOp, m.Type)
		}
	}
	callable, ok := callableUncast.(func(env map[any]any, a any) (any, error))
	if !ok {
		return nil, fmt.Errorf("%w: %#v", ErrInvalidOp, m.Type)
	}
	val, err := m.Val.Run(env)
	if err != nil {
		return nil, err
	}
	return callable(env, val)
}

type ModType string

const (
	ModNeg ModType = "-"
	ModNot ModType = "!"
)

type Ident struct {
	Name string
}

func (i *Ident) Run(env map[any]any) (any, error) {
	if v, ok := env[i.Name]; ok {
		return v, nil
	}
	if v, ok := defaultEnv[i.Name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("%w: %#v", ErrUnboundVar, i.Name)
}

type Value[T any] struct {
	Val T
}

func (v *Value[T]) Run(env map[any]any) (any, error) {
	return v.Val, nil
}

func Parse(expression string) (Evaluable, error) {
	return NewParser(expression).Parse()
}

func Eval(expression string, env map[any]any) (any, error) {
	val, err := Parse(expression)
	if err != nil {
		return nil, err
	}
	return val.Run(env)
}
