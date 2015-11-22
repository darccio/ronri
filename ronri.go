package ronri

import (
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
)

func Eval(expression string, context *Context) (result bool, err error) {
	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return
	}
	expr = removeStackedParens(expr)
	switch t := expr.(type) {
	case *ast.BinaryExpr:
		result, err = eval(expr, context)
	case *ast.UnaryExpr:
		result, err = eval(expr, context)
	case *ast.Ident:
		result, err = eval(expr, context)
	// TODO case *ast.CallExpr
	default:
		err = fmt.Errorf("only conditional expressions are supported, got %T", t)
	}
	return
}

func eval(expr ast.Expr, context *Context) (result bool, err error) {
	switch t := expr.(type) {
	case *ast.ParenExpr:
		x := removeStackedParens(t.X)
		result, err = eval(x, context)
	case *ast.BinaryExpr:
		result, err = resolveBinaryExpr(t, context)
	case *ast.UnaryExpr:
		result, err = resolveUnaryExpr(t, context)
	case *ast.Ident:
		result, err = single(resolveIdent(t, context))
	case *ast.BasicLit:
		result, err = single(resolveLiteral(t, context))
	default:
		err = fmt.Errorf("unexpected %T", t)
	}
	return
}

func resolve(expr ast.Expr, context *Context) (result interface{}, err error) {
	switch t := expr.(type) {
	case *ast.ParenExpr:
		x := removeStackedParens(t.X)
		result, err = resolve(x, context)
	case *ast.BinaryExpr:
		result, err = resolveBinaryExpr(t, context)
	case *ast.UnaryExpr:
		result, err = resolveUnaryExpr(t, context)
	case *ast.Ident:
		result, err = resolveIdent(t, context)
	case *ast.BasicLit:
		result, err = resolveLiteral(t, context)
	default:
		err = fmt.Errorf("unexpected %T", t)
	}
	return
}

func removeStackedParens(expr ast.Expr) (result ast.Expr) {
	result = expr
	parenExpr, ok := expr.(*ast.ParenExpr)
	for ok {
		result = parenExpr.X
		parenExpr, ok = result.(*ast.ParenExpr)
	}
	return
}

func single(value interface{}, verr error) (result bool, err error) {
	if verr != nil {
		err = verr
		return
	}
	if b, ok := value.(bool); ok {
		result = b
	} else {
		err = fmt.Errorf("non-bool (type %T) used as if condition", value)
	}
	return
}

func resolveLiteral(expr *ast.BasicLit, context *Context) (result interface{}, err error) {
	switch expr.Kind {
	case token.INT:
		result, err = strconv.ParseInt(expr.Value, 0, 64)
	case token.FLOAT:
		result, err = strconv.ParseFloat(expr.Value, 64)
	case token.CHAR:
		result = expr.Value[1]
	case token.STRING:
		value := expr.Value
		result = value[1 : len(value)-1]
	default:
		err = fmt.Errorf("literal type %s not implemented", expr.Kind)
	}
	return
}

func resolveIdent(expr *ast.Ident, context *Context) (result interface{}, err error) {
	name := expr.Name
	switch name {
	case "_":
		err = fmt.Errorf("cannot use _ as value")
		return
	case "true", "false":
		result, err = strconv.ParseBool(name)
		if err != nil {
			return
		}
	case "nil":
		result = nil
	default:
		value, ok := context.Get(name)
		if !ok {
			err = fmt.Errorf("undefined: %s", name)
			return
		}
		result = value
	}
	return
}

func resolveUnaryExpr(expr *ast.UnaryExpr, context *Context) (result bool, err error) {
	switch expr.Op {
	case token.NOT:
		result, err = eval(expr.X, context)
		result = !result
	default:
		err = fmt.Errorf("unary operator %s not implemented", expr.Op)
	}
	return
}

func resolveBinaryExpr(expr *ast.BinaryExpr, context *Context) (result bool, err error) {
	switch expr.Op {
	case token.LAND, token.LOR:
		result, err = resolveBinaryLogicalOp(expr, context)
	case token.EQL, token.NEQ:
		result, err = resolveComparable(expr, context)
	case token.LSS, token.LEQ, token.GTR, token.GEQ:
		result, err = resolveOrdered(expr, context)
	default:
		err = fmt.Errorf("binary operator %s not implemented", expr.Op)
	}
	return
}

func resolveBinaryLogicalOp(expr *ast.BinaryExpr, context *Context) (result bool, err error) {
	var ry bool
	rx, err := eval(expr.X, context)
	if err != nil {
		return
	}
	result = rx
	// TODO short-circuit evaluation https://en.wikipedia.org/wiki/Short-circuit_evaluation
	switch expr.Op {
	case token.LAND:
		ry, err = eval(expr.Y, context)
		if err != nil {
			return
		}
		result = rx && ry
	case token.LOR:
		ry, err = eval(expr.Y, context)
		if err != nil {
			return
		}
		result = rx || ry
	default:
		err = fmt.Errorf("binary operator %s not implemented", expr.Op)
	}
	return
}

