package main

import (
	crand "crypto/rand"
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

func genGenerator(p *big.Int) *big.Int {
	var G big.Int
	//TODO: All of this
	return &G
}

func genKeys(k int) (PubKey, PrivKey) {
	var pubKey PubKey
	var privKey PrivKey
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //Not crypto secure, but will work for now
	pubKey.p, _ = crand.Prime(rng, k)
	// pubKey.g = genGenerator(pubKey.p)
	privKey.b, _ = crand.Int(rng, pubKey.p)
	// pubKey.a.Exp(pubKey.g, privKey.b, pubKey.p)

	return pubKey, privKey
}
