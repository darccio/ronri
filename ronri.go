package ronri

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go/ast"
	"go/parser"
	"go/token"
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
	case *ast.Ident:
		result, err = eval(expr, context)
	default:
		err = fmt.Errorf("only conditional expressions are supported, got %T", t)
	}
	return
}

func eval(expr ast.Expr, context *Context) (result bool, err error) {
	fmt.Println(spew.Sdump(expr))
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

func removeStackedParens(expr ast.Expr) (result ast.Expr) {
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
		err = fmt.Errorf("non-bool %s (type %T) used as if condition", b)
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
		result = expr.Value
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
	rx, err := eval(expr.X, context)
	if err != nil {
		return
	}
	ry, err := eval(expr.Y, context)
	if err != nil {
		return
	}
	switch expr.Op {
	case token.LAND:
		result = rx && ry
	case token.LOR:
		result = rx || ry
	case token.EQL:
		result = rx == ry
	case token.NEQ:
		result = rx != ry
	case token.LSS:
		result = rx < ry
	case token.LEQ:
		result = rx <= ry
	case token.GTR:
		result = rx > ry
	case token.GEQ:
		result = rx <= ry
	default:
		err = fmt.Errorf("binary operator %s not implemented", expr.Op)
	}
	return
}
