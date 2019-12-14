package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"strings"
)

func broadcaster(msgs chan string, clis chan net.Conn) {
	var clients []net.Conn
	for {
		select {
		case cli := <-clis:
			clients = append(clients, cli)
		case msg := <-msgs:
			for i := 0; i < len(clients); i++ {
				fmt.Fprintf(clients[i], msg)
			}
		}
	}
}

func genAESCipher() (cipher.Block, []byte) {
	key := make([]byte, 32)
	rand.Read(key)
	cipher, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Failed to generate new aes key")
		panic(err)
	}
	return cipher, key
}

func handler(client net.Conn, msgs chan string) {
	k := 1024

	fmt.Println("Generating Keys...")
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
		default:
			fmt.Println(command)
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

	msgs := make(chan string, 1)
	clients := make(chan net.Conn, 1)

	go broadcaster(msgs, clients)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error: failed to connect to incoming client")
			continue
		}

		fmt.Print("Accepted connection: ")
		fmt.Println(conn.RemoteAddr().String())
		clients <- conn
		go handler(conn, msgs)
	}

}
