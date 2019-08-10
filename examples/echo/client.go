package main

import (
	"log"
	"net"
	"time"

	"github.com/ICKelin/galive"
)

func main() {
	log.SetFlags(log.Lshortfile)
	conn, err := net.Dial("tcp", "127.0.0.1:10009")
	if err != nil {
		log.Println(err)
		return
	}

	sess := galive.NewClientSession(conn)
	tc := time.NewTicker(time.Second * 1)
	defer tc.Stop()

	buf := make([]byte, 1024)
	for range tc.C {
		_, err = sess.Write([]byte("ping"))
		if err != nil {
			log.Println(err)
			break
		}

		nr, err := sess.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println(string(buf[:nr]))
	}
}
