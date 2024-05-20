package rpc

import (
	"bufio"
	"context"
	"crypto-trade-client/common/rpc/jsonrpc2"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func httpFramerCtor() jsonrpc2.Framer {
	return httpFramer{}
}

type httpFramer struct{}
type httpHeaderReader struct{ in *bufio.Reader }
type headerWriter struct{ out io.Writer }

func (h httpFramer) Reader(in io.Reader) jsonrpc2.Reader {
	return &httpHeaderReader{in: bufio.NewReader(in)}
}

func (h httpFramer) Writer(rw io.Writer) jsonrpc2.Writer {
	return &headerWriter{out: rw}
}

func (h httpHeaderReader) Read(ctx context.Context) (jsonrpc2.Message, int64, error) {
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
	}

	req, err := http.ReadRequest(h.in)
	if err != nil {
		return nil, 0, fmt.Errorf("failed reading header line: %w", err)
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed reading body: %w", err)
	}
	msg, err := jsonrpc2.DecodeMessage(data)
	return msg, 0, err
}

func (w *headerWriter) Write(ctx context.Context, msg jsonrpc2.Message) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	data, err := jsonrpc2.EncodeMessage(msg)
	if err != nil {
		return 0, fmt.Errorf("marshaling message: %v", err)
	}
	buf := []byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %v\r\n\r\n", len(data)))
	buf = append(buf, data...)
	w.out.Write(buf)
	return int64(len(buf)), err
}

func Test_client_jsonrpc2(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	listener, err := jsonrpc2.NetPipeListener(ctx)
	if err != nil {
		t.Fatal(err)
	}
	server, err := jsonrpc2.Serve(ctx, listener, binder{framer: httpFramerCtor()})
	if err != nil {
		t.Fatal(err)
	}
	initialTransport := http.DefaultTransport
	var conn io.ReadWriteCloser
	defer func() {
		http.DefaultTransport = initialTransport
		if conn != nil {
			conn.Close()
		}
		listener.Close()
		server.Wait()
	}()
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, ierr := listener.Dialer().Dial(ctx)
		if ierr != nil {
			return nil, ierr
		}
		conn = c
		return c.(net.Conn), nil
	}
	http.DefaultTransport = &http.Transport{
		DialContext: dial,
	}
	var s svc
	err = NewClient(ctx, "http://localhost", "test", &s, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	v, err := s.Inc(1)
	require.Nil(t, err)
	assert.Equal(t, 1, v)

	resp, err := s.Hello()
	require.Nil(t, err)
	assert.Equal(t, "hello", resp)
}

type SimpleServerHandler struct {
	n            int
	cacheCounter *atomic.Int32
}

func (h SimpleServerHandler) Echo(writer http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.Write([]byte(fmt.Sprintf("server error: %v", err)))
		return
	}
	result, _ := json.Marshal(fmt.Sprintf("%s", string(data)))
	writer.Write(result)
}

func (h SimpleServerHandler) CacheTest(writer http.ResponseWriter, req *http.Request) {
	h.cacheCounter.Add(1)
	bz, _ := json.Marshal(h.cacheCounter.Load())
	writer.Write(bz)
}

func (h SimpleServerHandler) DefaultNoCache(writer http.ResponseWriter, req *http.Request) {
	h.cacheCounter.Add(1)
	bz, _ := json.Marshal(h.cacheCounter.Load())
	writer.Write(bz)
}

func (h *SimpleServerHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/echo":
		h.Echo(writer, req)
	case "/cache-test":
		h.CacheTest(writer, req)
	case "/default-no-cache":
		h.DefaultNoCache(writer, req)
	}
}

func Test_client_RestRpc(t *testing.T) {
	hclog.L().SetLevel(hclog.Trace)
	rpcHandler := SimpleServerHandler{}
	server := httptest.NewServer(&rpcHandler)
	defer server.Close()
	var s svc
	err := NewClient(context.Background(), "http://"+server.Listener.Addr().String(), "test", &s, map[string]string{})
	require.Nil(t, err)
	r, err := s.Echo("ping-ping")
	require.Nil(t, err)
	assert.Equal(t, "ping-ping", r)
}

