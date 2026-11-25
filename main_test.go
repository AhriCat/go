package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func position(stones map[int]int, turn int) Game {
	g := newGame(9)
	for p, v := range stones {
		g.Board[p] = v
	}
	g.Turn = turn
	g.History = []string{key(g.Board)}
	return g
}
func TestCapture(t *testing.T) {
	g := position(map[int]int{10: 2, 1: 1, 9: 1, 11: 1}, 1)
	next, e := g.move("play", 19)
	if e != nil || next.Board[10] != 0 || next.Captures[1] != 1 || next.Turn != 2 || g.Board[10] != 2 {
		t.Fatalf("capture: %+v %v", next, e)
	}
}
func TestSuicide(t *testing.T) {
	g := position(map[int]int{1: 2, 9: 2, 11: 2, 19: 2}, 1)
	if _, e := g.move("play", 10); e == nil {
		t.Fatal("accepted suicide")
	}
}
func TestKo(t *testing.T) {
	g := position(map[int]int{1: 1, 9: 1, 11: 1, 10: 2, 18: 2, 20: 2, 28: 2}, 1)
	next, e := g.move("play", 19)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = next.move("play", 10); e == nil || !strings.Contains(e.Error(), "Superko") {
		t.Fatalf("expected ko rejection, got %v", e)
	}
}
func TestPassScore(t *testing.T) {
	g := position(map[int]int{0: 1, 1: 1, 9: 1, 80: 2}, 1)
	g, e := g.move("pass", 0)
	if e != nil {
		t.Fatal(e)
	}
	g, e = g.move("pass", 0)
	if e != nil || g.Passes != 2 {
		t.Fatal("expected end")
	}
	s := g.score()
	if s.Black != 3 || s.White != 8.5 {
		t.Fatalf("score %+v", s)
	}
	if _, e = g.move("play", 3); e == nil {
		t.Fatal("accepted move after end")
	}
}
func TestTerritory(t *testing.T) {
	g := position(map[int]int{1: 1, 9: 1, 80: 2}, 1)
	s := g.score()
	if s.Black != 3 || s.White != 8.5 {
		t.Fatalf("score %+v", s)
	}
}
func TestAPI(t *testing.T) {
	h := handler()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Go / Board") {
		t.Fatal("index unavailable")
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/new?size=9", nil))
	var g Game
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &g) != nil || !g.valid() {
		t.Fatal("new game failed")
	}
	payload, _ := json.Marshal(Request{Game: g, Action: "play", Point: 0})
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/move", strings.NewReader(string(payload))))
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
}
