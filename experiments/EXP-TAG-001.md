# EXP-TAG-001 — Malformed Matching Private-Tag Event Aborts Scanning

## 1. Experiment status

**Status:** PASS — focused regression test executed successfully on 2026-11-15.

**Subsystem:** `enygma_retail_payments/shield_demo/engine` private-tag notification scanner.

**Official source repository:** `https://github.com/raylsnetwork/Enygma`

**Recorded official source revision:** `79c82dfb57595681b03d8da344702ef606f3a621`

**Test source path:** `enygma_retail_payments/shield_demo/engine/tags_test.go`

**Test function:** `TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent`

**Preserved research artifact:** `enygma_retail_payments/shield_demo/engine/tags_test.go`

**Test artifact SHA-256:** `74de6a796f951f876fafa226d63802a8dd72a1748888a105384d149d910f1530`

The test artifact is research-authored and was preserved in the research repository. It was untracked in the official Enygma working tree at the time of the recorded investigation. It must not be represented as a test included in the recorded upstream commit.

## 2. Hypothesis and research question

**Hypothesis:** A malformed `TagPublished` event whose tag matches a recipient's private-tag channel can abort `scanTags()` before a later valid event in the same fetched log batch is processed.

If the scan cursor is not advanced after the error, a subsequent scan may request the same block range and encounter the same malformed event again.

The experiment tests this scanner behaviour with a controlled log-reader fixture. It does not test whether an arbitrary external party can publish a matching event on a deployed network.

## 3. Test construction

The test uses an injected `fakeTagLogReader` implementing the `tagLogReader` interface.

It constructs a recipient channel with a shared secret and spend public key, then creates two `TagPublished` log fixtures for block 10:

1. A malformed matching event whose ABI-encoded payload contains a single byte (`0x01`).
2. A later matching event carrying an independently valid encrypted payment note.

The valid note contains:

| Field | Value |
|---|---:|
| Amount | 5 |
| Token ID | 7 |
| Salt | 9 |
| Plaintext length | 96 bytes |

Before testing the scanner, the test calls `openTag()` directly on the later event and checks that the decrypted amount, token ID, and salt match the fixture. This establishes that the later event is independently decryptable.

The fake reader returns the malformed event first and the valid event second. Its initial chain tip is block 10, while the actor's scan cursor is block 9.

## 4. Reproduction

Run from the official Enygma source checkout, in the shield-demo module:

```bash
cd /workspaces/Enygma/enygma_retail_payments/shield_demo
GOTOOLCHAIN=local go test ./engine -run '^TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent$' -v -count=1
```

**Recorded result:** PASS.

The focused test execution is preserved in:

`evidence/focused-test-execution-2026-11-15.md`

