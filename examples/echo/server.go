package main

import (
	"flag"
	"log"
	"net"

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
	buf := make([]byte, 1024)
	for {
		nr, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println(string(buf[:nr]))

		_, err = conn.Write(buf[:nr])
		if err != nil {
			log.Println(err)
			return
		}
	}
}
