# ML-KEM-768 Adversarial Scanner Research — Milestone Register

## 1. Research objective

Independently investigate how Rayls Enygma handles adversarial ML-KEM-768 ciphertexts.

The investigation covers:
- Incorrect ciphertext lengths.
- Correct-length but corrupted ciphertexts.
- Recovery of valid events after malformed events.
- Rejection reporting and scanner errors.
- ERC20, ERC721, ERC1155 and ZK DvP scanner coverage.
- Production input reachability.

This research does not establish system-wide post-quantum security, production exploitability, remote denial of service, or formal FIPS/CMVP validation.

## 2. Repository separation

**Official source repository**
- Path: `/workspaces/Enygma`
- Source under test: `enygma_dvp/src`
- Branch: `main`
- Revision: `79c82dfb57595681b03d8da344702ef606f3a621`

**Private research repository**
- Path: `/workspaces/rayls-enygma-mlkem-adversarial-research`
- GitHub: `Abasiekemeobotrealms/rayls-enygma-mlkem-adversarial-research`
- Visibility: Private

The research repository is separate from the official source repository.

## 3. Recorded environment

- Go: `go1.27.3 linux/amd64`
- Node.js: `v24.22.2`
- npm: `11.19.1`
- Git: `2.56.0`
- GitHub CLI: `2.103.3`

## 4. Milestone M-001 — Shared scanner implementation

**Status:** Source inspection completed.

Inspected `enygma_dvp/src/core/scan.go`.

The shared `scanEvents()` function:
1. Iterates through the supplied events.
2. Calls `Decapsulate()` for each event capsule.
3. Records `RejectBadCapsule` and continues when decapsulation returns an error.
4. Passes a successfully decapsulated shared secret to the scanner-specific opening function.
5. Collects successfully opened notes or swaps.
6. Silently skips `errNotAddressed`.
7. Records typed `RejectReason` errors.
8. Returns a fatal scan error for unexpected errors.

Rejection reasons observed in source:
- `RejectBadCapsule = "bad capsule"`
- `RejectCommitmentMismatch = "commitment mismatch"`
- `RejectCommitmentNotComputable = "commitment not computable"`

All four scanner families inspected:
- ERC20
- ERC721
- ERC1155
- ZK DvP

## 5. Milestone M-002 — ML-KEM ciphertext semantics

**Status:** Source and standard-library documentation inspected.

Inspected `core.Decapsulate()` in `enygma_dvp/src/core/utils.go` and Go documentation for `crypto/mlkem.DecapsulationKey768.Decapsulate`.

Observed:
- ML-KEM-768 ciphertext size is 1088 bytes (`CiphertextSize768`).
- Incorrect ciphertext length returns a decapsulation error.
- Correct-length but invalid ciphertext follows implicit rejection: decapsulation returns a shared secret without an error, but it does not match the sender's shared secret.

Research implication: malformed-length and correctly sized corrupted ciphertexts exercise different branches in the scanner. They must be tested separately.


## 6. Milestone M-003 — Adversarial scanner regression experiments

**Status:** Passed. Permanent tests executed successfully.

Test file:
`/workspaces/Enygma/enygma_dvp/src/tests/scan_mlkem_adversarial_test.go`

### EXP-MLK-002 — Invalid ciphertext lengths

Test: `TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated`

Tested lengths: 0, 1, 1087 and 1089 bytes.

Results:
- No fatal scanner error.
- Each malformed event was reported as `RejectBadCapsule`.
- The valid ERC20 event following it was recovered.
- Exactly one rejection was reported at index 0.

### EXP-MLK-003 — Correct-length corruption in ERC20

Test: `TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote`

A valid 1088-byte ciphertext was mutated using `CipherText[0] ^= 0x01`.

Results:
- No fatal scanner error.
- The later valid note, amount 5, was recovered.
- No rejection was reported for the corrupted event.

### EXP-MLK-004 — Correct-length corruption across scanner families

Test: `TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners`

Tested ERC20, ERC721, ERC1155 and ZK DvP.

Results:
- All four scanner subtests passed.
- Each recovered the later valid event.
- No fatal scan error or rejection was reported for the corrupted event.

