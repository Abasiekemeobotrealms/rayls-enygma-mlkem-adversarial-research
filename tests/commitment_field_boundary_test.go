package tests

import (
	"math/big"
	"testing"

	"github.com/raylsnetwork/enygma_dvp/src/core"
)

func TestErc20CommitmentV2_FieldBoundary(t *testing.T) {
	q, ok := new(big.Int).SetString(
		"21888242871839275222246405745257275088548364400416034343698204186575808495617",
		10,
	)
	if !ok {
		t.Fatal("failed to parse BN254 scalar field modulus")
	}

	qMinusOne := new(big.Int).Sub(new(big.Int).Set(q), big.NewInt(1))

	base := []*big.Int{
		big.NewInt(123),
		big.NewInt(456),
		big.NewInt(5),
		big.NewInt(10),
	}

	tests := []struct {
		name       string
		inputIndex int
	}{
		{"salt_minus_one_equals_q_minus_one", 1},
		{"amount_minus_one_equals_q_minus_one", 2},
		{"token_id_minus_one_equals_q_minus_one", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			negativeInputs := make([]*big.Int, len(base))
			canonicalInputs := make([]*big.Int, len(base))

			for i := range base {
				negativeInputs[i] = new(big.Int).Set(base[i])
				canonicalInputs[i] = new(big.Int).Set(base[i])
			}

			negativeInputs[tc.inputIndex] = big.NewInt(-1)
			canonicalInputs[tc.inputIndex] = new(big.Int).Set(qMinusOne)

			negativeHash, err := core.Erc20CommitmentV2(
				negativeInputs[0], negativeInputs[1],
				negativeInputs[2], negativeInputs[3],
			)
			if err != nil {
				t.Fatalf("negative input returned error: %v", err)
			}

			canonicalHash, err := core.Erc20CommitmentV2(
				canonicalInputs[0], canonicalInputs[1],
				canonicalInputs[2], canonicalInputs[3],
			)
			if err != nil {
				t.Fatalf("Q-1 input returned error: %v", err)
			}

			if negativeHash.Cmp(canonicalHash) != 0 {
				t.Fatalf("hash mismatch: -1=%s, Q-1=%s",
					negativeHash, canonicalHash)
			}
		})
	}

	for _, value := range []*big.Int{
		new(big.Int).Set(q),
		new(big.Int).Add(new(big.Int).Set(q), big.NewInt(1)),
	} {
		_, err := core.Erc20CommitmentV2(
			base[0], value, base[2], base[3],
		)
		if err == nil {
			t.Errorf("expected input %s to be rejected", value)
		}
	}
}
