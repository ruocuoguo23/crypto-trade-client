package rpc

import (
	"context"
	"crypto-trade-client/common/cache"
	"crypto-trade-client/common/rpc/jsonrpc2"
	"crypto-trade-client/common/tag"
	"crypto-trade-client/common/web"
	"crypto-trade-client/common/web/fetch"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"go/token"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

// ErrClient is an error which occurred on the client side the library
type ErrClient struct {
	err error
}

func (e *ErrClient) Error() string {
	return fmt.Sprintf("RPC client error: %s", e.err)
}

// Unwrap unwraps the actual error
func (e *ErrClient) Unwrap(err error) error {
	return e.err
}

// ClientCloser is used to close Client from further use
type ClientCloser func()

type Client struct {
	fetch        fetch.ClientInterface
	idCtr        int64
	cacheStorage *cache.Cache
	log          hclog.Logger
}

// NewClient creates new jsonrpc 2.0 client
//
// The parameter `nsPrefix` is deprecated.
// handler must be pointer to a struct with function fields
// Returned value closes the client connection
func NewClient(ctx context.Context, endpoint, clientName string, handler interface{}, headers map[string]string) error {
	l := hclog.L().Named("jsonrpc2." + clientName)

	fetchClient := fetch.NewClientWithEndpoint(endpoint, l).AddHeaders(headers)
	client := &Client{log: l, fetch: fetchClient, cacheStorage: cache.NewCache()}

	return client.provide(handler)
}

func NewClientWithCustomFetch(ctx context.Context, fetchClient fetch.ClientInterface, clientName string, handler interface{}) error {
	l := hclog.L().Named("jsonrpc2." + clientName)

	client := &Client{log: l, fetch: fetchClient, cacheStorage: cache.NewCache()}

	return client.provide(handler)
}

func (c *Client) provide(handler interface{}) error {
	htype := reflect.TypeOf(handler)
	if htype.Kind() != reflect.Ptr {
		return errors.New("handler must be a pointer")
	}
	etype := htype.Elem()
	if etype.Kind() != reflect.Struct {
		return errors.New("handler must point to a struct")
	}

	val := reflect.ValueOf(handler)

	// register custom http fetch client middlewares
	if h, ok := handler.(middleware); ok {
		switch f := c.fetch.(type) {
		case *fetch.Client:
			for _, m := range h.BeforeRequest() {
				f.OnBeforeRequest(m)
			}
		case *fetch.RetryableClient:
			for _, m := range h.BeforeRequest() {
				f.OnBeforeRequest(m)
			}
		}
	}

	convention := Original
	var namePrefix string
	if h, ok := handler.(methodName); ok {
		convention = h.MethodNamingConvention()
	}
	if h, ok := handler.(namespace); ok {
		namePrefix = h.Namespace() + h.NamespaceSeparator()
	}

	for i := 0; i < etype.NumField(); i++ {
		fn, err := c.makeRpcFunc(etype.Field(i), convention, namePrefix)
		if err != nil {
			c.log.Warn("unsupported method type", "err", err)
			return err
		}
		val.Elem().Field(i).Set(fn)
	}

	return nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

func (c *Client) makeRpcFunc(field reflect.StructField, convention NamingConvention, namePrefix string) (reflect.Value, error) {
	ftyp := field.Type
	if ftyp.Kind() != reflect.Func {
		return reflect.Value{}, errors.New("handler field must be func")
	}

	hasCtx := false
	var argType reflect.Type
	argPos := 0

	if ftyp.NumIn() >= 1 && ftyp.In(0) == typeOfContext {
		argPos = 1
		hasCtx = true
	}

	if ftyp.NumIn() > argPos {
		argType = ftyp.In(argPos)
		if !isExportedOrBuiltinType(argType) {
			return reflect.Value{}, errors.New(fmt.Sprintf("%s is not exported", field.Name))
		}
	}

	// Method out at most 2.
	if ftyp.NumOut() > 2 {
		return reflect.Value{}, errors.New(fmt.Sprintf("%s out num are greater than 2", field.Name))
	}

	errPos := -1
	valOutPos := -1

	// Error must be returned, and it must be the last returned value.
	switch {
	case ftyp.NumOut() == 1:
		if ftyp.Out(0) != typeOfError {
			return reflect.Value{}, errors.New("error must be returned")
		}
		errPos = 0
	case ftyp.NumOut() == 2:
		if ftyp.Out(1) != typeOfError {
			return reflect.Value{}, errors.New("error must be the last return value")
		}
		valOutPos = 0
		errPos = 1
	}

	// rpcType
	var rt rpcType
	switch field.Tag.Get("rpc") {
	case "jsonrpc2":
		rt = typeJsonrpc2
	case "rest":
		rt = typeRest
	default:
		rt = typeJsonrpc2
	}

	// name override
	name, found := field.Tag.Lookup("name")
	if !found {
		switch convention {
		case CamelCase:
			name = CamelCaseName(field.Name)
		case SnakeCase:
			name = SnakeCaseName(field.Name)
		case LowerCase:
			name = LowerCaseName(field.Name)
		default:
			name = field.Name
		}
	}
	name = namePrefix + name

	// rpc parameter container type
	containerT := arrayParamType
	containerTypeStr, found := field.Tag.Lookup("container")
	if found {
		containerT = paramContainerType(containerTypeStr)
	}

	// http method
	httpMethod, found := field.Tag.Lookup("method")
	if !found {
		httpMethod = "GET"
	}
	httpMethod = strings.ToUpper(strings.TrimSpace(httpMethod))

	// cache control
	cacheTag, found := field.Tag.Lookup("cache")
	cacheCtl := cacheControl{
		cacheable: false,
		key:       "",
		ttl:       time.Second * 1,
	}

	// only GET method support cache
	if found {
		cacheCtl.cacheable = true
		settings := tag.ParseTagSettings(cacheTag, ",")
		cacheCtl.key = c.cacheKeyOf(settings, httpMethod, name)
		ttl, foundTtl := settings["ttl"]
		if foundTtl {
			d, err := time.ParseDuration(ttl)
			if err == nil {
				cacheCtl.ttl = d
			}
		}
		c.log.Trace("cache control of "+name, "settings", cacheCtl)
	}

	f := &rpcFunc{
		client:             c,
		ftyp:               ftyp,
		name:               name,
		paramContainerType: containerT,
		nout:               ftyp.NumOut(),
		errOut:             errPos,
		valOut:             valOutPos,
		retry:              false,
		httpMethod:         httpMethod,
		rpcType:            rt,
		hasCtx:             hasCtx,
		cacheCtl:           cacheCtl,
	}

	return reflect.MakeFunc(ftyp, f.handleRpcCall), nil
}

func (c *Client) cacheKeyOf(settings map[string]string, httpMethod, fname string) string {
	k, found := settings["key"]
	if found {
		switch k {
		case "-name":
			return fname
		}
	}
	// fallback
	return httpMethod + ":" + fname
}

func (c *Client) sendJsonrpc2(req *jsonrpc2.Request) (*jsonrpc2.Response, error) {
	data, err := jsonrpc2.EncodeMessage(req)
	if err != nil {
		return nil, err
	}
	rawResp, err := c.fetch.Post("").SetJSONBody(data).Execute()
	if err != nil {
		return nil, err
	}
	msg, err := jsonrpc2.DecodeMessage(rawResp.BodyBytes())
	if err != nil {
		return nil, err
	}
	return msg.(*jsonrpc2.Response), err
}

func (c *Client) sendRest(method, name string, param interface{}) (*fetch.Response, error) {
	return c.fetch.Do(method, name).SetJSONBody(param).Execute()
}

type rpcType int

const (
	typeJsonrpc2 rpcType = iota
	typeRest
)

type cacheControl struct {
	cacheable bool
	key       string
	ttl       time.Duration
}

type paramContainerType string

const (
	arrayParamType  = paramContainerType("array")
	objectParamType = paramContainerType("object")
)

type rpcFunc struct {
	client *Client

	ftyp reflect.Type
	name string

	paramContainerType paramContainerType

	nout   int
	valOut int
	errOut int

	hasCtx               bool
	returnValueIsChannel bool

	rpcType    rpcType
	httpMethod string // http rpc method, it is ignored when rpcType == typeJsonrpc2
	retry      bool

	// is cache enabled, default is false
	cacheCtl cacheControl
}

func (fn *rpcFunc) responseFromCache() ([]reflect.Value, bool) {
	if fn.cacheCtl.cacheable && fn.cacheCtl.key != "" {
		if v, ok := fn.client.cacheStorage.Get(fn.cacheCtl.key); ok {
			out := make([]reflect.Value, fn.nout)
			if fn.errOut != -1 {
				out[fn.errOut] = reflect.New(typeOfError).Elem()
			}
			if fn.valOut != -1 {
				out[fn.valOut] = v.(reflect.Value)
			}
			return out, true
		}
	}
	return nil, false
}

func (fn *rpcFunc) processResponse(err error, rval reflect.Value) []reflect.Value {
	out := make([]reflect.Value, fn.nout)

	if fn.valOut != -1 {
		out[fn.valOut] = rval
		if fn.cacheCtl.cacheable && err == nil {
			// cache the result when cache is enabled and err is nil
			fn.client.cacheStorage.Save(fn.cacheCtl.key, rval, fn.cacheCtl.ttl)
		}
	}
	if fn.errOut != -1 {
		out[fn.errOut] = reflect.New(typeOfError).Elem()
		if err != nil {
			out[fn.errOut].Set(reflect.ValueOf(err))
		}
	}

	return out
}

func (fn *rpcFunc) processError(err error) []reflect.Value {
	out := make([]reflect.Value, fn.nout)

	if fn.valOut != -1 {
		out[fn.valOut] = reflect.New(fn.ftyp.Out(fn.valOut)).Elem()
	}
	if fn.errOut != -1 {
		out[fn.errOut] = reflect.New(typeOfError).Elem()
		out[fn.errOut].Set(reflect.ValueOf(&ErrClient{err}))
	}

	return out
}

func (fn *rpcFunc) handleRpcCall(args []reflect.Value) (results []reflect.Value) {
	id := atomic.AddInt64(&fn.client.idCtr, 1)
	apos := 0
	var ctx context.Context
	if fn.hasCtx {
		ctx = args[0].Interface().(context.Context)
		apos = 1
	} else {
		ctx = context.Background()
	}

	retVal := func() reflect.Value { return reflect.Value{} }

	select {
	case <-ctx.Done():
		return fn.processError(ctx.Err())
	default:
	}

	// try to get the response from cache
	if r, ok := fn.responseFromCache(); ok {
		return r
	}

	// handle jsonrpc2
	if fn.rpcType == typeJsonrpc2 {
		var param interface{}
		if fn.paramContainerType == objectParamType {
			if apos < len(args) {
				param = args[apos].Interface()
			}
		} else {
			arrayParam := make([]interface{}, len(args)-apos)
			for i, arg := range args[apos:] {
				arrayParam[i] = arg.Interface()
			}
			param = arrayParam
		}

		req, err := jsonrpc2.NewCall(jsonrpc2.Int64ID(id), fn.name, param)
		if err != nil {
			return fn.processError(fmt.Errorf("create request failed: %w", err))
		}
		resp, err := fn.client.sendJsonrpc2(req)
		if err != nil {
			switch err.(type) {
			case web.ServerError:
				return fn.processResponse(err, reflect.New(fn.ftyp.Out(fn.valOut)).Elem())
			default:
				return fn.processError(fmt.Errorf("sendRequest failed: %w", err))
			}
		}

		if resp.ID != req.ID {
			return fn.processError(errors.New("request and response id didn't match"))
		}

		if fn.valOut != -1 {
			val := reflect.New(fn.ftyp.Out(fn.valOut))

			if resp.Result != nil {
				if err = json.Unmarshal(resp.Result, val.Interface()); err != nil {
					fn.client.log.Warn("unmarshalling failed", "message", string(resp.Result))
					return fn.processError(fmt.Errorf("unmarshalling result: %w", err))
				}
			}

			retVal = func() reflect.Value { return val.Elem() }
		}

		return fn.processResponse(resp.Error, retVal())
	} else {
		// handle rest
		var p interface{}
		if apos < len(args) {
			p = args[apos].Interface()
		}

		resp, err := fn.client.sendRest(fn.httpMethod, fn.name, p)
		if err != nil {
			switch err.(type) {
			case web.ServerError:
				return fn.processResponse(err, reflect.New(fn.ftyp.Out(fn.valOut)).Elem())
			default:
				return fn.processError(fmt.Errorf("sendRequest failed: %w", err))
			}
		}
		if fn.valOut != -1 {
			val := reflect.New(fn.ftyp.Out(fn.valOut))

			if resp.BodyBytes() != nil && len(resp.BodyBytes()) > 0 {
				if err = json.Unmarshal(resp.BodyBytes(), val.Interface()); err != nil {
					fn.client.log.Warn("unmarshalling failed", "message", string(resp.BodyBytes()))
					return fn.processError(fmt.Errorf("unmarshalling result: %w", err))
				}
			}

			retVal = func() reflect.Value { return val.Elem() }
		}
		return fn.processResponse(nil, retVal())
	}
}
