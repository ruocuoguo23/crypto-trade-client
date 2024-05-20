/*
Package rpc provides access to the exported methods of an object across a
network or other I/O connection.
Only methods that satisfy these criteria will be made available for remote
access; other methods will be ignored:
  - the method's type is exported.
  - the method is exported.
  - the method can have a context argument, but it must be the first.
  - the method can have an argument other than context argument, it must be exported (or builtin) type.
  - the method can have an error return type, but it must be the last one.
  - the method can have a return type other than the return type, it must be exported (or builtin) types.

In effect, the method must look schematically like

	func (t *T) MethodName(ctx context, argType T1) (T2, error)

where T1 and T2 can be marshalled by json encoding.
*/
package rpc

import (
	"context"
	"crypto-trade-client/common/rpc/jsonrpc2"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

const (
	defaultService = "_default"
)

type ServiceRegistry struct {
	serviceMap sync.Map // map[string]*service
}

type methodType struct {
	method    reflect.Method
	ArgType   []reflect.Type
	ReplyType reflect.Type
	hasCtx    bool
	errPos    int // err return idx, of -1 when method cannot return error
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

func (registry *ServiceRegistry) Register(rcvr interface{}) error {
	return registry.register(rcvr, "", false)
}

func (registry *ServiceRegistry) RegisterName(name string, rcvr interface{}) error {
	return registry.register(rcvr, name, true)
}

func (registry *ServiceRegistry) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		return fmt.Errorf("rpc.Register: no service name for type %s", s.typ.String())
	}
	if !token.IsExported(sname) && !useName {
		return fmt.Errorf("rpc.Register: type %s is not exported", sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ)

	if len(s.method) == 0 {
		return errors.New("rpc: service has not valid methods")
	}

	if _, dup := registry.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
MethodLoop:
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		hasCtx := false
		argPos := 1

		if mtype.NumIn() >= 2 && mtype.In(1) == typeOfContext {
			argPos = 2
			hasCtx = true
		}

		argl := mtype.NumIn() - argPos
		argType := make([]reflect.Type, argl)
		for i := 0; i < argl; i++ {
			at := mtype.In(argPos + i)
			if !isExportedOrBuiltinType(at) {
				continue MethodLoop
			}
			argType[i] = at
		}

		// Method out at most 2.
		if mtype.NumOut() > 2 {
			continue
		}

		errPos := -1
		var replyType reflect.Type

		// Error must be returned, and it must be the last returned value.
		switch {
		case mtype.NumOut() == 1:
			if mtype.Out(0) != typeOfError {
				continue
			}
			errPos = 0
		case mtype.NumOut() == 2:
			if mtype.Out(1) != typeOfError {
				continue
			}
			replyType = mtype.Out(0)
			errPos = 1
		}
		methods[strings.ToLower(mname)] = &methodType{method: method, ArgType: argType, ReplyType: replyType, hasCtx: hasCtx, errPos: errPos}
	}
	return methods
}

type serviceHandler struct {
	registry *ServiceRegistry
}

func NewJsonrpc2Handler(registry *ServiceRegistry) jsonrpc2.Handler {
	return &serviceHandler{
		registry: registry,
	}
}

func stringSplitter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

func (h *serviceHandler) Handle(ctx context.Context, r *jsonrpc2.Request) (interface{}, error) {
	nsm := stringSplitter(r.Method, "._")
	var sname, mname string
	if len(nsm) == 2 {
		sname = nsm[0]
		mname = nsm[1]
	} else {
		sname = defaultService
		mname = nsm[0]
	}
	svci, ok := h.registry.serviceMap.Load(sname)
	if !ok {
		return nil, jsonrpc2.ErrNotHandled
	}
	svc := svci.(*service)
	m := svc.method[strings.ToLower(mname)]
	if m == nil {
		return nil, jsonrpc2.ErrNotHandled
	}

	args := make([]reflect.Value, 0, 4)

	args = append(args, svc.rcvr)

	if m.hasCtx {
		args = append(args, reflect.ValueOf(ctx))
	}

	if len(m.ArgType) > 0 {
		var params []json.RawMessage
		d, _ := r.Params.MarshalJSON()
		if err := json.Unmarshal(d, &params); err != nil {
			return nil, jsonrpc2.ErrParse
		}

		for i, at := range m.ArgType {
			argv := reflect.New(at)
			raw := params[i]
			d1, _ := raw.MarshalJSON()
			if err := json.Unmarshal(d1, argv.Interface()); err != nil {
				return nil, jsonrpc2.ErrParse
			}

			args = append(args, argv.Elem())
		}
	}

	return m.call(ctx, args)
}

func (m *methodType) call(ctx context.Context, argv []reflect.Value) (res interface{}, errRes error) {
	function := m.method.Func

	// Catch panic while running the callback.
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
		}
	}()

	results := function.Call(argv)
	if len(results) == 0 {
		return nil, nil
	}
	if m.errPos >= 0 && !results[m.errPos].IsNil() {
		// Method has returned non-nil error value.
		err := results[m.errPos].Interface().(error)
		return reflect.Value{}, err
	}
	return results[0].Interface(), nil
}
