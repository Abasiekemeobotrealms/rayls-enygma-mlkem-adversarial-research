package core

import (
	"math/big"
	"testing"
)

func TestScanForErc20Notes_AcceptsTokenIDAtScalarFieldUpperBoundary(t *testing.T) {
	spend, err := NewSpendKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	view, err := NewViewKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Use the largest canonical BN254 scalar-field representative.
	tokenID := new(big.Int).Sub(
		new(big.Int).Set(SNARK_SCALAR_FIELD),
		big.NewInt(1),
	)
	amount := big.NewInt(100)

	ss, capsule, err := Encapsulate(view.EncapsKey)
	if err != nil {
		t.Fatal(err)
	}

	saltB, err := DerivePaymentSalt(ss)
	if err != nil {
		t.Fatal(err)
	}

	encKey, err := DerivePaymentKey(ss)
	if err != nil {
		t.Fatal(err)
	}

	saltBField := SaltBToField(saltB)

	encTxData, err := EncryptPayload(encKey, tokenID, amount)
	if err != nil {
		t.Fatal(err)
	}

	commitment, err := Erc20CommitmentV2(
		spend.PublicKey,
		saltBField,
		amount,
		tokenID,
	)
	if err != nil {
		t.Fatal(err)
	}

	event := OnChainErc20Event{
		Commitment: commitment,
		CipherText: capsule,
		EncTxData:  encTxData,
	}

	owned, rejected, err := ScanForErc20NotesWithRejected(
		view.DecapsKey,
		spend.PublicKey,
		[]OnChainErc20Event{event},
	)
	if err != nil {
		t.Fatalf("scanner returned an error: %v", err)
	}

	if len(rejected) != 0 {
		t.Fatalf("scanner rejected the event: %+v", rejected)
	}

	if len(owned) != 1 {
		t.Fatalf("scanner recovered %d notes, want 1", len(owned))
	}

	if owned[0].TokenId.Cmp(tokenID) != 0 {
		t.Fatalf("recovered token ID = %s, want %s",
			owned[0].TokenId, tokenID)
	}

	if owned[0].Amount.Cmp(amount) != 0 {
		t.Fatalf("recovered amount = %s, want %s",
			owned[0].Amount, amount)
	}

	t.Logf("accepted tokenID=R-1: %s", owned[0].TokenId)
	t.Logf("recovered amount: %s", owned[0].Amount)
}
