package mux

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type deadlineSpyConn struct {
	net.Conn
	deadlineCalls      atomic.Int32
	readDeadlineCalls  atomic.Int32
	writeDeadlineCalls atomic.Int32
}

func (c *deadlineSpyConn) SetDeadline(t time.Time) error {
	c.deadlineCalls.Add(1)
	return c.Conn.SetDeadline(t)
}

func (c *deadlineSpyConn) SetReadDeadline(t time.Time) error {
	c.readDeadlineCalls.Add(1)
	return c.Conn.SetReadDeadline(t)
}

func (c *deadlineSpyConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadlineCalls.Add(1)
	return c.Conn.SetWriteDeadline(t)
}

func TestStreamConnDeadlinesDoNotTouchSharedConn(t *testing.T) {
	clientRaw, serverRaw := net.Pipe()
	clientSpy := &deadlineSpyConn{Conn: clientRaw}
	serverSpy := &deadlineSpyConn{Conn: serverRaw}

	cfg := &Config{KeepAliveDisabled: true}
	clientSession, err := ClientSession(clientSpy, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	serverSession, err := ServerSession(serverSpy, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()

	clientConn, err := clientSession.GetConn()
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	serverConn, err := serverSession.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	deadline := time.Now().Add(time.Second)
	if err := clientConn.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.SetReadDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.SetWriteDeadline(deadline); err != nil {
		t.Fatal(err)
	}

	if got := clientSpy.deadlineCalls.Load(); got != 0 {
		t.Fatalf("SetDeadline reached shared connection %d times", got)
	}
	if got := clientSpy.readDeadlineCalls.Load(); got != 0 {
		t.Fatalf("SetReadDeadline reached shared connection %d times", got)
	}
	if got := clientSpy.writeDeadlineCalls.Load(); got != 0 {
		t.Fatalf("SetWriteDeadline reached shared connection %d times", got)
	}

	// Verify the peer side is independently scoped too.
	if err := serverConn.SetDeadline(deadline); err != nil {
		t.Fatal(err)
	}
	if got := serverSpy.deadlineCalls.Load(); got != 0 {
		t.Fatalf("peer SetDeadline reached shared connection %d times", got)
	}
}