### Interpretation and limitations

Incorrect-length ciphertexts cause decapsulation errors and are reported as bad capsules. The tested one-bit, correct-length corruption follows ML-KEM implicit rejection; the resulting incorrect shared secret causes payload opening to fail, and the event is silently skipped.

These findings apply to the tested fixtures and mutation. They do not establish behaviour for every possible corrupted ciphertext or prove remote denial of service or production exploitability.


## 7. Milestone M-004 — Encrypted-note ciphertext substitution

**Status:** Existing integration-test source inspected; independent execution not completed.

**Source file:**
`/workspaces/Enygma/enygma_dvp/test/26_v2_erc20_partial_withdraw_test.go`

**Test:**
`TestV2Erc20PartialWithdraw_RecoverChangeWithoutCiphertext`

### Test scenarios

The existing test creates a partial withdrawal from a note worth 100:
- Withdrawal: 10.
- Change note: 90.

It tests two settlement variants:
1. `partialWithdraw` with junk ciphertexts (`0xde, 0xad` and `0xbe, 0xef`).
2. `withdrawV2`, which omits the change-note ciphertexts.

The test checks successful settlement and reconstructs the change commitment.

### Recovery and spendability

The test then:
1. Reconstructs the change salt using `WithdrawChangeNote()`.
2. Reconstructs the commitment and compares it with the emitted change commitment.
3. Creates a recovered note worth 90.
4. Generates a new withdrawal proof for that note.
5. Submits the subsequent withdrawal.
6. Checks that the recipient receives 100 in total: 10 initially and the recovered 90.

The test therefore asserts that the reconstructed change note is spendable.


### Provenance and limitations

The test file last changed in commit `ebc4de585311111149084115246451db9dd04af3` (2026-10-21).

Independent execution was not possible in the inspected environment: the local Hardhat node, proof server, dependencies, compiled artifacts, deployment data and receipts were unavailable.

This is evidence of a documented ERC20 partial-withdrawal ciphertext-substitution path in an existing regression test. It does not establish fund theft, loss, double-spending, or exploitability across other settlement paths.


## 8. EXP-TAG-001 — Malformed private-tag event aborts the scan

**Status:** Passed. Focused regression test and full engine package suite passed.

**Subsystem:** `enygma_retail_payments/shield_demo/engine` private-tag notification scanner.

**Official source files involved:** `chain.go`, `tags.go`, and `tags_test.go`. The research copy of `tags_test.go` was verified byte-for-byte against the official source.

**SHA-256:** `74de6a796f951f876fafa226d63802a8dd72a1748888a105384d149d910f1530`

**Test:** `TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent`.

**Hypothesis:** A malformed `TagPublished` event with a tag matching the channel can abort scanning before a later valid event in the same fetched batch is processed. If the cursor is not advanced, a retry may fetch the same malformed event again.

**Observed result:** The malformed matching event caused `scanTags()` to return an error. The later valid event was not processed in that scan. The cursor remained unchanged, and retry queried the same block range. The later event was independently verified as decryptable.

**Execution:** The focused regression test passed, and `go test ./engine -count=1 -v` passed the complete engine package suite.

**Limitations:** This demonstrates the behavior for the tested event fixture. A matching tag is a precondition; this does not prove arbitrary external parties can create matching events, remote denial of service, production exploitability, asset loss, or system-wide outage. The `tagLogReader` interface is a test seam; production assigns the real Ethereum client.


## 9. EXP-FIELD-001 — Negative inputs accepted as BN254 field representatives

**Status:** Native commitment, Gnark circuit, and HTTP handler test evidence preserved. The relevant tests passed in the recorded executions.

**Subsystems:** `enygma_dvp/src`, `enygma_dvp/gnark_circuits`.

**BN254 scalar-field modulus:**

`21888242871839275222246405745257275088548364400416034343698204186575808495617`

### Preserved test artifacts

