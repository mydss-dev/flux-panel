package mws

import (
	"net"
	"testing"
	"time"
)

func TestPendingMuxSessionLifecycle(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	created := time.Now()
	session := &muxSession{conn: client, createdAt: created}

	if session.IsClosed() {
		t.Fatal("pending physical connection must not be treated as closed before SMUX handshake")
	}
	if session.ShouldRotate(created.Add(time.Minute), 2*time.Minute) {
		t.Fatal("fresh pending session should not rotate")
	}
	if !session.ShouldRotate(created.Add(3*time.Minute), 2*time.Minute) {
		t.Fatal("stuck pending session should eventually rotate")
	}
}
