package tests

import (
	"bytes"
	"context"
	"math/big"
	"sort"
	"testing"

	core "github.com/raylsnetwork/enygma_dvp/src/core"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TestEXP_MLK_006_MalformedCapsuleDepositRuntimeReachability tests whether
// Erc20CoinVault.depositV2 accepts malformed ML-KEM capsule lengths, emits
// those exact bytes in EncryptedNote, and exposes the events through FilterLogs.
//
// It also feeds the retrieved events to the production ERC20 scanner and
// verifies that malformed events are rejected without preventing recovery
// of a later valid note.
//
// Requires an already-running, initialized local Payment deployment.
// This test does not deploy or initialize contracts.
func TestEXP_MLK_006_MalformedCapsuleDepositRuntimeReachability(t *testing.T) {
	if !chainAvailable() {
		t.Skip("Hardhat node not running on localhost:8545")
	}

	ctx := context.Background()
	client, err := ethclient.Dial(hardhatRPC)
	if err != nil {
		t.Fatalf("ethclient.Dial: %v", err)
	}
	defer client.Close()

	receipts := loadPaymentReceipts(t)
	dvpAddr := common.HexToAddress(receipts["EnygmaDvp"].ContractAddress)
	vaultAddr := common.HexToAddress(receipts["Erc20CoinVault"].ContractAddress)
	tokenAddr := common.HexToAddress(receipts["ERC20"].ContractAddress)

	if dvpAddr == (common.Address{}) ||
		vaultAddr == (common.Address{}) ||
		tokenAddr == (common.Address{}) {
		t.Fatal("Payment deployment receipt contains a zero contract address")
	}

	skipUnlessPaymentDeployment(t, ctx, client, dvpAddr)

	vaultABI := loadOnchainABI(
		t,
		"core/contracts/vaults/Erc20CoinVault.sol/Erc20CoinVault.json",
	)
	tokenABI := loadOnchainABI(
		t,
		"erc20/contracts/RaylsERC20.sol/RaylsERC20.json",
	)

	vault := bind.NewBoundContract(vaultAddr, vaultABI, client, client, client)
	token := bind.NewBoundContract(tokenAddr, tokenABI, client, client, client)
	owner := hardhatAuth(t, client)

	spend, err := core.NewSpendKeyPair()
	if err != nil {
		t.Fatalf("NewSpendKeyPair: %v", err)
	}
	view, err := core.NewViewKeyPair()
	if err != nil {
		t.Fatalf("NewViewKeyPair: %v", err)
	}

	const (
		vaultID = 0
		tokenID = 0
	)

	amount := big.NewInt(1)
	totalAmount := big.NewInt(5)
	tokenIDValue := big.NewInt(tokenID)

	// Fund and approve all five deposits in one setup sequence.
	mintTx, err := token.Transact(owner, "mint", owner.From, totalAmount)
	if err != nil {
		t.Fatalf("ERC20.mint: %v", err)
	}
	mintReceipt, err := bind.WaitMined(ctx, client, mintTx)
	if err != nil || mintReceipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("ERC20.mint failed (receipt=%v err=%v)", mintReceipt, err)
	}

	approveTx, err := token.Transact(owner, "approve", vaultAddr, totalAmount)
	if err != nil {
		t.Fatalf("ERC20.approve: %v", err)
	}
	approveReceipt, err := bind.WaitMined(ctx, client, approveTx)
	if err != nil || approveReceipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("ERC20.approve failed (receipt=%v err=%v)", approveReceipt, err)
	}

	// Start the query range after funding and approval, so retrieved events
	// belong to this experiment rather than earlier deployment activity.
	startHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("HeaderByNumber before deposits: %v", err)
	}
	fromBlock := new(big.Int).Add(startHeader.Number, big.NewInt(1))

	// The final case is the valid 1088-byte control.
	cases := []struct {
		label           string
		submittedLength int
	}{
		{label: "empty", submittedLength: 0},
		{label: "one-byte", submittedLength: 1},
		{label: "short", submittedLength: 1087},
		{label: "oversized", submittedLength: 1089},
		{label: "valid-control", submittedLength: 1088},
	}

	expectedCommitments := make([]*big.Int, 0, len(cases))
	expectedCapsules := make([][]byte, 0, len(cases))
	expectedPayloads := make([][]byte, 0, len(cases))
	txHashes := make([]common.Hash, 0, len(cases))
	var lastReceipt *types.Receipt

	for _, tc := range cases {
		ss, validCapsule, err := core.Encapsulate(view.EncapsKey)
		if err != nil {
			t.Fatalf("%s: Encapsulate: %v", tc.label, err)
		}

		saltB, err := core.DerivePaymentSalt(ss)
		if err != nil {
			t.Fatalf("%s: DerivePaymentSalt: %v", tc.label, err)
		}
		encKey, err := core.DerivePaymentKey(ss)
		if err != nil {
			t.Fatalf("%s: DerivePaymentKey: %v", tc.label, err)
		}

		salt := core.SaltBToField(saltB)
		commitment, err := core.Erc20CommitmentV2(
			spend.PublicKey,
			salt,
			amount,
			tokenIDValue,
		)
		if err != nil {
			t.Fatalf("%s: Erc20CommitmentV2: %v", tc.label, err)
		}

		payload, err := core.EncryptPayload(encKey, tokenIDValue, amount)
		if err != nil {
			t.Fatalf("%s: EncryptPayload: %v", tc.label, err)
		}

		submittedCapsule := validCapsule
		if tc.submittedLength != 1088 {
			// Change only ciphertextI. The commitment and encrypted payload
			// are still derived from the valid encapsulation above.
			submittedCapsule = make([]byte, tc.submittedLength)
		}

		tx, err := vault.Transact(
			owner,
			"depositV2",
			[]*big.Int{amount, spend.PublicKey, salt, tokenIDValue},
			submittedCapsule,
			payload,
		)
		if err != nil {
			t.Fatalf("%s: depositV2 submission failed: %v", tc.label, err)
		}

		mined, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			t.Fatalf("%s: waiting for depositV2: %v", tc.label, err)
		}
		if mined.Status != types.ReceiptStatusSuccessful {
			t.Fatalf(
				"%s: depositV2 reverted: tx=%s block=%v status=%d",
				tc.label,
				tx.Hash().Hex(),
				mined.BlockNumber,
				mined.Status,
			)
		}

		expectedCommitments = append(expectedCommitments, new(big.Int).Set(commitment))
		expectedCapsules = append(expectedCapsules, append([]byte(nil), submittedCapsule...))
		expectedPayloads = append(expectedPayloads, append([]byte(nil), payload...))
		txHashes = append(txHashes, tx.Hash())
		lastReceipt = mined

		t.Logf(
			"deposit accepted: case=%s capsule_len=%d tx=%s block=%s",
			tc.label,
			len(submittedCapsule),
			tx.Hash().Hex(),
			mined.BlockNumber.String(),
		)
	}

	if lastReceipt == nil {
		t.Fatal("no deposit receipts were produced")
	}

	// Query the chain for the EncryptedNote events, matching the indexed
	// vaultId in the same way a scanner queries vault-specific notes.
	encryptedNoteTopic := crypto.Keccak256Hash(
		[]byte("EncryptedNote(uint256,uint256,bytes,bytes)"),
	)
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   new(big.Int).Set(lastReceipt.BlockNumber),
		Addresses: []common.Address{vaultAddr},
		Topics: [][]common.Hash{
			{encryptedNoteTopic},
			{common.BigToHash(big.NewInt(vaultID))},
		},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		t.Fatalf("FilterLogs for EncryptedNote: %v", err)
	}

	// Sort explicitly by chain position before comparing against submission order.
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].BlockNumber != logs[j].BlockNumber {
			return logs[i].BlockNumber < logs[j].BlockNumber
		}
		if logs[i].TxIndex != logs[j].TxIndex {
			return logs[i].TxIndex < logs[j].TxIndex
		}
		return logs[i].Index < logs[j].Index
	})

	if len(logs) != len(cases) {
		t.Fatalf(
			"FilterLogs returned %d EncryptedNote events; expected %d",
			len(logs),
			len(cases),
		)
	}

	event, err := vaultABI.EventByID(encryptedNoteTopic)
	if err != nil {
		t.Fatalf("vault ABI EncryptedNote event lookup: %v", err)
	}

	scanEvents := make([]core.OnChainErc20Event, 0, len(logs))

	for i, lg := range logs {
		if len(lg.Topics) != 3 {
			t.Fatalf("event %d: expected 3 topics, got %d", i, len(lg.Topics))
		}
		if lg.Topics[1] != common.BigToHash(big.NewInt(vaultID)) {
			t.Fatalf("event %d: unexpected indexed vault ID topic", i)
		}

		values, err := event.Inputs.NonIndexed().Unpack(lg.Data)
		if err != nil {
			t.Fatalf("event %d: decode EncryptedNote data: %v", i, err)
		}
		if len(values) != 2 {
			t.Fatalf("event %d: expected 2 non-indexed values, got %d", i, len(values))
		}

		emittedCapsule, ok := values[0].([]byte)
		if !ok {
			t.Fatalf("event %d: ciphertextI has unexpected ABI type %T", i, values[0])
		}
		emittedPayload, ok := values[1].([]byte)
		if !ok {
			t.Fatalf("event %d: ciphertextII has unexpected ABI type %T", i, values[1])
		}

		emittedCommitment := lg.Topics[2].Big()
		if emittedCommitment.Cmp(expectedCommitments[i]) != 0 {
			t.Fatalf("event %d: emitted commitment differs from submitted note commitment", i)
		}
		if len(emittedCapsule) != cases[i].submittedLength {
			t.Fatalf(
				"event %d (%s): emitted capsule length=%d; expected=%d",
				i,
				cases[i].label,
				len(emittedCapsule),
				cases[i].submittedLength,
			)
		}
		if !bytes.Equal(emittedCapsule, expectedCapsules[i]) {
			t.Fatalf("event %d (%s): emitted capsule bytes differ from submitted bytes", i, cases[i].label)
		}
		if !bytes.Equal(emittedPayload, expectedPayloads[i]) {
			t.Fatalf("event %d (%s): emitted ciphertextII differs from submitted payload", i, cases[i].label)
		}

		scanEvents = append(scanEvents, core.OnChainErc20Event{
			Commitment: emittedCommitment,
			CipherText: emittedCapsule,
			EncTxData:  emittedPayload,
		})

		t.Logf(
			"event retrieved: case=%s tx=%s block=%d log_index=%d capsule_len=%d payload_len=%d",
			cases[i].label,
			lg.TxHash.Hex(),
			lg.BlockNumber,
			lg.Index,
			len(emittedCapsule),
			len(emittedPayload),
		)
		if lg.TxHash != txHashes[i] {
			t.Fatalf("event %d: transaction hash does not match corresponding deposit", i)
		}
	}

	// The four malformed events should be rejected individually. The final
	// valid event must still be recovered by the same scanner invocation.
	owned, rejected, err := core.ScanForErc20NotesWithRejected(
		view.DecapsKey,
		spend.PublicKey,
		scanEvents,
	)
	if err != nil {
		t.Fatalf("ScanForErc20NotesWithRejected returned fatal error: %v", err)
	}

	if len(rejected) != 4 {
		t.Fatalf("scanner rejected %d events; expected 4", len(rejected))
	}
	for i, rejection := range rejected {
		if rejection.Index != i {
			t.Errorf("rejection %d has event index %d; expected %d", i, rejection.Index, i)
		}
		if rejection.Reason != core.RejectBadCapsule {
			t.Errorf(
				"event %d rejection reason=%q; expected %q",
				i,
				rejection.Reason,
				core.RejectBadCapsule,
			)
		}
	}

	if len(owned) != 1 {
		t.Fatalf("scanner recovered %d owned notes; expected 1 valid control note", len(owned))
	}
	if owned[0].Commitment.Cmp(expectedCommitments[4]) != 0 {
		t.Fatal("scanner recovered a note with a commitment other than the valid control")
	}
	if owned[0].Amount.Cmp(amount) != 0 {
		t.Fatalf("recovered amount=%s; expected=%s", owned[0].Amount, amount)
	}
	if owned[0].TokenId.Cmp(tokenIDValue) != 0 {
		t.Fatalf("recovered tokenId=%s; expected=%s", owned[0].TokenId, tokenIDValue)
	}

	t.Log("EXP-MLK-006 PASS")
	t.Log("All five depositV2 transactions mined successfully.")
	t.Log("FilterLogs retrieved all five EncryptedNote events with exact submitted bytes.")
	t.Log("Scanner rejected all four malformed capsules as bad capsule.")
	t.Log("Scanner recovered the later valid control note without a fatal error.")
}
