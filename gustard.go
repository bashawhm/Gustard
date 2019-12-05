package main

import (
	"fmt"
	"net"
)

func handler(client net.Conn) {
	//Hanlder for incomming connections
}

func main() {
	fmt.Println("Starting server...")
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error: port already in use")
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error: failed to connect to incoming client")
			continue
		}

		go handler(conn)
	}

}