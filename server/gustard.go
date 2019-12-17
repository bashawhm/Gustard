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

type Connection struct {
	cli net.Conn
	aes cipher.Block
}

func broadcaster(msgs chan string, clis chan Connection) {
	var clients []Connection
	for {
		select {
		case cli := <-clis:
			clients = append(clients, cli)
		case msg := <-msgs:
			fmt.Println("msg: ", msg)
			for i := 0; i < len(clients); i++ {
				s := make([]byte, 16)
				clients[i].aes.Encrypt(s, []byte(msg))
				fmt.Fprintf(clients[i].cli, string(s))
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

func handler(client net.Conn, msgs chan string, clis chan Connection) {
	k := 256

	fmt.Println("Generating Keys...")
	pubKey, privKey := genKeys(k)
	var clientPubKey PubKey
	var key []byte
	var clientKey []byte
	var aesBlock cipher.Block
	// fmt.Println(pubKey)
	// fmt.Println(privKey)

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
			// fmt.Println(clientPubKey)

			aesBlock, key = genAESCipher()
			keyCommand = "AESSYMKEY " + string(key) + "\n"
			cipher := encode_and_encrypt(keyCommand, &clientPubKey)
			cipherString := ""
			for i := 0; i < len(cipher); i++ {
				cipherString += cipher[i].String() + " "
			}
			fmt.Fprintf(client, cipherString+"\n")
			goto AES
		default:
			fmt.Println("Error: command in bad state: ", command)
		}
	}

	fmt.Println("Exchanging AES Key...")
AES:
	for nin.Scan() {
		command := nin.Text()
		parts := strings.Split(command, " ")
		var cs []Cipher
		for i := 0; i < len(parts)-1; i++ {
			cs = append(cs, toCipher(parts[i]))
		}
		command = decrypt_and_decode(cs, &pubKey, &privKey)
		parts = strings.Split(command, " ")
		switch parts[0] {
		case "AESSYMKEY":
			fmt.Println("AESKEYLEN: ", len(parts[1]))
			if len(parts[1]) < 32 {
				fmt.Println("Key exchange failed")
				return
			}
			clientKey = []byte(parts[1])[:32]
			clis <- Connection{cli: client, aes: aesBlock}
			goto normal
		}
	}

normal:
	cliAES, _ := aes.NewCipher(clientKey)
	fmt.Println("Client Ready")
	for {
		input := make([]byte, 16)
		_, err := client.Read(input)
		if err != nil {
			fmt.Println("Client Disconnected")
			return
		}
		// fmt.Println("Input: ", input)
		s := make([]byte, 16)
		cliAES.Decrypt(s, input)
		msgs <- string(s)
	}
}

func main() {
	fmt.Println("Starting server...")
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error: port already in use")
		return
	}

	msgs := make(chan string, 10)
	clients := make(chan Connection, 10)

	go broadcaster(msgs, clients)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error: failed to connect to incoming client")
			continue
		}

		fmt.Print("Accepted connection: ")
		fmt.Println(conn.RemoteAddr().String())
		go handler(conn, msgs, clients)
	}
}