func resolveComparable(expr *ast.BinaryExpr, context *Context) (result bool, err error) {
	rx, err := resolve(expr.X, context)
	if err != nil {
		return
	}
	ry, err := resolve(expr.Y, context)
	if err != nil {
		return
	}
	rxv := reflect.ValueOf(rx)
	ryv := reflect.ValueOf(ry)
	if !matchTypes(rxv, ryv) {
		err = fmt.Errorf("invalid operation: mismatched types %T and %T", rx, ry)
		return
	}
	switch rx.(type) {
	case int, int8, int16, int32, int64: // rune == int32
		result, err = compareInts(expr.Op, rxv, ryv)
	case uint, uint8, uint16, uint32, uint64:
		result, err = compareUints(expr.Op, rxv, ryv)
	case float32, float64:
		result, err = compareFloats(expr.Op, rxv, ryv)
	case string:
		result, err = compareStrings(expr.Op, rx, ry)
	default:
		result, err = compareInterfaces(expr.Op, rx, ry)
	}
	return
}

func resolveOrdered(expr *ast.BinaryExpr, context *Context) (result bool, err error) {
	rx, err := resolve(expr.X, context)
	if err != nil {
		return
	}
	ry, err := resolve(expr.Y, context)
	if err != nil {
		return
	}
	rxv := reflect.ValueOf(rx)
	ryv := reflect.ValueOf(ry)
	if !matchTypes(rxv, ryv) {
		err = fmt.Errorf("invalid operation: mismatched types %T and %T", rx, ry)
		return
	}
	if !ordered(rxv) {
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", expr.Op, rx)
		return
	}
	if !ordered(ryv) {
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", expr.Op, ry)
		return
	}
	switch rx.(type) {
	case int, int8, int16, int32, int64: // rune == int32
		result, err = compareInts(expr.Op, rxv, ryv)
	case uint, uint8, uint16, uint32, uint64:
		result, err = compareUints(expr.Op, rxv, ryv)
	case float32, float64:
		result, err = compareFloats(expr.Op, rxv, ryv)
	case string:
		result, err = compareStrings(expr.Op, rx, ry)
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", expr.Op, rx)
	}
	return
}

func matchTypes(x, y reflect.Value) (result bool) {
	result = x.Type() == y.Type()
	if result {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			result = false
		}
	}()
	result = true
	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // rune == int32
		y.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		y.Uint()
	case reflect.Float32, reflect.Float64:
		y.Float()
	default:
		result = false
	}
	return
}

func ordered(x reflect.Value) (result bool) {
	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result = true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = true
	case reflect.Float32, reflect.Float64:
		result = true
	case reflect.String:
		result = true
	}
	return
}

func compareInts(op token.Token, x, y reflect.Value) (result bool, err error) {
	sx := x.Int()
	sy := y.Int()
	switch op {
	case token.EQL:
		result = sx == sy
	case token.NEQ:
		result = sx != sy
	case token.LSS:
		result = sx < sy
	case token.LEQ:
		result = sx <= sy
	case token.GTR:
		result = sx > sy
	case token.GEQ:
		result = sx >= sy
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", op, sx)
	}
	return
}

func compareUints(op token.Token, x, y reflect.Value) (result bool, err error) {
	sx := x.Uint()
	sy := y.Uint()
	switch op {
	case token.EQL:
		result = sx == sy
	case token.NEQ:
		result = sx != sy
	case token.LSS:
		result = sx < sy
	case token.LEQ:
		result = sx <= sy
	case token.GTR:
		result = sx > sy
	case token.GEQ:
		result = sx >= sy
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", op, sx)
	}
	return
}

func compareFloats(op token.Token, x, y reflect.Value) (result bool, err error) {
	sx := x.Float()
	sy := y.Float()
	switch op {
	case token.EQL:
		result = sx == sy
	case token.NEQ:
		result = sx != sy
	case token.LSS:
		result = sx < sy
	case token.LEQ:
		result = sx <= sy
	case token.GTR:
		result = sx > sy
	case token.GEQ:
		result = sx >= sy
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", op, sx)
	}
	return
}

func compareStrings(op token.Token, x, y interface{}) (result bool, err error) {
	sx := x.(string)
	sy := y.(string)
	switch op {
	case token.EQL:
		result = sx == sy
	case token.NEQ:
		result = sx != sy
	case token.LSS:
		result = sx < sy
	case token.LEQ:
		result = sx <= sy
	case token.GTR:
		result = sx > sy
	case token.GEQ:
		result = sx >= sy
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", op, sx)
	}
	return
}

func compareInterfaces(op token.Token, x, y interface{}) (result bool, err error) {
	switch op {
	case token.EQL:
		result = x == y
	case token.NEQ:
		result = x != y
	default:
		err = fmt.Errorf("invalid operation: operator %s not defined on %T", op, x)
	}
	return
}
