package tests

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/raylsnetwork/enygma_dvp/src/core"
)

func TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote(t *testing.T) {
	w := newWallet(t)

	corrupted := goodErc20Event(t, w, big.NewInt(10), big.NewInt(6))
	if len(corrupted.CipherText) != 1088 {
		t.Fatalf("ciphertext length = %d, want 1088", len(corrupted.CipherText))
	}

	// Change one ciphertext bit without changing its length.
	corrupted.CipherText[0] ^= 0x01

	if len(corrupted.CipherText) != 1088 {
		t.Fatalf("mutated ciphertext length = %d, want 1088", len(corrupted.CipherText))
	}

	good := goodErc20Event(t, w, big.NewInt(10), big.NewInt(5))

	owned, rejected, err := core.ScanForErc20NotesWithRejected(
		w.view.DecapsKey,
		w.spend.PublicKey,
		[]core.OnChainErc20Event{corrupted, good},
	)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(owned) != 1 || owned[0].Amount.Cmp(big.NewInt(5)) != 0 {
		t.Fatalf("owned = %+v, want exactly the later valid note of 5", owned)
	}

	if len(rejected) != 0 {
		t.Fatalf("rejected = %+v, want no reported rejection for a payload that does not open", rejected)
	}
}

func TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated(t *testing.T) {
	invalidLengths := []int{0, 1, 1087, 1089}

	for _, length := range invalidLengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			w := newWallet(t)

			malformed := goodErc20Event(t, w, big.NewInt(10), big.NewInt(6))
			malformed.CipherText = make([]byte, length)

			good := goodErc20Event(t, w, big.NewInt(10), big.NewInt(5))

			owned, rejected, err := core.ScanForErc20NotesWithRejected(
				w.view.DecapsKey,
				w.spend.PublicKey,
				[]core.OnChainErc20Event{malformed, good},
			)
			if err != nil {
				t.Fatalf("scan failed: %v", err)
			}

			if len(owned) != 1 || owned[0].Amount.Cmp(big.NewInt(5)) != 0 {
				t.Fatalf("owned = %+v, want exactly the later valid note of 5", owned)
			}

			if len(rejected) != 1 {
				t.Fatalf("rejected = %+v, want exactly one rejection", rejected)
			}

			if rejected[0].Index != 0 || rejected[0].Reason != core.RejectBadCapsule {
				t.Fatalf("rejected[0] = %+v, want index 0 with reason %q",
					rejected[0], core.RejectBadCapsule)
			}
		})
	}
}

func TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners(t *testing.T) {
	corrupt := func(t *testing.T, ciphertext []byte) []byte {
		t.Helper()
		if len(ciphertext) != 1088 {
			t.Fatalf("ciphertext length = %d, want 1088", len(ciphertext))
		}
		mutated := append([]byte(nil), ciphertext...)
		mutated[0] ^= 0x01
		return mutated
	}

	t.Run("ERC20", func(t *testing.T) {
		w := newWallet(t)
		bad := goodErc20Event(t, w, big.NewInt(10), big.NewInt(6))
		bad.CipherText = corrupt(t, bad.CipherText)
		good := goodErc20Event(t, w, big.NewInt(10), big.NewInt(5))

		owned, rejected, err := core.ScanForErc20NotesWithRejected(
			w.view.DecapsKey, w.spend.PublicKey,
			[]core.OnChainErc20Event{bad, good},
		)
		if err != nil || len(owned) != 1 || owned[0].Amount.Cmp(big.NewInt(5)) != 0 || len(rejected) != 0 {
			t.Fatalf("owned=%+v rejected=%+v err=%v", owned, rejected, err)
		}
	})

	t.Run("ERC721", func(t *testing.T) {
		w := newWallet(t)
		bad := goodErc721Event(t, w, big.NewInt(6))
		bad.CipherText = corrupt(t, bad.CipherText)
		good := goodErc721Event(t, w, big.NewInt(5))

		owned, rejected, err := core.ScanForErc721NotesWithRejected(
			w.view.DecapsKey, w.spend.PublicKey,
			[]core.OnChainErc721Event{bad, good},
		)
		if err != nil || len(owned) != 1 || owned[0].TokenId.Cmp(big.NewInt(5)) != 0 || len(rejected) != 0 {
			t.Fatalf("owned=%+v rejected=%+v err=%v", owned, rejected, err)
		}
	})

	t.Run("ERC1155", func(t *testing.T) {
		w := newWallet(t)
		bad := goodErc1155Event(t, w, big.NewInt(6))
		bad.CipherText = corrupt(t, bad.CipherText)
		good := goodErc1155Event(t, w, big.NewInt(5))

		owned, rejected, err := core.ScanForErc1155NotesWithRejected(
			w.view.DecapsKey, w.spend.PublicKey,
			[]core.OnChainErc1155Event{bad, good},
		)
		if err != nil || len(owned) != 1 || owned[0].Amount.Cmp(big.NewInt(5)) != 0 || len(rejected) != 0 {
			t.Fatalf("owned=%+v rejected=%+v err=%v", owned, rejected, err)
		}
	})

	t.Run("ZK_DvP", func(t *testing.T) {
		alice, bob := newWallet(t), newWallet(t)
		bad := goodSwapEvent(t, alice, bob, big.NewInt(6))
		bad.CipherText = corrupt(t, bad.CipherText)
		good := goodSwapEvent(t, alice, bob, big.NewInt(5))

		swaps, rejected, err := core.ScanForZkDvpSwapWithRejected(
			bob.view.DecapsKey,
			alice.spend.PublicKey,
			bob.spend.PublicKey,
			swapAmountIn,
			swapTokenIn,
			[]core.OnChainZkDvpEvent{bad, good},
		)
		if err != nil || len(swaps) != 1 || swaps[0].AmountOut.Cmp(big.NewInt(5)) != 0 || len(rejected) != 0 {
			t.Fatalf("swaps=%+v rejected=%+v err=%v", swaps, rejected, err)
		}
	})
}
