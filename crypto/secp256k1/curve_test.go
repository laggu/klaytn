package secp256k1

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAddSamePoint(t *testing.T) {
	/*
		This test is intended to highlight the bug in go-gxplatform/crypto/secp256k1/curve.go#affineFromJacobian.
		When passed with same points, BitCurve.Add invokes affineFromJacobian(0, 0, 0) which then invokes
		(big.Int).Mul(nil, nil).

		Although executing (big.Int).Mul(nil, nil) is not problematic in Go 1.10.3, it has been found invoking the same
		is fatal in Go 1.11 (causing SIGSEGV; terminating the program).

	*/
	x0, _ := hex.DecodeString("4f52e337ad8bf1ce10cbb72ab91d9954474cea39811040df5558297df3e3c1bf") // Alice
	x1, _ := hex.DecodeString("4f52e337ad8bf1ce10cbb72ab91d9954474cea39811040df5558297df3e3c1bf") // Alice

	G := S256()
	P0x, P0y := G.ScalarBaseMult(x0)
	P1x, P1y := G.ScalarBaseMult(x1)

	Qx, Qy := G.Add(P0x, P0y, P1x, P1y)

	fmt.Println(Qx, Qy) // should print 0 0 with no fatal error
}
