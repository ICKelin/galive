package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/ICKelin/galive"
)

func main() {
	flgLocal := flag.String("l", "", "127.0.0.1:10009")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	lis, err := net.Listen("tcp", *flgLocal)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			break
		}

		sess := galive.NewServerSession(conn)
		go onConn(sess)
	}
}

func onConn(conn net.Conn) {
	defer conn.Close()
	tc := time.NewTicker(time.Second * 3)
	defer tc.Stop()

	for range tc.C {
		_, err := conn.Write([]byte("push message from server"))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
