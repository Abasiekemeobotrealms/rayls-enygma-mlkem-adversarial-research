# EXP-MLK-006 — On-Chain Runtime Reachability of Malformed ML-KEM Capsules in Enygma Payment Deposits

## 1. Experiment status

**Status:** PASS — executed against an initialized local Hardhat Payment deployment.

**Subsystems:**
- Enygma Payment `Erc20CoinVault.depositV2`
- `EncryptedNote` event emission and Ethereum JSON-RPC `FilterLogs`
- Production Enygma native ERC-20 note scanner

**Official source repository:** `https://github.com/raylsnetwork/Enygma`

**Recorded source revision:** `79c82dfb57595681b03d8da344702ef606f3a621`

**Official test path:** `enygma_dvp/test/29_v2_malformed_capsule_deposit_test.go`

**Preserved research artifact:** `tests/29_v2_malformed_capsule_deposit_test.go`

**Test function:** `TestEXP_MLK_006_MalformedCapsuleDepositRuntimeReachability`

**SHA-256:** `388223157b4aee12883f09071af249a428c668cdc896812a958e30e11daeb73d`

The preserved research artifact was copied from the official source checkout and verified byte-for-byte with `cmp`.

## 2. Hypothesis and research question

**Hypothesis:** `Erc20CoinVault.depositV2` may accept `ciphertextI` values whose byte lengths are invalid for ML-KEM-768, emit those exact bytes in an `EncryptedNote` event, and make those events retrievable through the RPC log-query mechanism.

The experiment also tests whether the production native ERC-20 scanner isolates those malformed events as bad capsules and continues to recover a later valid note in the same scan.

This is an end-to-end local runtime reachability test. It goes beyond source inspection and isolated scanner fixtures.

## 3. Test construction

The test requires an already-running and initialized local Payment deployment. It does not deploy or initialize contracts itself.

It creates spend and view key pairs, derives a valid ML-KEM encapsulation and shared-secret-based payment data for each case, and submits five `depositV2` transactions.

The tested capsule lengths are:

| Case | Submitted capsule length | Purpose |
|---|---:|---|
| Empty | 0 bytes | Empty malformed capsule |
| One-byte | 1 byte | Minimal non-empty malformed capsule |
| Short | 1087 bytes | One byte below ML-KEM-768 size |
| Oversized | 1089 bytes | One byte above ML-KEM-768 size |
| Valid control | 1088 bytes | Correct-length control |

For malformed cases, only the submitted capsule bytes are replaced with a byte slice of the selected length. The commitment and encrypted payload are derived from the valid encapsulation created for that case. The valid control retains its original 1088-byte capsule.

Each transaction is submitted to the local `Erc20CoinVault.depositV2` contract and awaited until mined. The test requires successful transaction receipts.

After submission, the test queries the vault address for `EncryptedNote` events using `FilterLogs`, constrained by the event signature, indexed vault ID `0`, and the experiment block range.

The test checks that all five events were retrieved and that their commitments, capsule bytes, payload bytes, and transaction hashes match the corresponding submissions.

The retrieved event data is converted to `core.OnChainErc20Event` values and passed to `core.ScanForErc20NotesWithRejected`.

The test expects four `RejectBadCapsule` rejections at event indexes 0 through 3, no fatal scanner error, and recovery of the later valid control note with amount `1` and token ID `0`.

## 4. Reproduction

Prerequisites: official Enygma checkout at the recorded revision, Go dependencies, and initialized local Hardhat Payment deployment at `127.0.0.1:8545` (chain ID 1337).

Working directory: `enygma_dvp/test`.

Command: `GOTOOLCHAIN=local GOPROXY=off go test -mod=readonly -v -count=1 -run '^TestEXP_MLK_006_MalformedCapsuleDepositRuntimeReachability$' .`

Result: PASS in approximately 3.03 seconds.

## 5. Runtime evidence

All five deposit transactions were mined, and all five `EncryptedNote` events were retrieved through RPC `FilterLogs`:

| Case | Capsule bytes | Block | Transaction hash |
|---|---:|---:|---|
| Empty | 0 | 27 | `0x36a9e482ca5d588aad41d328d53d7d22f9e21e4fa3d7863f64342f0a19f96f7e` |
| One-byte | 1 | 28 | `0xd07cd9927cce30c93677fa02ffa65a4a95bdfcf1ecbb2581b76ddae690ade839` |
| Short | 1087 | 29 | `0x443c2ee6d1dc7d54f5dd51c41daeccc29a614e7d118c0105f50260f74f558084` |
| Oversized | 1089 | 30 | `0x7595133b60e3ca0ed8b0997827f62c76dadf00f8323aff4b1d10781c8cf586d4` |
| Valid control | 1088 | 31 | `0x8b0902830bb53c7004e41e33e94da762f7db12cdca7356592a33683f0481e866` |

The scanner rejected all four malformed capsules as `RejectBadCapsule` and recovered the later valid control note (amount 1, token ID 0) without a fatal error.

## 6. Limitations

This demonstrates runtime reachability on the initialized local Hardhat deployment across `depositV2`, event emission, RPC retrieval, and the production native ERC-20 scanner. It does not establish production Rayls network behaviour, exploitation, theft, asset loss, double-spending, or denial of service. The malformed capsule bytes were substituted while the commitment and encrypted payload were derived from a valid encapsulation; this isolates capsule-length handling rather than demonstrating a complete attacker-controlled payment flow. Transaction hashes and block numbers are local-chain evidence only.
