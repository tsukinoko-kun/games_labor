package games

import (
	"gameslabor/internal/server/hub"

	"github.com/google/uuid"
)

type (
	Game struct {
		ID      string             `json:"id"`
		Players map[string]*Player `json:"players"`
		State   GameState          `json:"state"`
	}

	GameState uint8

	Player struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	}
)

var Games = make(map[string]*Game)

func New() *Game {
	return newWithId(uuid.NewString())
}

func newWithId(id string) *Game {
	game := &Game{
		ID:      id,
		State:   GameStateInit,
		Players: make(map[string]*Player),
	}
	Games[id] = game
	return game
}

func (g *Game) AddPlayer(playerID string) {
	if _, ok := g.Players[playerID]; ok {
		return
	}
	g.Players[playerID] = &Player{ID: playerID}
}

func (g *Game) SetPlayerDescription(playerID string, description string) {
	if player, ok := g.Players[playerID]; ok {
		player.Description = description
		hub.Broadcast(g.ID, g)
	}
}

func (g *Game) Start(playerID string, scenario string, violenceLevel uint8, duration uint8) {

}

func init() {
	// just for testing
	newWithId("28603f7e-77c7-487b-8d06-548354c35178")
}

const (
	GameStateInit GameState = iota
	GameStateRunning
)
