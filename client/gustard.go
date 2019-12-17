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

func aesSplitMsg(s string) []string {
	var res []string
	tmp := ""
	res = append(res, "BEGINMSG********")
	for i := 0; i < len(s); i += 16 {
		if i+16 > len(s) {
			tmp = ""
			tmp = s[i:]
			for j := (i / 16) + len(tmp); j <= 16; j++ {
				tmp += "*"
			}
			res = append(res, tmp)
			break
		}
		res = append(res, s[i:i+16])
	}
	res = append(res, "ENDMSG**********")
	return res
}

func aesMergeMsg(s []string) string {
	res := ""
	for i := 1; i < len(s)-1; i++ {
		res += s[i]
	}
	return res
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

func chatter(conn net.Conn, aesBlock cipher.Block) {
	cin := bufio.NewScanner(bufio.NewReader(os.Stdin))
	cin.Split(bufio.ScanLines)

	for cin.Scan() {
		ss := aesSplitMsg(cin.Text())
		for i := 0; i < len(ss); i++ {
			s := make([]byte, 16)
			aesBlock.Encrypt(s, []byte(ss[i][:16]))
			fmt.Fprintf(conn, string(s))
		}
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
	var serverKey []byte
	var aesBlock cipher.Block

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

			aesBlock, key = genAESCipher()
			keyCommand = "AESSYMKEY " + string(key) + "\n"
			cipher := encode_and_encrypt(keyCommand, &serverPubKey)
			cipherString := ""
			for i := 0; i < len(cipher); i++ {
				cipherString += cipher[i].String() + " "
			}
			fmt.Fprintf(conn, cipherString+"\n")
			goto AES
		default:
			fmt.Println(command)
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
			serverKey = []byte(parts[1])[:32]
			goto normal
		}
	}

normal:
	servBlock, _ := aes.NewCipher(serverKey)
	go chatter(conn, aesBlock)
	fmt.Println("Server Ready")
	var sBlock []string
	input := make([]byte, 16)
	for {
		s := make([]byte, 16)
		_, err := conn.Read(input)
		if err != nil {
			panic(err)
		}

		servBlock.Decrypt(s, input)
		sBlock = append(sBlock, string(s))
		if string(s) == "ENDMSG**********" {
			res := aesMergeMsg(sBlock)
			sBlock = []string{}
			res = strings.Trim(res, "*")
			fmt.Println(res)
		}
	}
}
