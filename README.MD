# go-snip

Simple screenshot tool written in Go.

# Features
- Capture screenshot of the entire screen or a selected area
- Save screenshots to a configurable output directory
- Screenshot hotkeys

# Architecture

```
go-snip/
├── cmd/
│   └── main.go           # Entry point, handles hotkeys and orchestrates the app
├── internal/
│   ├── capture/
│   │   └── capture.go    # Wraps screenshot logic (capture screen, crop image)
│   ├── overlay/
│   │   └── selection.go  # Fyne window logic (fullscreen, mouse drag, visual rect)
│   └── utils/
│       └── file_save.go  # Helper to save images to disk
├── screenshots/          # Output folder (auto-created)
├── go.mod
└── go.sum
```

# Tech stack
- Go
- Fyne
- Screenshot library: github.com/kbinani/screenshot
- Hotkey library: golang.design/x/hotkey
- Formatting, tests: built-in tools (go fmt, go test)

# Instructions

- Can be run with `go run cmd/main.go`
- Tests should be light, but should exist at least once for each file
- Run `go fmt` to format the code
- Run `go vet` to run the linter
- Run `go test` to run the tests
- Run `go build` to build the binary
- Run `make sanity-fyne` to run all of the above commands
- After making changes, be sure to run `make sanity-fyne` to ensure the code is clean and working.