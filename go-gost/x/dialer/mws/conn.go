package mws

import (
	"net"
	"time"

	"github.com/go-gost/x/internal/util/mux"
)

type muxSession struct {
	conn      net.Conn
	session   *mux.Session
	createdAt time.Time
}

func (session *muxSession) GetConn() (net.Conn, error) {
	return session.session.GetConn()
}

func (session *muxSession) Accept() (net.Conn, error) {
	return session.session.Accept()
}

func (session *muxSession) Close() error {
	if session == nil {
		return nil
	}

	var err error
	if session.session != nil {
		err = session.session.Close()
	}
	if session.conn != nil {
		_ = session.conn.Close()
	}
	return err
}

func (session *muxSession) IsClosed() bool {
	if session == nil {
		return true
	}
	if session.session == nil {
		return session.conn == nil
	}
	return session.session.IsClosed()
}

func (session *muxSession) ShouldRotate(now time.Time, maxLifetime time.Duration) bool {
	if session == nil || maxLifetime <= 0 || session.createdAt.IsZero() {
		return false
	}
	if now.Sub(session.createdAt) < maxLifetime {
		return false
	}
	if session.session == nil {
		return true
	}
	return session.session.NumStreams() == 0
}

func (session *muxSession) NumStreams() int {
	if session == nil || session.session == nil {
		return 0
	}
	return session.session.NumStreams()
}
