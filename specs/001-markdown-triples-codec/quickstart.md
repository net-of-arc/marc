# Quickstart & Validation Guide: Markdown ⇄ Triples Codec

How to prove the feature works end to end once it is implemented. The API is in [contracts/api.md](contracts/api.md) and the byte rules are in [contracts/format.md](contracts/format.md).

## Prerequisites

- Go 1.26 or newer (`go version`)
- Repository checked out on branch `001-markdown-triples-codec`

## 1. Build and the full test suite

```bash
CGO_ENABLED=0 go build ./...
go test -race -cover ./...
```

**Expected**: the build succeeds with CGO disabled, every test passes under the race detector, and coverage is reported.

## 2. Acceptance scenarios (spec → test, 1:1)

```bash
go test -race -run 'Test(Decode|Encode|RoundTrip|Prose|Malformed|MergeState)' -v .
```

| Spec story | Test file (`package marc_test`) | Fixture |
|---|---|---|
| 1 Read | `decode_test.go` | `testdata/core-3-example.md` (CORE §3, canonical) |
| 2 Write | `encode_test.go` | same file as the expected output |
| 3 Round trip | `roundtrip_test.go` | `testdata/canonical/*.md`, `testdata/forms/*.md` (body/inline forms) |
| 4 Prose | `prose_test.go` | `testdata/prose-untyped-links.md` |
| 5 Malformed | `malformed_test.go` | `testdata/malformed/*.md` |
| 6 Merge state | `mergestate_test.go` | `testdata/core-8-2-conflicts.md`, `testdata/core-8-3-provenance.md` |

**Expected**: every test names its scenario (`// Scenario X.Y from specs/001-markdown-triples-codec/spec.md`) and passes (SC-007).

## 3. Testable examples (SC-003)

```bash
go test -run Example -v .
```

**Expected**: `ExampleDecode` prints the CORE §3 triple listing, with the subject, the type and six statements, using at most three library calls. `ExampleEncode` prints the canonical CORE §3 file. `ExampleSelect` prints only the reference targets.

## 4. Round-trip property and fuzzing (SC-001, SC-002, SC-004)

```bash
go test -run TestRoundTripGenerated -v .        # ≥ 1,000 generated nodes: triples → md → triples
go test -run 'Fuzz' -v .                         # replays the seed corpora in testdata/fuzz/ (fast)
```

**Expected**: no failures. `go test ./...` and CI replay only the seed corpora under `testdata/fuzz/`.

**Optional, on request only.** Long fuzzing keeps every CPU core busy, so it is never part of the regular workflow:

```bash
go test -run '^$' -fuzz '^FuzzDecode$' -fuzztime 10m .     # no crash, hang, or unbounded memory
go test -run '^$' -fuzz '^FuzzRoundTrip$' -fuzztime 10m .  # decode ok ⇒ encode ⇒ decode is equal
```

Add `-parallel 2` to limit the number of fuzzing workers.

## 5. Performance budget (SC-006)

```bash
go test -run '^$' -bench 'Benchmark(Decode|Encode)' -benchmem .
```

**Expected**: decoding the 10 KB / 100-statement fixture takes under 1 ms per operation. Record the allocations per operation in ARCHITECTURE.md as the baseline for future `benchstat` comparisons.

## 6. Static checks (mirror of CI `check-code`)

```bash
test -z "$(gofmt -l .)"
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@<pinned> ./...
go mod tidy && git diff --exit-code go.mod go.sum
```

**Expected**: no output and a zero exit status from each command.

## 7. Manual smoke test

Decode a real node and list its triples from a scratch program outside the module:

```go
n, err := marc.Unmarshal(src)
for t := range n.Triples() { fmt.Println(t.Predicate, t.Object) }
```

**Expected**: one line per statement, with no lines for the `@id`, `@type`, `@conflicts` or `@provenance` keywords.
