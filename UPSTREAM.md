# Upstream parity

The canonical published copy is [`mvanhorn/printing-press-library/library/commerce/shopper`](https://github.com/mvanhorn/printing-press-library/tree/main/library/commerce/shopper). This standalone repository is synchronized with Printing Press Library release `2026.7.1` at source commit `11d7b4ea752c9c8fb8909037133058eb57560cc0`.

Functional source is kept equivalent. These packaging differences are intentional:

- Go module and internal import paths use `github.com/educrvz/shopper-pp-cli` here.
- Standalone install, release, Homebrew, and MCPB links point to `educrvz/shopper-pp-cli`.
- The standalone semantic version remains independent from the library's calendar-based release version.
- Printing Press research manuscripts and library-wide release automation files are not mirrored.

When updating, compare against the published directory, apply functional changes here, preserve the differences above, run `gofmt -w` on changed Go files, and run `go test ./...` with the Go version declared in `go.mod`.