| Layer | Research artifact | SHA-256 |
|---|---|---|
| Native commitment | `tests/commitment_field_boundary_test.go` | `5a927a0237ea29d8e8985f08cd3e34d320cd31075acfa8effc405dbfa61ed4ec` |
| Gnark circuit | `tests/payment_circuit_field_boundary_test.go` | `eea035ee1c7e6bd1de96e676d2e691ccabbd43e3ad7955e353a284408da32375` |
| HTTP handler | `tests/field_boundary_handler_test.go` | `8959c22dd02cc290760c234f864a04862112babae33ab5b8d6fe24e613d14125` |

All three research copies were compared byte-for-byte with their corresponding official source files.

### Hypothesis

Some Go commitment and proof-generation paths may accept negative integers as inputs to BN254 field operations rather than enforcing canonical nonnegative field representations at their boundaries. If so, a negative integer such as `-1` may be interpreted as its field-equivalent value `Q-1`.

### Observed results

1. **Native commitment:** `TestErc20CommitmentV2_FieldBoundary` demonstrated that `-1` in the tested salt, amount, and token-ID positions hashes identically to `Q-1`. The test also rejected `Q` and `Q+1` in the tested salt position.
2. **Gnark circuit:** `TestPaymentCircuit_NegativeInputSaltFieldEquivalentToQMinusOne` and `TestPaymentCircuit_NegativeTokenIdFieldEquivalentToQMinusOne` demonstrated successful proofs for the tested negative values, equivalent to `Q-1`, under BN254 Groth16 and PLONK.
3. **HTTP handler:** `TestHandler_AcceptsNegativeInputSaltFieldEquivalentToQMinusOne` and `TestHandler_AcceptsNegativeTokenIdFieldEquivalentToQMinusOne` demonstrated that the tested negative JSON values were accepted by the handler and passed into its witness path.

The native behavior is consistent with the inspected `go-iden3-crypto` field check: `CheckBigIntInField` checks whether the integer is less than `Q`, without separately rejecting negative values. Field arithmetic then operates modulo `Q`.

### Execution evidence

- Native test: `go test ./tests -run "^TestErc20CommitmentV2_FieldBoundary$" -count=1 -v` — passed.
- Handler package: `go test ./server/circuits/payment -count=1 -v` — passed, including both new handler tests and the existing package test.
- Circuit tests: both negative-input tests passed under Groth16 and PLONK in the recorded executions.

### Limitations and interpretation

- The native test demonstrates behavior of the tested commitment helper; it does not establish that every production caller can supply negative values.
- The Payment circuit constrains amounts separately. The negative salt and token-ID results must not be generalized to negative amount acceptance by that circuit.
- Handler acceptance does not by itself prove that a production client can reach the handler with such values, or that a deployed contract accepts a resulting transaction.
- No asset loss, unauthorized spend, forged commitment, or other exploit impact has been established.
- This is a field-representation and input-validation finding requiring further caller-reachability and end-to-end contract analysis.

**Assessment:** A reproducible negative-integer/canonical-field-equivalence behavior exists in the tested native, circuit, and handler paths. Whether it constitutes a security vulnerability depends on the missing production reachability and downstream validation evidence.

## 10. EXP-FIELD-002 — Native ERC-20 scanner accepts token ID R-1

**Status:** Executed successfully. The focused scanner test and the complete native `core` package test suite passed.

**Subsystem:** `enygma_dvp/src/core`.

**BN254 scalar-field modulus (R):**

`21888242871839275222246405745257275088548364400416034343698204186575808495617`

**Tested token ID (R-1):**

`21888242871839275222246405745257275088548364400416034343698204186575808495616`

### Preserved test artifact

| Layer | Research artifact | SHA-256 |
|---|---|---|
| Native ERC-20 scanner | `tests/erc20_scanner_token_id_boundary_test.go` | `dfed734a7fa35c1ea129703f91010796bd709ef71e04862d21dc754e827880c9` |

The preserved artifact was compared byte-for-byte with the test file executed in the official Enygma source repository.

### Hypothesis

The native ERC-20 scanner may accept a correctly encrypted note whose token ID is the largest canonical BN254 scalar-field representative, R-1, without applying a separate semantic token-ID restriction during note recovery.

### Test construction

The test creates a spend key pair and view key pair, then constructs a note using:

