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

type (
	WsFullOverwrite struct {
		Method string `json:"method"`
		Value  *Game  `json:"value"`
	}
	WsSetOrPush struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Value  any    `json:"value"`
	}
)

var Games = make(map[string]*Game)

func New() *Game {
	return newWithId(uuid.NewString())
}

func newWithId(id string) *Game {
	game := &Game{
		ID:      id,
		AI:      ai.Empty(),
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

func (g *Game) SetPlayerDescription(p Player) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if player, ok := g.Players[p.ID]; ok {
		*player = p
		hub.Broadcast(g.ID, WsSetOrPush{"set", "players." + p.ID, p})
	}
}

func (g *Game) PlayerInput(playerID string, input string) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateRunning {
		return
	}

	{
		newChatMessage := ai.ChatMessage{Role: "user", PlayerID: playerID, Message: input}
		g.AI.ChatHistory = append(g.AI.ChatHistory, newChatMessage)
		hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})
	}

	{
		resp := g.AI.Continue(fmt.Sprintf(`Führe die Geschichte nach dem Input von Spieler %s weiter.`, playerID))
		newChatMessage := ai.ChatMessage{Role: "model", Message: resp.NarratorText}
		g.AI.ChatHistory = append(g.AI.ChatHistory, newChatMessage)
		hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})
	}

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
	hub.Broadcast(g.ID, WsFullOverwrite{Method: "full_overwrite", Value: g})

	resp := g.AI.Start(s)
	newChatMessage := ai.ChatMessage{Role: "model", Message: resp.NarratorText}
	g.AI.ChatHistory = append(g.AI.ChatHistory, newChatMessage)
	fmt.Printf("First message: %s\n", resp.JSON())
	hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})

	go g.addAllMissingAudio()
}

func (g *Game) addAllMissingAudio() {
	g.mut.Lock()
	defer g.mut.Unlock()

	for i, m := range g.AI.ChatHistory {
		if len(m.Audio) > 0 || m.Role != "model" {
			continue
		}
		if audio, err := g.AI.TTS(m.Message); err != nil {
			fmt.Println("error during tts:", err.Error())
			continue
		} else {
			g.AI.ChatHistory[i].Audio = audio
			hub.Broadcast(
				g.ID,
				WsSetOrPush{
					"set",
					fmt.Sprintf("ai.chat_history.%d.audio", i),
					audio,
				},
			)
		}
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
