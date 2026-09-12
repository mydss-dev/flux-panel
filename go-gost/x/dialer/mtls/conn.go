package mtls

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

// IsClosed treats a just-created physical connection as alive while the first
// TLS/SMUX handshake is still pending. Treating session == nil as "closed"
// causes concurrent logical dials to delete each other's pending connection and
// produces intermittent "unrecognized connection"/handshake failures.
func (session *muxSession) IsClosed() bool {
	if session == nil {
		return true
	}
	if session.session == nil {
		return session.conn == nil
	}
	return session.session.IsClosed()
}

// ShouldRotate bounds the lifetime of an otherwise apparently healthy SMUX
// session. smux.OpenStream is local-only and can succeed even when the peer no
// longer services new streams, so IsClosed alone is not a sufficient liveness
// test. We only rotate an initialized session when it is idle to avoid cutting
// active application streams.
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
