package main

import (
	"log"
	"net"

	"github.com/ICKelin/galive"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	conn, err := net.Dial("tcp", "127.0.0.1:10010")
	if err != nil {
		log.Println(err)
		return
	}

	sess := galive.NewClientSession(conn)
	buf := make([]byte, 1024)
	for {
		nr, err := sess.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println(string(buf[:nr]))
	}
}
