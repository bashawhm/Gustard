package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

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

func chatter(conn net.Conn) {
	cin := bufio.NewScanner(bufio.NewReader(os.Stdin))
	cin.Split(bufio.ScanLines)

	for cin.Scan() {
		//TODO: encryption
		fmt.Fprintf(conn, cin.Text()+"\n")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <IP>")
		return
	}
	fmt.Println("Estabolishing connection...") //Establishing or Abolishing? Which one Hunter?
	var serverPubKey PubKey
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		fmt.Println("Failed to connect: " + err.Error())
		return
	}
	fmt.Println("Connection Estabolished")
	fmt.Println("Generating Key...")
	k := 256
	pubKey, privKey := genKeys(k)
	var key []byte
	var aes cipher.Block
	fmt.Println("CLI Pub Key ", pubKey)
	fmt.Println("CLI Priv Key ", privKey)

	go chatter(conn)

	nin := bufio.NewScanner(bufio.NewReader(conn))
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
			serverPubKey.p = &p
			serverPubKey.g = &g
			serverPubKey.a = &a
			fmt.Println("SERV Pub Key ", serverPubKey)
			keyCommand := "PUBKEY " + pubKey.p.String() + " " + pubKey.g.String() + " " + pubKey.a.String() + "\n"
			fmt.Fprintf(conn, keyCommand)

			aes, key = genAESCipher()
			fmt.Println("aes: ", aes)
			keyCommand = "AESSYMKEY " + string(key) + "\n"
			cipher := encode_and_encrypt(keyCommand, &serverPubKey)
			cipherString := ""
			for i := 0; i < len(cipher); i++ {
				cipherString += cipher[i].String() + " "
			}
			fmt.Fprintf(conn, cipherString+"\n")
			// fmt.Println("Encoded: ", cipherString)
			goto AES
		default:
			fmt.Println(command)
		}
	}

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
			fmt.Println("Parts: ", parts)
		}
	}
}
