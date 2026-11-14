package engine

import (
	"context"
	"math/big"
	"testing"

	tags "github.com/raylsnetwork/enygma_retail_payments/private_tags/src"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type fakeTagLogReader struct {
	tip         uint64
	logs        []ethtypes.Log
	filterCalls []ethereum.FilterQuery
}

func (f *fakeTagLogReader) BlockNumber(context.Context) (uint64, error) {
	return f.tip, nil
}

func (f *fakeTagLogReader) FilterLogs(_ context.Context, q ethereum.FilterQuery) ([]ethtypes.Log, error) {
	f.filterCalls = append(f.filterCalls, q)
	return append([]ethtypes.Log(nil), f.logs...), nil
}

func makeTagPublishedTestLog(t *testing.T, block uint64, channel tags.Channel, payload []byte) ethtypes.Log {
	t.Helper()

	tag, err := tags.DeriveTag(block, channel.PkSpend, channel.SharedSecret)
	if err != nil {
		t.Fatalf("DeriveTag: %v", err)
	}

	data, err := abiBytes().Pack(payload)
	if err != nil {
		t.Fatalf("ABI-pack payload: %v", err)
	}

	return ethtypes.Log{
		BlockNumber: block,
		Topics: []common.Hash{
			tagPublishedTopic,
			common.BigToHash(new(big.Int).SetUint64(block)),
			common.BytesToHash(tag[:]),
			common.BytesToHash([]byte{1}),
		},
		Data: data,
	}
}

func TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent(t *testing.T) {
	const blockNumber = uint64(10)

	channel := tags.Channel{
		SharedSecret: []byte("0123456789abcdef0123456789abcdef"),
		PkSpend:      big.NewInt(12345),
	}

	// Construct a valid 96-byte payment note and encrypt it for this
	// channel and block. This is the event that follows the malformed one.
	amount := big.NewInt(5)
	tokenID := big.NewInt(7)
	salt := big.NewInt(9)

	plaintext := make([]byte, 96)
	amount.FillBytes(plaintext[0:32])
	tokenID.FillBytes(plaintext[32:64])
	salt.FillBytes(plaintext[64:96])

	blockKey := tags.DeriveBlockKey(channel.SharedSecret, blockNumber, channel.PkSpend)
	validPayload, err := tags.EncryptPayload(blockKey, plaintext)
	if err != nil {
		t.Fatalf("EncryptPayload: %v", err)
	}

	// Both events have the correct recipient tag. The first payload is
	// deliberately too short for AES-GCM decryption.
	malformedLog := makeTagPublishedTestLog(t, blockNumber, channel, []byte{0x01})
	validLog := makeTagPublishedTestLog(t, blockNumber, channel, validPayload)

	// Prove the later event is independently valid and decryptable.
	opened, _, err := openTag(validLog, channel)
	if err != nil {
		t.Fatalf("control: later valid event did not decrypt: %v", err)
	}
	if opened.Amount.Cmp(amount) != 0 ||
		opened.TokenId.Cmp(tokenID) != 0 ||
		opened.Salt.Cmp(salt) != 0 {
		t.Fatalf("control: decrypted note differs from fixture")
	}

	reader := &fakeTagLogReader{
		tip:  blockNumber,
		logs: []ethtypes.Log{malformedLog, validLog},
	}
	e := &Engine{
		chain: &chain{tagReader: reader},
	}
	a := &Actor{
		Name: "recipient",
		tags: tagState{
			in:        []tags.Channel{channel},
			lastBlock: blockNumber - 1,
		},
	}

	_, checked, err := e.scanTags(context.Background(), a)
	if err == nil {
		t.Fatal("expected malformed matching event to abort the scan")
	}
	if checked != 1 {
		t.Fatalf("checked %d events before abort, want 1", checked)
	}
	if a.tags.lastBlock != blockNumber-1 {
		t.Fatalf("scan cursor advanced to %d after failure, want %d", a.tags.lastBlock, blockNumber-1)
	}

	// A retry must request the same block range because the failed scan
	// did not advance the cursor.
	_, retryChecked, retryErr := e.scanTags(context.Background(), a)
	if retryErr == nil {
		t.Fatal("expected retry to encounter the same malformed event")
	}
	if retryChecked != 1 {
		t.Fatalf("retry checked %d events before abort, want 1", retryChecked)
	}
	if len(reader.filterCalls) != 2 {
		t.Fatalf("FilterLogs called %d times, want 2", len(reader.filterCalls))
	}

	for i, q := range reader.filterCalls {
		if q.FromBlock == nil || q.ToBlock == nil ||
			q.FromBlock.Cmp(big.NewInt(int64(blockNumber))) != 0 ||
			q.ToBlock.Cmp(big.NewInt(int64(blockNumber))) != 0 {
			t.Fatalf("query %d used range [%v, %v], want [%d, %d]",
				i, q.FromBlock, q.ToBlock, blockNumber, blockNumber)
		}
	}

	if a.tags.lastBlock != blockNumber-1 {
		t.Fatalf("retry advanced cursor to %d after failure, want %d", a.tags.lastBlock, blockNumber-1)
	}
}
