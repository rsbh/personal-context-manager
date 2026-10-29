# Release

Release with GoReleaser and Homebrew. A `v*` tag builds archives, a GitHub release and the formula in `rsbh/homebrew-tap`.

## Acceptance

```gherkin
Scenario: Snapshot release builds
  When I run "make release-check"
  Then goreleaser check passes
  And dist/ has archives for darwin arm64, linux amd64 and linux arm64
  And each archive has pcm, pcm-server, LICENSE and README.md
  And dist/checksums.txt lists every archive

Scenario: Binary has the release version
  When I run the darwin arm64 pcm from dist/ with --version
  Then the output has the snapshot version and the commit sha, not "dev"
```

## Files

| File | Purpose |
|---|---|
| `.goreleaser.yaml` | Builds `pcm` and `pcm-server` (no CGO, UI embedded) for darwin arm64, linux amd64 and linux arm64. Archives, checksums, changelog, and the Homebrew cask |
| `scripts/release-check.sh` | Checks the snapshot in `dist/` (the Acceptance above) |
| `make release-check` | `goreleaser check`, a snapshot build, then the script. Nothing is published |

## Version

`internal/version` gets `Version`, `Commit` and `Date` from ldflags. `pcm --version` and `pcm-server --version` print them.

## Homebrew

GoReleaser writes `Casks/pcm.rb` to `rsbh/homebrew-tap` (token: `HOMEBREW_TAP_TOKEN`, a fine-grained PAT with Contents write on the tap). A cask works on macOS and Linux brew. The macOS quarantine flag is removed in `postflight_steps`, because the binaries are not notarized. macOS Intel is not built.

Install: `brew install rsbh/tap/pcm`. Then `pcm server start` and `pcm skill install`.