func Test_client_cache(t *testing.T) {
	hclog.L().SetLevel(hclog.Trace)
	rpcHandler := SimpleServerHandler{
		cacheCounter: atomic.NewInt32(0),
	}
	server := httptest.NewServer(&rpcHandler)
	defer server.Close()
	var s svc
	err := NewClient(context.Background(), "http://"+server.Listener.Addr().String(), "test", &s, map[string]string{})
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		r, err := s.CacheTest(strconv.Itoa(i))
		require.NoError(t, err)
		assert.Equal(t, 1, r)
	}
	// cache expired
	time.Sleep(2 * time.Second)
	r, err := s.CacheTest("10")
	require.NoError(t, err)
	assert.EqualValues(t, 2, r)
	// default no cache
	for i := 0; i < 10; i++ {
		r, err := s.DefaultNoCache(strconv.Itoa(i + 11))
		require.NoError(t, err)
		assert.Equal(t, 3+i, r)
	}
}

func Test_client_jsonrpc2_cache(t *testing.T) {
	hclog.L().SetLevel(hclog.Trace)
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	listener, err := jsonrpc2.NetPipeListener(ctx)
	if err != nil {
		t.Fatal(err)
	}
	server, err := jsonrpc2.Serve(ctx, listener, binder{framer: httpFramerCtor()})
	if err != nil {
		t.Fatal(err)
	}
	initialTransport := http.DefaultTransport
	var conn io.ReadWriteCloser
	defer func() {
		http.DefaultTransport = initialTransport
		if conn != nil {
			conn.Close()
		}
		listener.Close()
		server.Wait()
	}()
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, ierr := listener.Dialer().Dial(ctx)
		if ierr != nil {
			return nil, ierr
		}
		conn = c
		return c.(net.Conn), nil
	}
	http.DefaultTransport = &http.Transport{
		DialContext: dial,
	}
	var s svc
	err = NewClient(ctx, "http://localhost", "test", &s, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		r, err := s.DefaultNoCacheJsonrcp2()
		require.NoError(t, err)
		assert.EqualValues(t, i+1, r)
	}
}

type svc struct {
	Inc                    func(int) (int, error)       `rpc:"jsonrpc2"`
	Echo                   func(string) (string, error) `rpc:"rest" method:"POST" name:"echo"`
	Hello                  func() (string, error)       `rpc:"jsonrpc2"`
	DefaultNoCacheJsonrcp2 func() (int, error)          `rpc:"jsonrpc2"`
	CacheTest              func(string) (int, error)    `rpc:"rest" method:"GET" cache:"key:cache-test,ttl:1s" name:"cache-test"`
	DefaultNoCache         func(string) (int, error)    `rpc:"rest" method:"GET" name:"default-no-cache"`
}

type binder struct {
	framer jsonrpc2.Framer
}

func (b binder) Bind(ctx context.Context, conn *jsonrpc2.Connection) (jsonrpc2.ConnectionOptions, error) {
	h := &handler{
		conn:         conn,
		waitersBox:   make(chan map[string]chan struct{}, 1),
		calls:        make(map[string]*jsonrpc2.AsyncCall),
		cacheCounter: atomic.NewInt32(0),
	}
	h.waitersBox <- make(map[string]chan struct{})
	return jsonrpc2.ConnectionOptions{
		Framer:  b.framer,
		Handler: h,
	}, nil
}

type handler struct {
	conn         *jsonrpc2.Connection
	accumulator  int
	waitersBox   chan map[string]chan struct{}
	calls        map[string]*jsonrpc2.AsyncCall
	cacheCounter *atomic.Int32
}

func (h *handler) Handle(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	switch req.Method {
	case "Inc":
		var params []json.RawMessage
		json.Unmarshal(req.Params, &params)
		var v int
		if err := json.Unmarshal(params[0], &v); err != nil {
			return nil, fmt.Errorf("%w: %s", jsonrpc2.ErrParse, err)
		}
		h.accumulator += v
		return h.accumulator, nil
	case "Get":
		if len(req.Params) > 0 {
			return nil, fmt.Errorf("%w: expected no params", jsonrpc2.ErrInvalidParams)
		}
		return h.accumulator, nil
	case "Echo":
		var v []interface{}
		if err := json.Unmarshal(req.Params, &v); err != nil {
			return nil, fmt.Errorf("%w: %s", jsonrpc2.ErrParse, err)
		}
		var result interface{}
		err := h.conn.Call(ctx, v[0].(string), v[1]).Await(ctx, &result)
		return result, err
	case "Hello":
		return "hello", nil
	case "DefaultNoCacheJsonrcp2":
		h.cacheCounter.Add(1)
		return h.cacheCounter.Load(), nil
	default:
		return nil, jsonrpc2.ErrNotHandled
	}
}
