# Rayls Enygma — ML-KEM-768 Adversarial Research

Independent, evidence-led research into ML-KEM-768 ciphertext handling, payment-note validation, scanner failure isolation, and selected BN254 field-boundary behaviours in Rayls Enygma.

## Research scope

The investigation examines:

- Invalid ML-KEM-768 ciphertext lengths.
- Correct-length but corrupted ciphertexts.
- Scanner recovery after malformed or corrupted events.
- Behaviour across ERC20, ERC721, ERC1155, and ZK DvP scanners.
- Payment-note plaintext-length validation.
- Private-tag event scanning and cursor/retry behaviour.
- Selected BN254 field-representation and token-ID boundary cases.
- Relevant Solidity and existing integration-test paths, with production reachability limitations recorded.

This repository preserves research artifacts and findings. It is separate from the official Enygma source repository and does not modify that repository.

## Repository structure

- `docs/milestone-register.md` — research scope, source revision, experiment records, results, provenance, and limitations.
- `tests/scan_mlkem_adversarial_test.go` — malformed-length and correct-length corruption scanner experiments.
- `tests/commitment_field_boundary_test.go` — native commitment field-boundary test.
- `tests/payment_circuit_field_boundary_test.go` — Gnark payment-circuit field-boundary tests.
- `tests/field_boundary_handler_test.go` — payment HTTP-handler field-boundary tests.
- `tests/erc20_scanner_token_id_boundary_test.go` — native ERC20 scanner token-ID upper-boundary test.
- `tests/29_v2_malformed_capsule_deposit_test.go` — local Payment runtime test for malformed ML-KEM capsule deposits and scanner rejection.
- `experiments/EXP-MLK-006.md` — experiment design, setup, and evidence record for the local runtime test.
- `enygma_retail_payments/private_tags/src/payment_note_validation_test.go` — payment-note plaintext-length validation test.
- `enygma_retail_payments/shield_demo/engine/tags_test.go` — malformed matching private-tag event scanning test.

## Source provenance

The milestone register records the official Enygma source repository, branch, environment, and tested revision. The recorded source revision is:

`79c82dfb57595681b03d8da344702ef606f3a621`

The research artifacts are preserved copies of tests associated with that source investigation. Refer to the milestone register for the provenance and byte-comparison records of individual artifacts.

## Reproduction

The test files are preserved as evidence artifacts, not as a self-contained Go module. To reproduce an experiment, use the corresponding official Enygma source checkout at the recorded revision and place or compare the relevant test file in its original package.

The milestone register and individual experiment records preserve exact reproduction commands where available. Where an exact command or raw terminal transcript was not preserved, that limitation is stated rather than reconstructed.

Do not assume that every test can be run from this repository root. The original source package, dependencies, configuration, and—in integration cases—local services or deployment artifacts may be required.

## Evidence interpretation

The register distinguishes executed experiments, source inspection, and existing-test-source analysis. A passing test demonstrates the behaviour of its tested code path and fixture; it does not automatically establish production reachability or exploitability. EXP-MLK-006 additionally demonstrates malformed-capsule deposit and scanner behaviour on an initialized local Hardhat Payment deployment; production Rayls network behaviour remains unverified.

This research does not establish:

- System-wide post-quantum security.
- Remote denial of service.
- Unauthorized spending, theft, or asset loss.
- Double-spending or forged proofs.
- Exploitability of every inspected source-level observation.
- Formal FIPS or CMVP validation.

Findings and limitations should be read together. The milestone register is the authoritative detailed record.

## Repository status

This is a research evidence repository. It is not an official Rayls or Enygma code release, security advisory, or formal audit certification.
