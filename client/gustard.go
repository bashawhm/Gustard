package main

import (
	"bufio"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <IP>")
		return
	}
	fmt.Println("Estabolishing connection...")
	var pubKeys PubKey
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		fmt.Println("Failed to connect: " + err.Error())
		return
	}

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
			pubKeys.p = &p
			pubKeys.g = &g
			pubKeys.a = &a
			fmt.Println(pubKeys)
		}
	}
}
