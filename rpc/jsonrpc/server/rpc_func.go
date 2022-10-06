package server

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
)

// RegisterRPCFuncs adds a route for each function in the funcMap, as well as
// general jsonrpc and websocket handlers for all functions. "result" is the
// interface on which the result objects are registered, and is popualted with
// every RPCResponse
func RegisterRPCFuncs(mux *http.ServeMux, funcMap map[string]*RPCFunc, logger log.Logger) {
	// HTTP endpoints
	for funcName, rpcFunc := range funcMap {
		mux.HandleFunc("/"+funcName, makeHTTPHandler(rpcFunc, logger))
	}

	// JSONRPC endpoints
	mux.HandleFunc("/", handleInvalidJSONRPCPaths(makeJSONRPCHandler(funcMap, logger)))
}

// Function introspection
const (
	Cacheable = "cacheable"
	Ws        = "ws"
)

// RPCFunc contains the introspected type information for a function
type RPCFunc struct {
	f        reflect.Value          // underlying rpc function
	args     []reflect.Type         // type of each function arg
	returns  []reflect.Type         // type of each return arg
	argNames []string               // name of each argument
	opts     map[string]interface{} // options of RPCFunc
}

// NewRPCFunc wraps a function for introspection.
// f is the function, args are comma separated argument names
func NewRPCFunc(f interface{}, args string, options ...string) *RPCFunc {
	return newRPCFunc(f, args, options...)
}

// NewWSRPCFunc wraps a function for introspection and use in the websockets.
func NewWSRPCFunc(f interface{}, args string, options ...string) *RPCFunc {
	options = append(options, Ws)
	return newRPCFunc(f, args, options...)
}

func newRPCFunc(f interface{}, args string, options ...string) *RPCFunc {
	var argNames []string
	if args != "" {
		argNames = strings.Split(args, ",")
	}

	opts := make(map[string]interface{})
	for _, opt := range options {
		switch opt {
		case Ws: // Ws indicates the rpc connection is using a websocket connection
			opts[Ws] = true
		case Cacheable: // Cacheable is a bool value to allow the client proxy server to cache the RPC results
			opts[Cacheable] = true
		}
	}

	return &RPCFunc{
		f:        reflect.ValueOf(f),
		args:     funcArgTypes(f),
		returns:  funcReturnTypes(f),
		argNames: argNames,
		opts:     opts,
	}
}

// return a function's argument types
func funcArgTypes(f interface{}) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumIn()
	typez := make([]reflect.Type, n)
	for i := 0; i < n; i++ {
		typez[i] = t.In(i)
	}
	return typez
}

// return a function's return types
func funcReturnTypes(f interface{}) []reflect.Type {
	t := reflect.TypeOf(f)
	n := t.NumOut()
	typez := make([]reflect.Type, n)
	for i := 0; i < n; i++ {
		typez[i] = t.Out(i)
	}
	return typez
}

//-------------------------------------------------------------

// NOTE: assume returns is result struct and error. If error is not nil, return it
func unreflectResult(returns []reflect.Value) (interface{}, error) {
	errV := returns[1]
	if errV.Interface() != nil {
		return nil, fmt.Errorf("%v", errV.Interface())
	}
	rv := returns[0]
	// the result is a registered interface,
	// we need a pointer to it so we can marshal with type byte
	rvp := reflect.New(rv.Type())
	rvp.Elem().Set(rv)
	return rvp.Interface(), nil
}
