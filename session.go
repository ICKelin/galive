package galive

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	conn          net.Conn
	sendBuf       chan *writeReq
	recvBufMu     sync.Mutex
	recvBuf       [][]byte
	dataReady     chan struct{}
	writeDeadline atomic.Value

	chReadError  chan struct{}
	readError    atomic.Value
	readDeadline atomic.Value
	isclose      int32
}

type writeReq struct {
	buf            []byte
	nwrite         int
	chWriteSuccess chan struct{}
	chWriteError   chan struct{}
	writeError     atomic.Value
}

func NewServerSession(conn net.Conn) *Session {
	s := &Session{
		conn:        conn,
		chReadError: make(chan struct{}),
		dataReady:   make(chan struct{}),
		sendBuf:     make(chan *writeReq),
		recvBuf:     make([][]byte, 0),
	}
	go s.open()
	return s
}

func NewClientSession(conn net.Conn) *Session {
	s := &Session{
		conn:        conn,
		chReadError: make(chan struct{}),
		dataReady:   make(chan struct{}),
		sendBuf:     make(chan *writeReq),
		recvBuf:     make([][]byte, 0),
	}
	go s.open()
	return s
}

func (s *Session) open() {
	ctx := context.TODO()
	go s.sender(ctx)
	go s.receiver(ctx)
	s.heartbeat(ctx)
}

func (s *Session) sender(ctx context.Context) {
	defer s.conn.Close()
	defer func() { atomic.StoreInt32(&s.isclose, 1) }()

	for w := range s.sendBuf {
		if atomic.LoadInt32(&s.isclose) == 1 {
			w.writeError.Store(fmt.Errorf("session close"))
			close(w.chWriteError)
			return
		}

		nw, err := s.conn.Write(w.buf)
		if err != nil {
			w.writeError.Store(err)
			close(w.chWriteError)
			return
		}

		w.nwrite = nw
		close(w.chWriteSuccess)
	}

}

func (s *Session) receiver(ctx context.Context) {
	defer s.conn.Close()
	defer func() { atomic.StoreInt32(&s.isclose, 1) }()

	var h hdr
	var readerr error
	for {
		if atomic.LoadInt32(&s.isclose) == 1 {
			return
		}

		_, err := io.ReadFull(s.conn, h[:])
		if err != nil {
			readerr = err
			goto errexit
		}

		// verify version
		if h.version() != majorVersion {
			log.Println("ERR: version not match")
			readerr = fmt.Errorf("ERR: galive version not match")
			goto errexit
		}

		switch h.cmd() {
		case cmd_heartbeat:

		case cmd_data:
			bodylen := h.bodylen()
			buf := make([]byte, bodylen)
			nr, err := io.ReadFull(s.conn, buf)
			if err != nil {
				log.Println(err)
				readerr = err
				goto errexit
			}

			s.recvBufMu.Lock()
			s.recvBuf = append(s.recvBuf, buf[:nr])
			s.recvBufMu.Unlock()
			// notify Read(), data ready
			select {
			case s.dataReady <- struct{}{}:
			default:
			}
		}
	}
errexit:
	s.readError.Store(readerr)
	close(s.chReadError)
}

func (s *Session) heartbeat(ctx context.Context) {
	defer func() { atomic.StoreInt32(&s.isclose, 1) }()
	tc := time.NewTicker(time.Second * 2)
	defer tc.Stop()

	var hb = [4]byte{byte(cmd_heartbeat), byte(majorVersion)}
	for range tc.C {
		if atomic.LoadInt32(&s.isclose) == 1 {
			break
		}

		_, err := s.write(hb[:])
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (s *Session) Read(buf []byte) (int, error) {
	n := len(buf)
	if n <= 0 {
		return 0, nil
	}

	nr := 0
	for {
		s.recvBufMu.Lock()
		if len(s.recvBuf) > 0 {
			nr = copy(buf, s.recvBuf[0])
			s.recvBuf[0] = s.recvBuf[0][nr:]
			if len(s.recvBuf[0]) == 0 {
				s.recvBuf = s.recvBuf[1:]
			}
		}
		s.recvBufMu.Unlock()

		if nr > 0 {
			return nr, nil
		}

		var deadline <-chan time.Time
		d, ok := s.readDeadline.Load().(time.Time)
		if ok && !d.IsZero() {
			tm := time.NewTimer(time.Until(d))
			deadline = tm.C
		}

		select {
		case <-s.dataReady:
			continue
		case <-deadline:
			return -1, fmt.Errorf("read timeout")
		case <-s.chReadError:
			err, ok := s.readError.Load().(error)
			if ok {
				return -1, err
			}
			return 0, nil
		}

	}
}

func (s *Session) Write(buf []byte) (int, error) {
	writes := make([]byte, 0)
	writes = append(writes, byte(cmd_data))
	writes = append(writes, byte(majorVersion))

	blen := make([]byte, 2)
	binary.BigEndian.PutUint16(blen, uint16(len(buf)))
	writes = append(writes, blen...)
	writes = append(writes, buf...)

	return s.write(writes)
}

func (s *Session) write(writes []byte) (int, error) {
	req := &writeReq{
		buf:            writes,
		chWriteSuccess: make(chan struct{}),
		chWriteError:   make(chan struct{}),
	}

	s.sendBuf <- req

	select {
	case <-req.chWriteSuccess:
		close(req.chWriteError)
		return req.nwrite, nil

	case <-req.chWriteError:
		close(req.chWriteSuccess)
		return -1, req.writeError.Load().(error)
	}
}

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
	s.readDeadline.Store(t)
	s.writeDeadline.Store(t)
	return nil
}

func (s *Session) SetReadDeadline(t time.Time) error {
	s.readDeadline.Store(t)
	return nil
}

func (s *Session) SetWriteDeadline(t time.Time) error {
	s.writeDeadline.Store(t)
	return nil
}
