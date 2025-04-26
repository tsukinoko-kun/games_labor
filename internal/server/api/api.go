package api

import (
	"gameslabor/internal/games"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type apiFunc = func(http.ResponseWriter, *http.Request)

func newGame(w http.ResponseWriter, r *http.Request) {
	game := games.New()
	id := game.ID
	http.Redirect(w, r, "/game?id="+id, http.StatusSeeOther)
}

func init() {
	apiRegister["/new_game"] = newGame
	apiRegister["/game_state"] = gameState
}

var apiRegister = map[string]apiFunc{}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	if f, ok := apiRegister[path]; ok {
		f(w, r)
		return
	}
	if f, ok := apiRegister[path+"/"]; ok {
		f(w, r)
		return
	}
}
