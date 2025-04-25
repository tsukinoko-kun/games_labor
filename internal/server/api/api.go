package api

import (
	"fmt"
	"gameslabor/internal/games"
	"log"
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

func gameState(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	fmt.Printf("gameState called with id: %s\n", id)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("websocket upgrade error: %v\n", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ws read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(websocket.TextMessage, message) // echo back message
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
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
