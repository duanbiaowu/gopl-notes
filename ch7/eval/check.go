package eval

import (
	"fmt"
	"strings"
)

// Var 和 literal 的求值不可能出错，直接返回 nil
func (v Var) Check(vars map[Var]bool) error {
	vars[v] = true
	return nil
}
func (literal) Check(vars map[Var]bool) error {
	return nil
}

// unary 和 binary 的方法，首先检查操作符是否合法，再递归地检查操作数
func (u unary) Check(vars map[Var]bool) error {
	if !strings.ContainsRune("+-", u.op) {
		return fmt.Errorf("unexpected unary op %q", u.op)
	}
	return u.x.Check(vars)
}
func (b binary) Check(vars map[Var]bool) error {
	if !strings.ContainsRune("+-*/", b.op) {
		return fmt.Errorf("unexpected binary op %q", b.op)
	}
	if err := b.x.Check(vars); err != nil {
		return err
	}
	return b.y.Check(vars)
}

// call 的方法首先检查函数是否是已知的，然后检查参数的个数，最后递归检查每个参数
var numParams = map[string]int{"pow": 2, "sin": 1, "sqrt": 1}

func (c call) Check(vars map[Var]bool) error {
	arity, ok := numParams[c.fn]
	if !ok {
		return fmt.Errorf("unknown function %q", c.fn)
	}
	if len(c.args) != arity {
		return fmt.Errorf("call to %s has %d args, want %d", c.fn, len(c.args), arity)
	}
	for _, arg := range c.args {
		if err := arg.Check(vars); err != nil {
			return err
		}
	}
	return nil
}
