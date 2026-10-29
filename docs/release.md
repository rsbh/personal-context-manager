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
