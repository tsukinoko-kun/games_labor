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

	PlayerData struct {
		Name       string `json:"name"`
		Age        string `json:"age"`
		Origin     string `json:"origin"`
		Appearance string `json:"appearance"`
	}

	Player struct {
		ID          string     `json:"id"`
		Description PlayerData `json:"description"`
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

func runningWithId(id string) *Game {
	p1 := uuid.NewString()
	game := &Game{
		ID:    id,
		State: GameStateRunning,
		Players: map[string]*Player{
			p1: {
				ID: p1,
				Description: PlayerData{
					Name:       "Geralt von Riva",
					Age:        "69 Jahre",
					Origin:     "Aufgewachsen im Hexer Bergfried Kaer Morhen. Nicht aus Riva.",
					Appearance: "Lange aschblonde Haare, Katzenaugen, bleiche Haut.",
				},
			},
		},
		mut: sync.Mutex{},
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

func (g *Game) SetPlayerDescription(p Player) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if player, ok := g.Players[p.ID]; ok {
		*player = p
		hub.Broadcast(g.ID, g)
	}
}

func (g *Game) PlayerInput(playerID string, input string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateRunning {
		return
	}

	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "user", PlayerID: playerID, Message: input})
	hub.Broadcast(g.ID, g)

	resp := g.AI.Continue(fmt.Sprintf(`Führe die Geschichte nach dem Input von Spieler %s weiter.`, playerID))
	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: resp.NarratorText})
	hub.Broadcast(g.ID, g)

	go g.addAllMissingAudio()
}

func (g *Game) Start(scenario string, violenceLevel uint8, duration uint8) {
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

	s += "\n\nZiel-Gewaltgrad: " + scenarios.ViolenceLevel(violenceLevel).String()
	s += "\n\nZiel-Länge der gesammten Kampagne: " + scenarios.Duration(duration).String()

	g.State = GameStateRunning
	g.AI, err = ai.New(ctx)
	if err != nil {
		log.Printf("failed to create AI: %v", err)
		return
	}

	for _, player := range g.Players {
		g.AI.EntityData["player_"+player.ID] = player.Description.Slice()
	}
	hub.Broadcast(g.ID, g)

	resp := g.AI.Start(s)
	g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: resp.NarratorText})
	fmt.Printf("First message: %s\n", resp.JSON())
	hub.Broadcast(g.ID, g)
	go g.addAllMissingAudio()
}

func (g *Game) addAllMissingAudio() {
	g.mut.Lock()
	defer g.mut.Unlock()

	fmt.Println("addAllMissingAudio")

	doneSomething := false

	for i, m := range g.AI.ChatHistory {
		if len(m.Audio) > 0 || m.Role != "model" {
			continue
		}
		fmt.Println("addAllMissingAudio", m.Message[:16], "...")
		if audio, err := g.AI.TTS(m.Message); err != nil {
			fmt.Println("error during tts:", err.Error())
			continue
		} else {
			// m is not a reference
			g.AI.ChatHistory[i].Audio = audio
			doneSomething = true
			fmt.Println("DONE: addAllMissingAudio", m.Message[:16], "...")
		}
	}
	if doneSomething {
		hub.Broadcast(g.ID, g)
	}
}

func init() {
	// just for testing
	newWithId("28603f7e-77c7-487b-8d06-548354c35178")
}

const (
	GameStateInit GameState = iota
	GameStateRunning
)

func (pd PlayerData) Slice() []string {
	return []string{
		"name: " + pd.Name,
		"alter: " + pd.Age,
		"aussehen: " + pd.Appearance,
		"herkunft: " + pd.Origin,
	}
}
