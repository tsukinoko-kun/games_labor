package games

import (
	"context"
	"fmt"
	"gameslabor/internal/ai"
	"gameslabor/internal/games/scenarios"
	"gameslabor/internal/server/hub"
	"log"
	"sync"

	"github.com/google/uuid"
)

type (
	Game struct {
		ID      string             `json:"id"`
		AI      *ai.AI             `json:"ai"`
		Players map[string]*Player `json:"players"`
		mut     sync.Mutex         `json:"-"`
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
		mut:     sync.Mutex{},
	}
	Games[id] = game
	return game
}

func (g *Game) AddPlayer(playerID string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if _, ok := g.Players[playerID]; ok {
		return
	}
	g.Players[playerID] = &Player{ID: playerID}
}

func (g *Game) SetPlayerDescription(playerID string, description string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if player, ok := g.Players[playerID]; ok {
		player.Description = description
		hub.Broadcast(g.ID, g)
	}
}

func (g *Game) PlayerInput(playerID string, input string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateRunning {
		return
	}

	resp := g.AI.Continue(fmt.Sprintf(`Spieler %s sagt: %s`, playerID, input))
	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "user", PlayerID: playerID, Message: input})
	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: resp.NarratorText})
	hub.Broadcast(g.ID, g)
}

func (g *Game) Start(playerID string, scenario string, violenceLevel uint8, duration uint8) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateInit {
		return
	}

	ctx := context.Background()
	s, err := scenarios.FromID(scenario)
	if err != nil {
		log.Printf("failed to get scenario description: %v", err)
		return
	}
	g.State = GameStateRunning
	g.AI, err = ai.New(ctx)
	if err != nil {
		log.Printf("failed to create AI: %v", err)
		return
	}

	for _, player := range g.Players {
		g.AI.CharacterData[player.ID] = []string{player.Description}
	}
	hub.Broadcast(g.ID, g)

	resp := g.AI.Start(s)
	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: resp.NarratorText})
	fmt.Printf("First message: %s\n", resp.JSON())
	hub.Broadcast(g.ID, g)
}

func init() {
	// just for testing
	newWithId("28603f7e-77c7-487b-8d06-548354c35178")
}

const (
	GameStateInit GameState = iota
	GameStateRunning
)
