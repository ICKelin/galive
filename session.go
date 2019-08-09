package galive

import (
	"net"
	"time"
)

type Session struct {
	conn    net.Conn
	isclose int32
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) Read(buf []byte) (int, error) {
	return -1, nil
}

func (s *Session) Write(buf []byte) (int, error) { return -1, nil }

func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Session) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *Session) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Session) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}
