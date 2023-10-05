package api

import (
	"strconv"

	"github.com/jmespath-community/go-jmespath/pkg/binding"
	"github.com/jmespath-community/go-jmespath/pkg/functions"
	"github.com/jmespath-community/go-jmespath/pkg/interpreter"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
)

// JMESPath is the representation of a compiled JMES path query. A JMESPath is
// safe for concurrent use by multiple goroutines.
type JMESPath interface {
	Search(interface{}) (interface{}, error)
	SearchWithParams(interface{}, map[string]interface{}) (interface{}, error)
}

type jmesPath struct {
	node   parsing.ASTNode
	caller interpreter.FunctionCaller
}

func newJMESPath(node parsing.ASTNode, funcs ...functions.FunctionEntry) JMESPath {
	return jmesPath{
		node:   node,
		caller: interpreter.NewFunctionCaller(funcs...),
	}
}

// Compile parses a JMESPath expression and returns, if successful, a JMESPath
// object that can be used to match against data.
func Compile(expression string, funcs ...functions.FunctionEntry) (JMESPath, error) {
	parser := parsing.NewParser()
	ast, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	var f []functions.FunctionEntry
	f = append(f, functions.GetDefaultFunctions()...)
	f = append(f, funcs...)
	return newJMESPath(ast, f...), nil
}

// MustCompile is like Compile but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled
// JMESPaths.
func MustCompile(expression string, funcs ...functions.FunctionEntry) JMESPath {
	jmespath, err := Compile(expression, funcs...)
	if err != nil {
		panic(`jmespath: Compile(` + strconv.Quote(expression) + `): ` + err.Error())
	}
	return jmespath
}

// Search evaluates a JMESPath expression against input data and returns the result.
func (jp jmesPath) Search(data interface{}) (interface{}, error) {
	return jp.SearchWithParams(data, make(map[string]interface{}))
}
func (jp jmesPath) SearchWithParams(data interface{}, params map[string]interface{}) (interface{}, error) {
	bindings := binding.NewBindings()
	if params != nil {
		for key, val := range params {
			bindings = bindings.Register("$" + key, binding.NewBinding(val))
		}
	}
	intr := interpreter.NewInterpreter(data, jp.caller, bindings)
	return intr.Execute(jp.node, data)
}

// Search evaluates a JMESPath expression against input data and returns the result.
func Search(expression string, data interface{}, funcs ...functions.FunctionEntry) (interface{}, error) {
	return SearchWithParams(expression, data, make(map[string]interface{}), funcs...)
}
func SearchWithParams(expression string, data interface{}, params map[string]interface{}, funcs ...functions.FunctionEntry) (interface{}, error) {
	compiled, err := Compile(expression, funcs...)
	if err != nil {
		return nil, err
	}
	return compiled.SearchWithParams(data, params)
}
