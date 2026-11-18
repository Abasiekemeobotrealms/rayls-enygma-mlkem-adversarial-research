# EXP-MLK-004 — Correct-length Corruption Across Scanner Families

## 1. Objective

Determine whether a correctly sized, one-bit-corrupted ML-KEM-768 ciphertext prevents subsequent valid-note recovery across the Enygma ERC20, ERC721, ERC1155, and ZK-DvP scanner families.

## 2. Hypothesis

For the tested one-bit mutation, each scanner is expected to continue processing after the corrupted event, recover the subsequent valid event, and return without a fatal scan error or explicit rejection entry for the corrupted event.

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

**Test function:** `TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners`

**Source lines:** 83–157 of `scan_mlkem_adversarial_test.go`.

The test defines a helper that verifies a ciphertext is exactly 1,088 bytes, copies it, and flips one bit in its first byte.

It then executes four independent subtests:

| Scanner | Corrupted event | Later valid event | Recovery assertion |
|---|---|---|---|
| ERC20 | Amount 6 | Amount 5 | One note with amount 5 |
| ERC721 | Token ID 6 | Token ID 5 | One note with token ID 5 |
| ERC1155 | Amount 6 | Amount 5 | One note with amount 5 |
| ZK-DvP | Output amount 6 | Output amount 5 | One swap with output amount 5 |

Each subtest asserts that the scan returns no fatal error and no explicit rejection entries.

## 5. Reproduction

Run from the official Enygma source checkout, in the `enygma_dvp/src` Go module:

    cd /workspaces/Enygma/enygma_dvp/src
    GOTOOLCHAIN=local go test ./tests -run '^TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners$' -v -count=1

**Recorded result:** PASS. All four subtests passed.

The recorded command and terminal output are preserved in `evidence/focused-test-execution-2026-11-15.md`.

See the [EXP-MLK-004 execution transcript](../evidence/focused-test-execution-2026-11-15.md#L59).

## 6. Observed result

All four scanner-family subtests passed.

For each scanner, the tested corrupted event did not cause a fatal scan error. The later valid event was recovered with the expected value:

- ERC20: amount 5.
- ERC721: token ID 5.
- ERC1155: amount 5.
- ZK-DvP: output amount 5.

None of the four subtests recorded an explicit rejection entry for the corrupted event.

## 7. Interpretation

The experiment demonstrates consistent observed behaviour across the four tested scanner families for the specific one-bit mutation applied to a valid 1,088-byte ciphertext.

The result is consistent with the shared scanner control flow documented in EXP-MLK-003: a payload-opening failure classified as `errNotAddressed` is skipped without an explicit rejection entry, allowing subsequent events to be processed.

This experiment does not establish that all correctly sized corrupted ciphertexts behave identically across every scanner family.

## 8. Limitations

- The test is research-authored, not an upstream test from the recorded source revision.
- The experiment tests one specific one-bit mutation of a valid 1,088-byte ciphertext.
- It does not establish behaviour for every possible corrupted ciphertext or every scanner input.
- The test uses controlled fixtures; it does not establish that an arbitrary external party can submit such events through production transaction paths.
- It does not demonstrate remote denial of service, production exploitability, unauthorized spending, asset loss, or system-wide outage.
- The execution was against the recorded source working tree, not a production Rayls network.

**Assessment:** EXP-MLK-004 demonstrates that, for the tested one-bit ciphertext mutation, all four scanner families recovered their subsequent valid event without a fatal scan error or explicit rejection entry. Production reachability and security impact are not established.

<!-- EXP-MLK-004-consolidated-evidence -->
## 9. Consolidated execution and submission evidence

**Focused test:** `TestEXP_MLK_004_CorrectLengthCorruptionIsIsolatedAcrossScanners`

**Execution record:** The exact reproduction command and its recorded Go test output are preserved in the shared transcript:

[Focused test execution transcript](../evidence/focused-test-execution-2026-11-15.md#L59)

**Transcript SHA-256:**

`c7fc9ff0d277c8fda99570a2775129dd6ad416e35e7dfec7c063a2152512322b`

**Evidence date:** 2026-11-15.

**Recorded walkthrough and submission context:** The researcher confirms that the video recording showing this experiment's command and terminal output is included in the corresponding experiment section of the submitted X article. The article also references the GitHub research repository for reviewer access to the reproducibility materials. The recording is existing submission evidence; it is not being recreated or represented as a separate repository artifact.

**Evidence boundary:** The transcript documents the focused test execution and its observed output. The conclusions and limitations remain those stated in this experiment report; the test result alone does not establish production exploitability or an on-chain attack.
