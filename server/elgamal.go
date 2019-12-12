package main

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"time"
)

type PubKey struct {
	p *big.Int
	g *big.Int
	a *big.Int
}

type PrivKey struct {
	b *big.Int
}

func getNumber(x int64) big.Int {
	var num big.Int
	num.SetInt64(x)
	return num
}

func genGenerator(rand io.Reader, p *big.Int, q *big.Int) *big.Int {
	var mod big.Int
	eight := getNumber(8)
	five := getNumber(5)
	three := getNumber(3)
	two := getNumber(2)
	mod.Mod(p, &eight)
	if mod.Cmp(&three) == 0 || mod.Cmp(&five) == 0 {
		return &two
	}
	neg_one := getNumber(-1)
	for G, _ := crand.Int(rand, p); ; G, _ = crand.Int(rand, p) {
		var mod_result big.Int
		mod_result.Exp(G, q, p)
		if mod_result.Cmp(&neg_one) != 0 {
			return G
		}
	}
}

func genKeys(k int) (PubKey, PrivKey) {
	var pubKey PubKey
	var privKey PrivKey
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //Not crypto secure, but will work for now

	is_safe := false
	var safe big.Int
	var unit big.Int       //We need a literal 1
	var even_prime big.Int //We need a literal 2
	var q *big.Int
	unit.SetInt64(1)
	even_prime.SetInt64(2)
	for !is_safe {
		q, _ = crand.Prime(rng, k) //Bruh moment, we need a safe prime
		safe.Mul(q, &even_prime)
		safe.Add(&safe, &unit) //safe now contains 2p+1
		is_safe = safe.ProbablyPrime(10)
	}
	pubKey.p = &safe
	pubKey.g = genGenerator(rng, pubKey.p, q)
	privKey.b, _ = crand.Int(rng, pubKey.p)
	fmt.Println("g=",pubKey.g.String())
	fmt.Println("b=",privKey.b.String())
	pubKey.a = new(big.Int)
	pubKey.a.Exp(pubKey.g, privKey.b, pubKey.p)
	fmt.Println("a=",pubKey.a.String())
	return pubKey, privKey
}
