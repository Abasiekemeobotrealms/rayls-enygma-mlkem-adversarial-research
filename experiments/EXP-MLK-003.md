# EXP-MLK-003 — Correct-length Corrupted ML-KEM-768 Ciphertext

## 1. Objective

Determine how the Enygma native ERC20 scanner handles a correctly sized ML-KEM-768 ciphertext whose contents have been corrupted, and whether that event prevents recovery of a subsequent valid note.

## 2. Hypothesis

A correctly sized but corrupted ML-KEM-768 ciphertext may pass the scanner's capsule-length validation and produce a decapsulation result that does not successfully open the encrypted payload.

The expected behaviour is that the scanner silently skips the event as not addressed to the scanning key and continues processing subsequent events.

## 3. Source and artifact provenance

**Official source repository:** `raylsnetwork/Enygma`

**Recorded source revision:** `79c82dfb57595681b03d8da344702ef606f3a621`

**Branch:** `main`

**Production implementation:** `enygma_dvp/src/core/scan.go`

**Production implementation SHA-256:**

`a593f6650b1ac223bc1762f33c9cec8222cff9a70664af1dcc870674d0d47617`

**Git blob:** `12d65eb704b235190cfc6fb537205bd5754cf5dc`

**Research-authored test artifact:** `enygma_dvp/src/tests/scan_mlkem_adversarial_test.go`

**Test artifact SHA-256:**

`c089c37e8056429dc9653cc171aa8a14b8b79d2f9123f5cd514e2a31ad86e189`

The test was authored for this investigation and was not part of the recorded upstream source revision. It was executed against the tracked scanner implementation.

## 4. Test design

**Test function:** `TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote`

The test:

1. Constructs a valid ERC20 event with amount 6.
2. Confirms that its ciphertext is exactly 1,088 bytes.
3. Flips one bit in the first ciphertext byte using `corrupted.CipherText[0] ^= 0x01`.
4. Confirms that the modified ciphertext remains 1,088 bytes.
5. Appends a valid ERC20 event with amount 5.
6. Scans both events using `ScanForErc20NotesWithRejected`.

The test asserts that the scanner returns no fatal error, recovers exactly the later valid note with amount 5, and produces no explicit rejection entries.

## 5. Reproduction

Run from the official Enygma source checkout, in the `enygma_dvp/src` Go module:

    cd /workspaces/Enygma/enygma_dvp/src
    GOTOOLCHAIN=local go test ./tests -run '^TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote$' -v -count=1

**Recorded result:** PASS.

The recorded command and terminal output are preserved in `evidence/focused-test-execution-2026-11-15.md`.

See the [EXP-MLK-003 execution transcript](../evidence/focused-test-execution-2026-11-15.md#L38).

## 6. Observed result

The test passed.

The scanner returned no fatal error and recovered the later valid ERC20 note with amount 5.

The corrupted event produced no explicit rejection entry.

## 7. Source-level interpretation

The shared scanner in `core/scan.go` distinguishes between decapsulation errors and payload-opening failures.

At lines 53–57, a decapsulation error is recorded as `RejectBadCapsule`.

At lines 97–100, `openNote` maps a payload decryption failure to `errNotAddressed`.

At lines 61–67, `scanEvents` silently skips `errNotAddressed` without adding a rejection entry. Processing then continues with the next event.

This control flow is consistent with the observed result.

The experiment demonstrates the tested behaviour for a one-bit mutation. It does not establish that every correctly sized corrupted ciphertext follows the same path.

## 8. Limitations

- The test is research-authored, not an upstream test from the recorded source revision.
- The experiment tests one specific one-bit mutation of a valid 1,088-byte ciphertext.
- It does not establish behaviour for every possible corrupted ciphertext.
- It does not establish that an arbitrary external party can submit such an event through a production transaction path.
- It does not demonstrate remote denial of service, production exploitability, unauthorized spending, asset loss, or system-wide outage.
- The execution was against the recorded source working tree, not a production Rayls network.

**Assessment:** EXP-MLK-003 demonstrates that the tested correctly sized, one-bit-corrupted ciphertext did not produce a fatal scanner error or explicit rejection entry, while a later valid ERC20 note was recovered. Production reachability and security impact are not established.

<!-- EXP-MLK-003-consolidated-evidence -->
## 9. Consolidated execution and submission evidence

**Focused test:** `TestEXP_MLK_003_CorrectLengthCorruptedCiphertextDoesNotBlockLaterNote`

**Execution record:** The exact reproduction command and its recorded Go test output are preserved in the shared transcript:

[Focused test execution transcript](../evidence/focused-test-execution-2026-11-15.md#L38)

**Transcript SHA-256:**

`c7fc9ff0d277c8fda99570a2775129dd6ad416e35e7dfec7c063a2152512322b`

**Evidence date:** 2026-11-15.

**Recorded walkthrough and submission context:** The researcher confirms that the video recording showing this experiment's command and terminal output is included in the corresponding experiment section of the submitted X article. The article also references the GitHub research repository for reviewer access to the reproducibility materials. The recording is existing submission evidence; it is not being recreated or represented as a separate repository artifact.

**Evidence boundary:** The transcript documents the focused test execution and its observed output. The conclusions and limitations remain those stated in this experiment report; the test result alone does not establish production exploitability or an on-chain attack.
