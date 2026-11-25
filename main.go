package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static/*
var assets embed.FS

type Game struct {
	Size     int      `json:"size"`
	Board    []int    `json:"board"`
	Turn     int      `json:"turn"`
	Captures [3]int   `json:"captures"`
	Passes   int      `json:"passes"`
	History  []string `json:"history"`
	Last     int      `json:"last"`
	Komi     float64  `json:"komi"`
}
type Score struct {
	Black float64 `json:"black"`
	White float64 `json:"white"`
}
type Request struct {
	Game   Game   `json:"game"`
	Action string `json:"action"`
	Point  int    `json:"point"`
}

func validSize(n int) bool { return n == 9 || n == 13 || n == 19 }
func newGame(n int) Game {
	g := Game{Size: n, Board: make([]int, n*n), Turn: 1, Last: -1, Komi: 7.5}
	g.History = []string{key(g.Board)}
	return g
}
func key(b []int) string {
	var s strings.Builder
	for _, v := range b {
		s.WriteByte(byte('0' + v))
	}
	return s.String()
}
func neighbors(p, n int) []int {
	a := make([]int, 0, 4)
	if p%n > 0 {
		a = append(a, p-1)
	}
	if p%n < n-1 {
		a = append(a, p+1)
	}
	if p >= n {
		a = append(a, p-n)
	}
	if p < n*(n-1) {
		a = append(a, p+n)
	}
	return a
}
func group(b []int, p, n int) ([]int, bool) {
	stones := []int{p}
	seen := map[int]bool{p: true}
	liberty := false
	for i := 0; i < len(stones); i++ {
		for _, q := range neighbors(stones[i], n) {
			if b[q] == 0 {
				liberty = true
			}
			if b[q] == b[p] && !seen[q] {
				seen[q] = true
				stones = append(stones, q)
			}
		}
	}
	return stones, liberty
}
func (g Game) valid() bool {
	if !validSize(g.Size) || len(g.Board) != g.Size*g.Size || g.Turn < 1 || g.Turn > 2 || g.Passes < 0 || g.Passes > 2 || g.Komi != 7.5 || len(g.History) == 0 {
		return false
	}
	for _, v := range g.Board {
		if v < 0 || v > 2 {
			return false
		}
	}
	return g.History[len(g.History)-1] == key(g.Board)
}
func (g Game) move(action string, p int) (Game, error) {
	if !g.valid() {
		return g, errors.New("Invalid game state")
	}
	if g.Passes >= 2 {
		return g, errors.New("Game has ended")
	}
	if action == "pass" {
		g.Passes++
		g.Turn = 3 - g.Turn
		g.Last = -1
		return g, nil
	}
	if action != "play" || p < 0 || p >= len(g.Board) || g.Board[p] != 0 {
		return g, errors.New("Choose an empty intersection")
	}
	b := append([]int(nil), g.Board...)
	b[p] = g.Turn
	captured := 0
	for _, q := range neighbors(p, g.Size) {
		if b[q] != 3-g.Turn {
			continue
		}
		stones, liberty := group(b, q, g.Size)
		if !liberty {
			for _, s := range stones {
				b[s] = 0
				captured++
			}
		}
	}
	if _, liberty := group(b, p, g.Size); !liberty {
		return g, errors.New("Suicide is not allowed")
	}
	k := key(b)
	for _, h := range g.History {
		if h == k {
			return g, errors.New("Superko: this board position has occurred before")
		}
	}
	g.Board = b
	g.History = append(append([]string(nil), g.History...), k)
	g.Captures[g.Turn] += captured
	g.Turn = 3 - g.Turn
	g.Passes = 0
	g.Last = p
	return g, nil
}
func (g Game) score() Score {
	s := Score{White: g.Komi}
	seen := map[int]bool{}
	for p, v := range g.Board {
		if v == 1 {
			s.Black++
			continue
		}
		if v == 2 {
			s.White++
			continue
		}
		if seen[p] {
			continue
		}
		region := []int{p}
		seen[p] = true
		borders := [3]bool{}
		for i := 0; i < len(region); i++ {
			for _, q := range neighbors(region[i], g.Size) {
				if g.Board[q] == 0 {
					if !seen[q] {
						seen[q] = true
						region = append(region, q)
					}
				} else {
					borders[g.Board[q]] = true
				}
			}
		}
		if borders[1] && !borders[2] {
			s.Black += float64(len(region))
		}
		if borders[2] && !borders[1] {
			s.White += float64(len(region))
		}
	}
	return s
}
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func handler() http.Handler {
	mux := http.NewServeMux()
	files, _ := fs.Sub(assets, "static")
	mux.Handle("GET /", http.FileServer(http.FS(files)))
	mux.HandleFunc("GET /api/new", func(w http.ResponseWriter, r *http.Request) {
		n := 19
		if x := r.URL.Query().Get("size"); x != "" {
			if _, e := fmt.Sscan(x, &n); e != nil || !validSize(n) {
				reply(w, 400, map[string]string{"error": "Use size 9, 13, or 19"})
				return
			}
		}
		reply(w, 200, newGame(n))
	})
	mux.HandleFunc("POST /api/move", func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); e != nil {
			reply(w, 400, map[string]string{"error": "Invalid request"})
			return
		}
		g, e := req.Game.move(req.Action, req.Point)
		if e != nil {
			reply(w, 400, map[string]string{"error": e.Error()})
			return
		}
		reply(w, 200, struct {
			Game  Game  `json:"game"`
			Score Score `json:"score"`
		}{g, g.score()})
	})
	return mux
}
func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Printf("Go board: http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, handler()))
}