1. Token ID R-1 and amount 100.
2. ML-KEM encapsulation to the view key.
3. Payment salt and encryption-key derivation from the shared secret.
4. AES-GCM encryption of the token ID and amount.
5. An ERC-20 Poseidon commitment over the spend public key, derived salt field, amount, and token ID.
6. An `OnChainErc20Event` containing the commitment, ML-KEM capsule, and encrypted transaction data.
7. `ScanForErc20NotesWithRejected` using the corresponding view decapsulation key and spend public key.

### Observed result

The scanner accepted and recovered exactly one note.

- Rejected events: `0`.
- Recovered notes: `1`.
- Recovered token ID: exactly R-1.
- Recovered amount: `100`.
- Scanner returned no error.

The test therefore demonstrates acceptance of this canonical upper-bound token ID through the tested native note-recovery path.

### Execution evidence

Executed from `/workspaces/Enygma/enygma_dvp/src`:

- `go test ./core -run '^TestScanForErc20Notes_AcceptsTokenIDAtScalarFieldUpperBoundary$' -count=1 -v` — passed.
- `go test ./core -count=1` — complete core package passed.
- `gofmt -d enygma_dvp/src/core/token_id_field_boundary_test.go` — no formatting differences.
- `cmp` between the official test and preserved research artifact — byte-for-byte match.

### Limitations and interpretation

- This is an executed native scanner result, not merely source-level inference.
- The test constructs a mutually consistent note. It does not demonstrate acceptance of arbitrary malformed or attacker-substituted notes.
- R-1 is a positive canonical field representative; this experiment does not test literal negative integer serialization.
- The result does not establish that the ERC-20 vault treats token ID as an asset selector. The inspected vault uses a fixed ERC-20 asset contract.
- No unauthorized spend, forged proof, asset loss, or other on-chain exploit impact has been demonstrated.
- The intended policy for nonzero token IDs in the ERC-20 vault remains a separate question.

**Assessment:** The native ERC-20 scanner accepts and recovers a correctly encrypted note with token ID R-1. This is a reproducible boundary-acceptance observation. Its security significance depends on intended token-ID semantics and downstream application and contract validation.

## 11. EXP-MLK-005 — Payment-note plaintext-length validation

**Status:** Test artifact preserved; the test was included in the completed payment-note validation investigation. The exact invocation and raw terminal transcript are not recorded in this register.

**Subsystem:** `enygma_retail_payments/private_tags/src`.

**Test:** `TestDecryptPaymentNote_PlaintextLength`.

**Preserved test artifact:** `enygma_retail_payments/private_tags/src/payment_note_validation_test.go`

**SHA-256:** `7e403619192a3dc3b0f54d82e29c64abbc02cde6425a033c8ed993cafee251ad`

### Hypothesis

`DecryptPaymentNote` should enforce the expected payment-note plaintext size after decrypting an authenticated payload, rejecting plaintexts that do not match the required 96-byte structure.

### Test construction

The test derives a block key and encrypts plaintexts of three lengths using `EncryptPayload`. It then passes each ciphertext to `DecryptPaymentNote`:

- 95 bytes — expected rejection.
- 96 bytes — expected acceptance.
- 97 bytes — expected rejection.

For the two invalid lengths, the test checks for an error containing `unexpected plaintext length`. For the valid length, it checks that decryption succeeds and returns a non-nil note.

### Recorded test assertions

The preserved test asserts rejection of 95-byte and 97-byte plaintexts and acceptance of a 96-byte plaintext. The test source is retained in this repository for reviewer inspection.

### Limitations and interpretation

- This test concerns payment-note plaintext-length validation after authenticated decryption.
- It does not establish that malformed ciphertext can be emitted by an on-chain contract.
- It does not, by itself, demonstrate interruption of private-tag scanning, cursor non-advancement, retry behaviour, remote denial of service, or production exploitability.
- The exact execution command and raw output should not be inferred from the test source.

**Assessment:** The preserved test documents the expected 96-byte payment-note plaintext boundary. Its result should be interpreted within that validation path and not conflated with the separate private-tag scan-interruption experiment EXP-TAG-001.