See the [raw focused execution transcript](../evidence/focused-test-execution-2026-11-15.md#L115).

The transcript records the focused test execution; it is not evidence of a live-chain test.

## 5. Observed results

The test asserts all of the following:

- The first malformed matching event causes `scanTags()` to return an error.
- Exactly one event is checked before the scan aborts.
- The later valid event is not processed during that scan.
- The actor's cursor remains at block 9.
- A retry also fails after checking one event.
- The fake reader receives two `FilterLogs` calls.
- Both calls request the same block range: block 10 through block 10.
- The cursor remains at block 9 after the retry.

The later valid event was independently verified as decryptable before the scanner test.

## 6. Source-level failure path

The inspected `scanTags()` implementation performs these operations:

1. Reads the chain tip through `tagReader.BlockNumber`.
2. Fetches `TagPublished` logs from the actor's last scanned block plus one through the tip.
3. Iterates over the returned logs.
4. Calls `openTag()` for each log and recipient channel.
5. Returns immediately when `openTag()` returns an error.
6. Advances `a.tags.lastBlock` to the tip only after the complete log loop succeeds.

In `openTag()`, a matching event is ABI-decoded and passed to `tags.DecryptPaymentNote`. An ABI decoding or payment-note decryption error is returned to the caller.

Consequently, in the tested sequence, the malformed matching event triggers the immediate error return. The later valid event is not reached, and the cursor-update statement is not executed.

The test directly checks the resulting cursor state and repeated query range.

## 7. Provenance and test seam

The test uses an interface introduced in the local working tree:

```go
type tagLogReader interface {
    BlockNumber(context.Context) (uint64, error)
    FilterLogs(context.Context, ethereum.FilterQuery) ([]ethtypes.Log, error)
}
```

The `chain` struct contains a `tagReader` field. The inspected production initialization in `dialChain()` assigns the real Ethereum client to it:

```go
c.client = client
c.tagReader = client
```

The test instead injects `fakeTagLogReader` through `chain{tagReader: reader}`.

The local source diffs inspected during this investigation show:

- `chain.go` adds the `tagLogReader` interface and `tagReader` field, and assigns the real client to the interface during production initialization.
- `tags.go` routes the relevant `BlockNumber` and `FilterLogs` calls through `tagReader`.

The recorded official source revision is `79c82dfb57595681b03d8da344702ef606f3a621`. The local working-tree versions of `chain.go` and `tags.go` differed from that revision when inspected:

| File | SHA-256 at recorded revision | Local working-tree SHA-256 |
|---|---|---|
| `engine/tags.go` | `3ded4059648c59933a478b669c2d81a4ce2f7213dbf2c2367e7da96faaf34ce7` | `4a55e1abc8111840250ff4387814a671399e961d72846eb12ace87777ef39304` |
| `engine/chain.go` | `2de59abdec987d90c579358c4c93177d55077393eb499ff40588c556ee36272f` | `51c61a980d2321ca4da2def2535a6881aef31e55e07fed91f21c6f69dc7c55c4` |

Accordingly, this experiment is described as a research-authored regression test executed against the local working tree with the test seam present. It is not described as execution against an untouched checkout of the recorded upstream revision.

## 8. Limitations and interpretation

This experiment establishes the tested scanner behaviour under a controlled fixture:

- A malformed event with a matching recipient tag can abort the current scan.
- A later valid event in the same fetched batch is not processed during that invocation.
- The cursor remains unchanged after the failure.
- Retrying the scan requests the same block range and encounters the malformed event again.

It does **not** establish:

- That arbitrary external parties can construct or publish an event matching a particular recipient's private tag.
- That an attacker can trigger this condition against a real recipient on a production network.
- That the condition causes a remote denial of service or system-wide outage.
- That funds can be stolen, lost, or double-spent.
- That all private-tag event failures have the same behaviour.

The test uses a fake log reader and a controlled pair of event fixtures. It is not a live Ethereum RPC or deployed-contract reproduction.

**Assessment:** EXP-TAG-001 demonstrates a reproducible scan-abort and cursor-retry behaviour for a malformed matching private-tag event. Production reachability and security impact remain unestablished.

<!-- EXP-TAG-001-consolidated-evidence -->
## 9. Consolidated execution and submission evidence

**Focused test:** `TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent`

**Execution record:** The exact reproduction command and its recorded Go test output are preserved in the shared transcript:

[Focused test execution transcript](../evidence/focused-test-execution-2026-11-15.md#L115)

**Transcript SHA-256:**

`c7fc9ff0d277c8fda99570a2775129dd6ad416e35e7dfec7c063a2152512322b`

**Evidence date:** 2026-11-15.

**Recorded walkthrough and submission context:** The researcher confirms that the video recording showing this experiment's command and terminal output is included in the corresponding experiment section of the submitted X article. The article also references the GitHub research repository for reviewer access to the reproducibility materials. The recording is existing submission evidence; it is not being recreated or represented as a separate repository artifact.

**Evidence boundary:** The transcript documents the focused test execution and its observed output. The conclusions and limitations remain those stated in this experiment report; the test result alone does not establish production exploitability or an on-chain attack.
