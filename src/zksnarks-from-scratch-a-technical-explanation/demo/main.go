package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
)

func main() {
	// circuit function
	// y = x^5 + 2*x + 6
	code := `
		func exp5(private a):
			b = a * a
			c = a * b
			d = a * c
			e = a * d
			return e

		func main(private s0, public s1):
			s2 = exp5(s0)
			s3 = s0 * 2
			s4 = s3 + s2
			s5 = s4 + 6
			equals(s1, s5)
			out = 1 * 1
	`
	fmt.Print("\ncode of the circuit:")
	fmt.Println(code)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	fmt.Println("\ncircuit data:", circuit)
	circuitJson, _ := json.Marshal(circuit)
	fmt.Println("circuit:", string(circuitJson))

	b8 := big.NewInt(int64(8))
	privateInputs := []*big.Int{b8}
	b32790 := big.NewInt(int64(32790))
	publicSignals := []*big.Int{b32790}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	if err != nil {
		panic(err)
	}

	// code to R1CS
	fmt.Println("\ngenerating R1CS from code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, TODO remove
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	_, _, _, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	// calculate trusted setup
	setup, err := snark.GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nt:", setup.Toxic.T)

	// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation

	proof, err := snark.GenerateProofs(*circuit, setup.Pk, w, px)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n proofs:")
	fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nsignals:", circuit.Signals)
	fmt.Println("witness:", w)
	b32790Verif := big.NewInt(int64(32790))
	publicSignalsVerif := []*big.Int{b32790Verif}
	before := time.Now()
	if !snark.VerifyProof(setup.Vk, proof, publicSignalsVerif, true) {
		fmt.Println("Verification not passed")
	}
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	if snark.VerifyProof(setup.Vk, proof, wrongPublicSignalsVerif, true) {
		fmt.Println("Verification should not have passed")
	}
}
