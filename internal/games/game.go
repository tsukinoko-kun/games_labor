package games

import "github.com/google/uuid"

type Game struct {
	ID string `json:"id"`
}

var Games = make(map[string]*Game)

func New() *Game {
	id := uuid.New().String()
	game := &Game{ID: id}
	Games[id] = game
	return game
}

func init() {
	game := &Game{ID: "28603f7e-77c7-487b-8d06-548354c35178"}
	Games[game.ID] = game
}
