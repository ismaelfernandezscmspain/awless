/*
Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"github.com/wallix/awless/template/driver"
	"github.com/wallix/awless/template/internal/ast"
)

type Template struct {
	ID string
	*ast.AST
}

func (s *Template) Run(env *Env) (*Template, error) {
	vars := map[string]interface{}{}

	current := &Template{AST: &ast.AST{}}
	current.ID = ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()

	for _, sts := range s.Statements {
		clone := sts.Clone()
		current.Statements = append(current.Statements, clone)
		ctx := map[string]interface{}{
			"Variables":  env.ResolvedVariables,
			"References": env.ResolvedVariables, // retro-compatibility with v0.1.2
		}
		switch n := clone.Node.(type) {
		case *ast.CommandNode:
			n.ProcessRefs(vars)
			if env.IsDryRun {
				if v, ok := n.Command.(dryRunner); ok {
					n.CmdResult, n.CmdErr = v.DryRun(ctx, n.ToDriverParams())
					n.CmdErr = prefixError(n.CmdErr, "dry run")
				}
			} else {
				n.CmdResult, n.CmdErr = n.Run(ctx, n.ToDriverParams())
				if n.CmdErr != nil {
					return current, nil
				}
			}
		case *ast.DeclarationNode:
			ident := n.Ident
			expr := n.Expr
			switch n := expr.(type) {
			case *ast.CommandNode:
				n.ProcessRefs(vars)
				if env.IsDryRun {
					if v, ok := n.Command.(dryRunner); ok {
						n.CmdResult, n.CmdErr = v.DryRun(ctx, n.ToDriverParams())
						n.CmdErr = prefixError(n.CmdErr, "dry run")
					}
				} else {
					n.CmdResult, n.CmdErr = n.Run(ctx, n.ToDriverParams())
					if n.CmdErr != nil {
						return current, nil
					}
				}
				vars[ident] = n.Result()
			default:
				return current, fmt.Errorf("unknown type of node: %T", expr)
			}
		default:
			return current, fmt.Errorf("unknown type of node: %T", clone.Node)
		}
	}

	return current, nil
}

func prefixError(err error, prefix string) error {
	if err == nil {
		return err
	}
	return fmt.Errorf("%s: %s", prefix, err.Error())
}

func (s *Template) Validate(rules ...Validator) (all []error) {
	for _, rule := range rules {
		errs := rule.Execute(s)
		all = append(all, errs...)
	}

	return
}

func (t *Template) HasErrors() bool {
	for _, cmd := range t.CommandNodesIterator() {
		if cmd.CmdErr != nil {
			return true
		}
	}
	return false
}

func (t *Template) UniqueDefinitions(fn DefinitionLookupFunc) (definitions Definitions) {
	unique := make(map[string]Definition)
	for _, cmd := range t.CommandNodesIterator() {
		key := fmt.Sprintf("%s%s", cmd.Action, cmd.Entity)
		if def, ok := fn(key); ok {
			if _, done := unique[key]; !done {
				unique[key] = def
				definitions = append(definitions, def)
			}
		}
	}

	return
}

var driverFunctionFailedErr = errors.New("Driver function call failed")

func runCmd(n *ast.CommandNode, env *Env, vars map[string]interface{}) error {
	n.ProcessRefs(vars)
	params := n.ToDriverParams()

	ctx := driver.NewContext(env.ResolvedVariables)

	if env.IsNewRunner {
		if cmd := env.Lookuper(n.Action, n.Entity); cmd != nil {
			if env.IsDryRun {
				n.CmdResult, n.CmdErr = driver.DryRun(cmd, ctx, params)
			}
			n.CmdResult, n.CmdErr = driver.Run(cmd, ctx, params)
		} else {
			return fmt.Errorf("new runner: no command for %s %s", n.Action, n.Entity)
		}
		return nil
	}

	fn, err := env.Driver.Lookup(n.Action, n.Entity)
	if err != nil {
		return err
	}
	n.CmdResult, n.CmdErr = fn(ctx, params)
	if n.CmdErr != nil {
		return driverFunctionFailedErr
	}
	return nil
}

func (s *Template) visitHoles(fn func(n ast.WithHoles)) {
	for _, n := range s.expressionNodesIterator() {
		if h, ok := n.(ast.WithHoles); ok {
			fn(h)
		}
	}
}

func (s *Template) visitCommandNodes(fn func(n *ast.CommandNode)) {
	for _, cmd := range s.CommandNodesIterator() {
		fn(cmd)
	}
}

func (s *Template) visitCommandNodesE(fn func(n *ast.CommandNode) error) error {
	for _, cmd := range s.CommandNodesIterator() {
		if err := fn(cmd); err != nil {
			return err
		}
	}

	return nil
}

func (s *Template) visitCommandDeclarationNodes(fn func(n *ast.DeclarationNode)) {
	for _, cmd := range s.commandDeclarationNodesIterator() {
		fn(cmd)
	}
}

func (s *Template) visitDeclarationNodes(fn func(n *ast.DeclarationNode)) {
	for _, dcl := range s.declarationNodesIterator() {
		fn(dcl)
	}
}

func (s *Template) CommandNodesIterator() (nodes []*ast.CommandNode) {
	for _, sts := range s.Statements {
		switch nn := sts.Node.(type) {
		case *ast.CommandNode:
			nodes = append(nodes, nn)
		case *ast.DeclarationNode:
			expr := sts.Node.(*ast.DeclarationNode).Expr
			switch expr.(type) {
			case *ast.CommandNode:
				nodes = append(nodes, expr.(*ast.CommandNode))
			}
		}
	}
	return
}

func (s *Template) WithRefsIterator() (nodes []ast.WithRefs) {
	for _, sts := range s.Statements {
		switch nn := sts.Node.(type) {
		case ast.WithRefs:
			nodes = append(nodes, nn)
		case *ast.DeclarationNode:
			expr := sts.Node.(*ast.DeclarationNode).Expr
			switch nnn := expr.(type) {
			case *ast.CommandNode:
				nodes = append(nodes, nnn)
			}
		}
	}
	return
}

func (s *Template) CommandNodesReverseIterator() (nodes []*ast.CommandNode) {
	for i := len(s.Statements) - 1; i >= 0; i-- {
		sts := s.Statements[i]
		switch sts.Node.(type) {
		case *ast.CommandNode:
			nodes = append(nodes, sts.Node.(*ast.CommandNode))
		case *ast.DeclarationNode:
			expr := sts.Node.(*ast.DeclarationNode).Expr
			switch expr.(type) {
			case *ast.CommandNode:
				nodes = append(nodes, expr.(*ast.CommandNode))
			}
		}
	}
	return
}

func (s *Template) commandDeclarationNodesIterator() (nodes []*ast.DeclarationNode) {
	for _, node := range s.declarationNodesIterator() {
		expr := node.Expr
		switch expr.(type) {
		case *ast.CommandNode:
			nodes = append(nodes, node)
		}
	}
	return
}

func (s *Template) declarationNodesIterator() (nodes []*ast.DeclarationNode) {
	for _, sts := range s.Statements {
		switch n := sts.Node.(type) {
		case *ast.DeclarationNode:
			nodes = append(nodes, n)
		}
	}
	return
}

func (s *Template) expressionNodesIterator() (nodes []ast.ExpressionNode) {
	for _, st := range s.Statements {
		if expr := extractExpressionNode(st); expr != nil {
			nodes = append(nodes, expr)
		}
	}
	return
}

func extractExpressionNode(st *ast.Statement) ast.ExpressionNode {
	switch n := st.Node.(type) {
	case *ast.DeclarationNode:
		return n.Expr
	case ast.ExpressionNode:
		return n
	}
	return nil
}

type Errors struct {
	errs []error
}

func (d *Errors) Errors() ([]error, bool) {
	return d.errs, len(d.errs) > 0
}

func (d *Errors) add(err error) {
	d.errs = append(d.errs, err)
}

func (d *Errors) Error() string {
	var all []string
	for _, err := range d.errs {
		all = append(all, err.Error())
	}
	return strings.Join(all, "\n")
}

func MatchStringParamValue(s string) bool {
	return ast.SimpleStringValue.MatchString(s)
}

type dryRunner interface {
	DryRun(ctx, params map[string]interface{}) (interface{}, error)
}
