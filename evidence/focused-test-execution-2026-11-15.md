# Focused Test Execution Evidence — 2026-11-15

This file preserves raw terminal output from five focused test executions in the Enygma source working tree. The working tree included research-authored test artifacts and local test changes; see the milestone register for per-experiment provenance.

These transcripts document test execution. They are not independent proof of production exploitability. Interpret each result within its test fixture and subsystem.

---

## EXP-MLK-002 — Invalid ML-KEM ciphertext lengths

Official source: enygma_dvp/src/tests/scan_mlkem_adversarial_test.go

Test: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated

Working directory: /workspaces/Enygma/enygma_dvp/src

Command: GOTOOLCHAIN=local go test ./tests -run TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated -v -count=1

Raw terminal output:

=== RUN   TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated
=== RUN   TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_0
=== RUN   TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1
=== RUN   TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1087
=== RUN   TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1089
--- PASS: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated (0.00s)
    --- PASS: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_0 (0.00s)
    --- PASS: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1 (0.00s)
    --- PASS: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1087 (0.00s)
    --- PASS: TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated/length_1089 (0.00s)
PASS
ok      github.com/raylsnetwork/enygma_dvp/src/tests    0.028s

Interpretation: The tested native ERC20 scanner classified all four malformed ciphertext lengths as RejectBadCapsule, continued scanning, and recovered the later valid note. One rejection was recorded at index 0, with no fatal scanner error.

---

## EXP-MLK-003 — Correct-length corrupted ciphertext

Official source: enygma_dvp/src/tests/scan_mlkem_adversarial_test.go

Test: TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote

Working directory: /workspaces/Enygma/enygma_dvp/src

Command: GOTOOLCHAIN=local go test ./tests -run '^TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote$' -v -count=1

Raw terminal output:

=== RUN   TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote
--- PASS: TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote (0.00s)
PASS
ok      github.com/raylsnetwork/enygma_dvp/src/tests    0.028s

Interpretation: The test flips one bit in a valid 1,088-byte ciphertext. The scanner returns no fatal error and recovers the later valid note with amount 5. The corrupted event produces no explicit rejection entry in this test.

---

## EXP-MLK-004 — Correct-length corruption across scanner families

Official source: enygma_dvp/src/tests/scan_mlkem_adversarial_test.go

Test: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners

Working directory: /workspaces/Enygma/enygma_dvp/src

Command: GOTOOLCHAIN=local go test ./tests -run '^TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners$' -v -count=1

Raw terminal output:

=== RUN   TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners
=== RUN   TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC20
=== RUN   TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC721
=== RUN   TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC1155
=== RUN   TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ZK_DvP
--- PASS: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners (0.00s)
    --- PASS: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC20 (0.00s)
    --- PASS: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC721 (0.00s)
    --- PASS: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ERC1155 (0.00s)
    --- PASS: TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners/ZK_DvP (0.00s)
PASS
ok      github.com/raylsnetwork/enygma_dvp/src/tests    0.029s

Interpretation: All four scanner-family subtests returned without a fatal scan error and recovered the later valid event. The tested recovered values were ERC20 amount 5, ERC721 token ID 5, ERC1155 amount 5, and ZK DvP output amount 5. None recorded an explicit rejection entry for the corrupted event.

---

## EXP-MLK-005 — Payment-note plaintext length boundary

Official source: enygma_retail_payments/private_tags/src/payment_note_validation_test.go

Test: TestDecryptPaymentNote_PlaintextLength

Working directory: /workspaces/Enygma/enygma_retail_payments/private_tags/src

Command: GOTOOLCHAIN=local go test . -run '^TestDecryptPaymentNote_PlaintextLength$' -v -count=1

Raw terminal output:

=== RUN   TestDecryptPaymentNote_PlaintextLength
=== RUN   TestDecryptPaymentNote_PlaintextLength/95_bytes_rejected
=== RUN   TestDecryptPaymentNote_PlaintextLength/96_bytes_accepted
=== RUN   TestDecryptPaymentNote_PlaintextLength/97_bytes_rejected
--- PASS: TestDecryptPaymentNote_PlaintextLength (0.00s)
    --- PASS: TestDecryptPaymentNote_PlaintextLength/95_bytes_rejected (0.00s)
    --- PASS: TestDecryptPaymentNote_PlaintextLength/96_bytes_accepted (0.00s)
    --- PASS: TestDecryptPaymentNote_PlaintextLength/97_bytes_rejected (0.00s)
PASS
ok      enygma_retail_payments/private_tags/src 0.034s

Interpretation: The test derives a block key, encrypts test plaintexts, and validates the decrypted payment note. A 96-byte plaintext is accepted; 95-byte and 97-byte plaintexts are rejected with an unexpected-plaintext-length error. This tests plaintext validation after authenticated decryption, not malformed ML-KEM capsule length or on-chain reachability.

---

## EXP-TAG-001 — Malformed matching private-tag event

Official source: enygma_retail_payments/shield_demo/engine/tags_test.go

Test: TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent

Working directory: /workspaces/Enygma/enygma_retail_payments/shield_demo

Command: GOTOOLCHAIN=local go test ./engine -run '^TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent$' -v -count=1

Raw terminal output:

=== RUN   TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent
--- PASS: TestScanTags_MalformedMatchingEventAbortsBeforeLaterValidEvent (0.00s)
PASS
ok      github.com/raylsnetwork/enygma_retail_payments/shield_demo/engine       0.032s

Interpretation: In this test fixture, a malformed matching event causes scanTags to return an error before processing a later valid event. The scan cursor remains at block 9 when block 10 fails. Retrying queries the same block range again. This focused test does not establish that arbitrary external parties can create matching events, nor does it demonstrate a remotely exploitable denial of service, asset loss, or system-wide outage. Production uses a real Ethereum client; the test uses a tagLogReader seam.

---
