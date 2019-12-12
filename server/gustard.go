package main

import (
	"bufio"
	"fmt"
	"math/big"
	"net"
	"strings"
)

func handler(client net.Conn) {
	k := 1024

	pubKey, privKey := genKeys(k)
	var clientPubKey PubKey
	fmt.Println(pubKey)
	fmt.Println(privKey)

	keyCommand := "PUBKEY " + pubKey.p.String() + " " + pubKey.g.String() + " " + pubKey.a.String() + "\n"
	fmt.Fprintf(client, keyCommand)

	nin := bufio.NewScanner(bufio.NewReader(client))
	nin.Split(bufio.ScanLines)
	for nin.Scan() {
		command := nin.Text()
		parts := strings.Split(command, " ")
		switch parts[0] {
		case "PUBKEY":
			if len(parts) < 4 {
				fmt.Println("Error parsing PUBKEY command")
				continue
			}
			var p big.Int
			var g big.Int
			var a big.Int
			p.SetString(parts[1], 10)
			g.SetString(parts[2], 10)
			a.SetString(parts[3], 10)
			clientPubKey.p = &p
			clientPubKey.g = &g
			clientPubKey.a = &a
			fmt.Println(clientPubKey)
		}
	}
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
