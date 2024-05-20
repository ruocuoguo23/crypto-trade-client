package jsonrpc2

import (
	"context"
	"errors"
	"go.uber.org/goleak"
	"testing"
	"time"
)

func TestIdleTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	listener, err := NetListener(ctx, "tcp", "localhost:0", NetListenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	listener = NewIdleListener(100*time.Millisecond, listener)
	defer listener.Close()
	server, err := Serve(ctx, listener, ConnectionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	connect := func() *Connection {
		client, err := Dial(ctx,
			listener.Dialer(),
			ConnectionOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return client
	}
	// Exercise some connection/disconnection patterns, and then assert that when
	// our timer fires, the server exits.
	conn1 := connect()
	conn2 := connect()
	if err := conn1.Close(); err != nil {
		t.Fatalf("conn1.Close failed with error: %v", err)
	}
	if err := conn2.Close(); err != nil {
		t.Fatalf("conn2.Close failed with error: %v", err)
	}
	conn3 := connect()
	if err := conn3.Close(); err != nil {
		t.Fatalf("conn3.Close failed with error: %v", err)
	}

	serverError := server.Wait()

	if !errors.Is(serverError, ErrIdleTimeout) {
		t.Errorf("run() returned error %v, want %v", serverError, ErrIdleTimeout)
	}
}

type msg struct {
	Msg string
}

type fakeHandler struct{}

func (fakeHandler) Handle(ctx context.Context, req *Request) (interface{}, error) {
	switch req.Method {
	case "ping":
		return &msg{"pong"}, nil
	default:
		return nil, ErrNotHandled
	}
}

func TestServe(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		factory func(context.Context) (Listener, error)
	}{
		{"tcp", func(ctx context.Context) (Listener, error) {
			return NetListener(ctx, "tcp", "localhost:0", NetListenOptions{})
		}},
		{"pipe", func(ctx context.Context) (Listener, error) {
			return NetPipeListener(ctx)
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake, err := test.factory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			conn, shutdown, err := newFake(t, ctx, fake)
			if err != nil {
				t.Fatal(err)
			}
			defer shutdown()
			var got msg
			if err := conn.Call(ctx, "ping", &msg{"ting"}).Await(ctx, &got); err != nil {
				t.Fatal(err)
			}
			if want := "pong"; got.Msg != want {
				t.Errorf("conn.Call(...): returned %q, want %q", got, want)
			}
		})
	}
}

func newFake(t *testing.T, ctx context.Context, l Listener) (*Connection, func(), error) {
	l = NewIdleListener(100*time.Millisecond, l)
	server, err := Serve(ctx, l, ConnectionOptions{
		Handler: fakeHandler{},
	})
	if err != nil {
		return nil, nil, err
	}

	client, err := Dial(ctx,
		l.Dialer(),
		ConnectionOptions{
			Handler: fakeHandler{},
		})
	if err != nil {
		return nil, nil, err
	}
	return client, func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}
		server.Wait()
	}, nil
}
