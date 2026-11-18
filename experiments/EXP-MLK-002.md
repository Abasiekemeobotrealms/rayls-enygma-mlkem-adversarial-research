# EXP-MLK-002 — Invalid ML-KEM-768 Ciphertext Lengths

## 1. Objective

Determine how the Enygma native scanner handles ML-KEM-768 capsules whose byte lengths differ from the required ciphertext size, and whether a malformed event prevents a later valid event from being recovered.

This experiment concerns malformed ciphertext lengths. Correct-length corrupted ciphertexts are examined separately in EXP-MLK-003 and EXP-MLK-004.

## 2. Hypothesis

ML-KEM-768 ciphertexts with invalid lengths cause decapsulation to return an error. The shared scanner should classify those events as `RejectBadCapsule`, continue scanning, and recover a subsequent valid event.

## 3. Source and artifact provenance

**Official source repository:** `raylsnetwork/Enygma`

**Recorded source revision:** `79c82dfb57595681b03d8da344702ef606f3a621`

**Branch:** `main`

**Production implementation inspected:** `enygma_dvp/src/core/scan.go`

The scanner implementation was tracked and unchanged at the recorded source revision.

**Production implementation SHA-256:**

`a593f6650b1ac223bc1762f33c9cec8222cff9a70664af1dcc870674d0d47617`

**Git blob:** `12d65eb704b235190cfc6fb537205bd5754cf5dc`

**Research-authored test artifact:** `tests/scan_mlkem_adversarial_test.go`

**Test artifact SHA-256:**

`c089c37e8056429dc9653cc171aa8a14b8b79d2f9123f5cd514e2a31ad86e189`

The test artifact was created for this investigation and was not part of the recorded upstream source revision. The experiment therefore represents a research-authored test executed against the tracked scanner implementation, not an upstream test already present in that revision.

## 4. Test design

**Test function:** `TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated`

The test constructs malformed capsule cases with the following lengths:

| Case | Capsule length |
|---|---:|
| Empty | 0 bytes |
| One byte | 1 byte |
| One byte below ML-KEM-768 size | 1087 bytes |
| One byte above ML-KEM-768 size | 1089 bytes |

The expected ML-KEM-768 ciphertext size is 1088 bytes.

A valid ERC20 event follows the malformed event. The test checks that the scanner does not return a fatal error, classifies the malformed capsule as `RejectBadCapsule`, recovers the later valid note with amount 5, and reports exactly one rejection at event index 0.

## 5. Reproduction

Run from the official Enygma source checkout, in the `enygma_dvp/src` Go module:

    cd /workspaces/Enygma/enygma_dvp/src
    GOTOOLCHAIN=local go test ./tests -run TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated -v -count=1

**Recorded result:** PASS. All four malformed-length subtests passed.

The recorded command and terminal output are preserved in `evidence/focused-test-execution-2026-11-15.md`.

See the [EXP-MLK-002 execution transcript](../evidence/focused-test-execution-2026-11-15.md#L9).

## 6. Observed result

The tested malformed capsule lengths were classified as `RejectBadCapsule`.

The scanner continued through the event sequence and recovered the later valid ERC20 note with amount 5. The test observed exactly one rejection at index 0 and no fatal scanner error.

This is consistent with the inspected scanner control flow: a decapsulation error is recorded as a bad-capsule rejection and does not, by itself, terminate the scan.

## 7. Interpretation

The experiment demonstrates that the tested invalid-length ML-KEM-768 capsules are isolated by the native scanner rejection path and do not prevent recovery of the later valid event in this fixture.

Correctly sized but corrupted capsules can follow a different decapsulation path and are covered by EXP-MLK-003 and EXP-MLK-004.

## 8. Limitations

- The test is a research-authored fixture, not an upstream test from the recorded source revision.
- The result applies to the four tested lengths and the tested ERC20 event sequence.
- It does not establish behaviour for every malformed input or every scanner execution context.
- It does not establish that an arbitrary external party can submit such an event through a production transaction path.
- It does not demonstrate remote denial of service, production exploitability, unauthorized spending, asset loss, or system-wide outage.
- The execution was against the recorded source working tree, not a production Rayls network.

**Assessment:** EXP-MLK-002 provides reproducible evidence that the tested invalid-length capsules are rejected as bad capsules while a later valid ERC20 event is recovered. Production reachability and security impact are not established.

<!-- EXP-MLK-002-consolidated-evidence -->
## 9. Consolidated execution and submission evidence

**Focused test:** `TestEXP_MLK_002_InvalidCiphertextLengthsAreIsolated`

**Execution record:** The exact reproduction command and its recorded Go test output are preserved in the shared transcript:

[Focused test execution transcript](../evidence/focused-test-execution-2026-11-15.md#L9)

**Transcript SHA-256:**

`c7fc9ff0d277c8fda99570a2775129dd6ad416e35e7dfec7c063a2152512322b`

**Evidence date:** 2026-11-15.

**Recorded walkthrough and submission context:** The researcher confirms that the video recording showing this experiment's command and terminal output is included in the corresponding experiment section of the submitted X article. The article also references the GitHub research repository for reviewer access to the reproducibility materials. The recording is existing submission evidence; it is not being recreated or represented as a separate repository artifact.

**Evidence boundary:** The transcript documents the focused test execution and its observed output. The conclusions and limitations remain those stated in this experiment report; the test result alone does not establish production exploitability or an on-chain attack.
