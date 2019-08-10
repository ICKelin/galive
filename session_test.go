package galive

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	log.SetFlags(log.Lshortfile | log.Ltime)
	listener, err := net.Listen("tcp", "127.0.0.1:10009")
	if err != nil {
		t.Error(err)
		return
	}

	defer listener.Close()
	go client(t)

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		go func() {
			sess := NewServerSession(conn)
			defer sess.Close()

			buf := make([]byte, 1024)
			for {
				nr, err := sess.Read(buf)
				if err != nil {
					log.Println(err)
					return
				}

				log.Println(string(buf[:nr]))
			}
		}()
	}
}

func client(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:10009")
	if err != nil {
		t.Error(err)
		return
	}

	// NewClient(conn)
	sess := NewClientSession(conn)
	tc := time.NewTicker(time.Second * 1)
	defer tc.Stop()

	for range tc.C {
		sess.Write([]byte("ping"))
	}
}
