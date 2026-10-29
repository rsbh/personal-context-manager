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

Scenario: A v* tag starts a release
  When I push the tag "v0.1.0"
  Then the "release" workflow runs: tests, then goreleaser
  And the GitHub release "v0.1.0" has 3 archives and checksums.txt
  And rsbh/homebrew-tap has Casks/pcm.rb with version 0.1.0
  And "brew install rsbh/tap/pcm" then "pcm --version" prints 0.1.0

Scenario: Missing token stops the release early
  Given HOMEBREW_TAP_TOKEN is empty
  When the workflow runs
  Then it fails before the build, with the message "HOMEBREW_TAP_TOKEN is not set"

Scenario: Branches and PRs do not release
  When I push a branch or open a PR
  Then the release workflow does not run

Scenario: Pre-release tag does not touch the tap
  When I push the tag "v0.1.0-rc.1"
  Then a GitHub pre-release is created
  And the tap is not changed
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

## Make a release

1. Merge to `main`. Wait for CI.
2. Try a pre-release first: `git tag v0.1.0-rc.1 && git push origin v0.1.0-rc.1`. Check the GitHub release page. The tap does not change.
3. Delete the pre-release and its tag: `gh release delete v0.1.0-rc.1 --cleanup-tag`.
4. Tag the release: `git tag v0.1.0 && git push origin v0.1.0`. The `release` workflow builds, publishes and updates the cask.
5. Check: `brew install rsbh/tap/pcm && pcm --version`.

The secret `HOMEBREW_TAP_TOKEN` is a fine-grained PAT (Contents write on `rsbh/homebrew-tap`). The workflow stops early if it is empty.
