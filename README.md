# Go / Board

A local two-player Go (Baduk / Weiqi) web application built with Go, HTML, CSS, and vanilla JavaScript. No third-party dependencies.

## Run

Install Go 1.22 or newer, then in this directory:

```sh
go run .
```

Open http://127.0.0.1:8080 in your browser. To use a different port, set the `ADDR` environment variable (e.g. `ADDR=127.0.0.1:8081 go run .` on macOS/Linux).

## Test and build

```sh
go test ./...
go build -o go-board .
```

Static assets are embedded in the compiled executable.

## Features

- 9×9, 13×13, and 19×19 boards; local two-player play.
- Server-side group/liberty detection, captures, suicide prevention, positional superko.
- Pass, undo, restart, capture counts, last-move indicator, keyboard arrow navigation.
- Two passes finish the game; Chinese-style area scoring with 7.5 komi.
- Responsive layout, no external assets or frameworks.

## Project layout

- `main.go` — Go HTTP server, game rules, JSON API, embedded static assets.
- `main_test.go` — tests for captures, suicide, ko, scoring, and HTTP endpoints.
- `static/index.html` — page structure.
- `static/style.css` — layout and design.
- `static/app.js` — browser interaction and undo history.

## Scope and limitations

The game is local, not online multiplayer or an AI opponent. State is held by the browser for the current page session; refresh resets the game. The API is stateless and trusts client-supplied history, so it is not suitable for competitive remote multiplayer. Endgame scoring assumes all remaining stones are alive; dead groups are not adjudicated automatically. Use undo and continue playing to remove dead groups before passing again.
