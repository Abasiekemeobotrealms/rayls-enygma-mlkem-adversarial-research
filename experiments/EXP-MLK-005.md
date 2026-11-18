# EXP-MLK-005 — Payment-note Plaintext Length Boundary

## 1. Objective

Determine whether `DecryptPaymentNote` enforces the expected 96-byte plaintext structure after decrypting a payment-note payload.

## 2. Hypothesis

After successful authenticated decryption, `DecryptPaymentNote` is expected to accept a plaintext of exactly 96 bytes and reject plaintexts shorter or longer than that boundary.

## 3. Source and artifact provenance

**Official source repository:** `raylsnetwork/Enygma`

**Recorded source revision:** `79c82dfb57595681b03d8da344702ef606f3a621`

**Branch:** `main`

**Production implementation:** `enygma_retail_payments/private_tags/src/payment_notify.go`

**Production implementation SHA-256:**

`168ee34ae76ba05d14fa0f725cbef3e81ab3235b26f7d76bff827791fc8e3276`

**Research-authored test artifact:** `enygma_retail_payments/private_tags/src/payment_note_validation_test.go`

**Test artifact SHA-256:**

`7e403619192a3dc3b0f54d82e29c64abbc02cde6425a033c8ed993cafee251ad`

The test was authored for this investigation and was not part of the recorded upstream source revision. It was executed against the tracked production implementation.

## 4. Test design

**Test function:** `TestDecryptPaymentNote_PlaintextLength`

The test uses a fixed channel shared secret, recipient spend public key, and block number. For each case, it derives the block key, encrypts a zero-filled plaintext of the selected length using `EncryptPayload`, and passes the resulting ciphertext to `DecryptPaymentNote`.

The three cases are:

| Plaintext length | Expected result |
|---:|---|
| 95 bytes | Rejected |
| 96 bytes | Accepted |
| 97 bytes | Rejected |

For rejected cases, the test requires an error containing `unexpected plaintext length`. For the accepted case, it requires no error and a non-nil payment note.

## 5. Reproduction

Run from the official Enygma source checkout, in the `enygma_retail_payments/private_tags/src` Go module:

    cd /workspaces/Enygma/enygma_retail_payments/private_tags/src
    GOTOOLCHAIN=local go test . -run '^TestDecryptPaymentNote_PlaintextLength$' -v -count=1

**Recorded result:** PASS. All three subtests passed.

The recorded command and terminal output are preserved in `evidence/focused-test-execution-2026-11-15.md`.

See the [EXP-MLK-005 execution transcript](../evidence/focused-test-execution-2026-11-15.md#L88).

## 6. Observed result

All three test cases passed:

- 95-byte plaintext: rejected with an unexpected-plaintext-length error.
- 96-byte plaintext: accepted, returning a non-nil payment note.
- 97-byte plaintext: rejected with an unexpected-plaintext-length error.

## 7. Source-level interpretation

`DecryptPaymentNote` first derives the block key and calls `DecryptPayload`.

If decryption fails, the function returns a decryption error. After successful decryption, it checks that the plaintext length is exactly 96 bytes.

The function then parses the plaintext into three 32-byte values:

- Bytes 0–31: amount.
- Bytes 32–63: token ID.
- Bytes 64–95: salt.

The test therefore exercises the plaintext-length check after successful authenticated decryption. It does not test malformed ML-KEM capsule lengths.

## 8. Limitations

- The test is research-authored, not an upstream test from the recorded source revision.
- It tests only the three plaintext lengths 95, 96, and 97 bytes.
- It uses controlled plaintexts encrypted with the derived block key; it does not test arbitrary ciphertext corruption or decryption failures.
- It does not establish that an external party can cause these plaintext lengths to reach the production function through an on-chain transaction path.
- It does not demonstrate production exploitability, denial of service, unauthorized spending, asset loss, or system-wide outage.
- The execution was against the recorded source working tree, not a production Rayls network.

**Assessment:** EXP-MLK-005 demonstrates that `DecryptPaymentNote` accepts the tested 96-byte plaintext and rejects the tested 95-byte and 97-byte plaintexts after successful decryption. This confirms the local plaintext-structure boundary; production reachability and security impact are not established.

<!-- EXP-MLK-005-consolidated-evidence -->
## 9. Consolidated execution and submission evidence

**Focused test:** `TestDecryptPaymentNote_PlaintextLength`

**Execution record:** The exact reproduction command and its recorded Go test output are preserved in the shared transcript:

[Focused test execution transcript](../evidence/focused-test-execution-2026-11-15.md#L88)

**Transcript SHA-256:**

`c7fc9ff0d277c8fda99570a2775129dd6ad416e35e7dfec7c063a2152512322b`

**Evidence date:** 2026-11-15.

**Recorded walkthrough and submission context:** The researcher confirms that the video recording showing this experiment's command and terminal output is included in the corresponding experiment section of the submitted X article. The article also references the GitHub research repository for reviewer access to the reproducibility materials. The recording is existing submission evidence; it is not being recreated or represented as a separate repository artifact.

**Evidence boundary:** The transcript documents the focused test execution and its observed output. The conclusions and limitations remain those stated in this experiment report; the test result alone does not establish production exploitability or an on-chain attack.
