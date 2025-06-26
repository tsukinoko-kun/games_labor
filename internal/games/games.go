package games

import (
	"cmp"
	"context"
	"fmt"
	"gameslabor/internal/ai"
	"gameslabor/internal/games/scenarios"
	"gameslabor/internal/karmicdice"
	"gameslabor/internal/server/hub"
	"log"
	"sync"

	"github.com/google/uuid"
)

type (
	Game struct {
		ID             string             `json:"id"`
		AI             *ai.AI             `json:"ai"`
		Players        map[string]*Player `json:"players"`
		mut            sync.Mutex         `json:"-"`
		Roll           *DiceRoll          `json:"roll"`
		State          GameState          `json:"state"`
		AcceptingInput bool               `json:"accepting_input"`
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

type DiceRoll struct {
	Difficulty uint8 `json:"difficulty"`
	Result     uint8 `json:"result"`
}

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
		id,
		ai.Empty(),
		make(map[string]*Player),
		sync.Mutex{},
		nil,
		GameStateInit,
		false,
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
		log.Printf("game state is not running, can't player input\n")
		return
	}

	if !g.AcceptingInput {
		log.Printf("accepting input is not enabled, can't player input\n")
		return
	}

	g.AcceptingInput = false
	hub.Broadcast(g.ID, WsSetOrPush{"set", "accepting_input", false})

	{
		newChatMessage := ai.ChatMessage{Role: "user", PlayerID: playerID, Message: input}
		g.AI.ChatHistory = append(g.AI.ChatHistory, newChatMessage)
		hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})
	}

	g.continueWithPrompt(fmt.Sprintf(`Führe die Geschichte nach dem Input von Spieler %s weiter.`, playerID))
	go g.addAllMissingAudio()
}

func (g *Game) continueWithPrompt(processingPrompt string) {
	resp := g.AI.Continue(processingPrompt)
	newChatMessage := ai.ChatMessage{Role: "model", Message: resp.NarratorText}
	g.AI.ChatHistory = append(g.AI.ChatHistory, newChatMessage)
	hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})
	if resp.RollDice != nil {
		r := karmicdice.Int(resp.RollDice.Difficulty)
		g.Roll = &DiceRoll{Difficulty: uint8(resp.RollDice.Difficulty), Result: uint8(r)}
	} else {
		g.Roll = nil
		g.AcceptingInput = true
		hub.Broadcast(g.ID, WsSetOrPush{"set", "accepting_input", true})
	}
	hub.Broadcast(g.ID, WsSetOrPush{"set", "roll", g.Roll})
}

func clamp[T cmp.Ordered](min, v, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (g *Game) Start(scenario string, violenceLevel uint8, duration uint8) {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateInit {
		log.Printf("game state is not init, can't start game\n")
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
	g.AcceptingInput = false
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
	if resp.RollDice != nil {
		r := karmicdice.Int(resp.RollDice.Difficulty)
		g.Roll = &DiceRoll{Difficulty: uint8(resp.RollDice.Difficulty), Result: uint8(r)}
	}
	hub.Broadcast(g.ID, WsSetOrPush{"push", "ai.chat_history", newChatMessage})
	hub.Broadcast(g.ID, WsSetOrPush{"set", "roll", g.Roll})

	g.AcceptingInput = true
	hub.Broadcast(g.ID, WsSetOrPush{"set", "accepting_input", true})

	go g.addAllMissingAudio()
}

func (g *Game) ContinueAfterRoll() {
	g.mut.Lock()
	defer g.mut.Unlock()

	if g.State != GameStateRunning {
		log.Printf("game state is not running, can't continue after roll\n")
		return
	}

	if g.Roll == nil {
		log.Printf("no roll, can't continue after roll\n")
		return
	}

	hub.Broadcast(g.ID, WsSetOrPush{"set", "roll", nil})

	if g.Roll.Result >= g.Roll.Difficulty {
		g.continueWithPrompt(fmt.Sprintf("Es wurde eine %d von %d gewürfelt, der Roll ist damit erfolgreich. Führe die Geschichte fort.", g.Roll.Result, g.Roll.Difficulty))
	} else {
		g.continueWithPrompt(fmt.Sprintf("Es wurde eine %d von %d gewürfelt, der Roll ist damit fehlgeschlagen. Führe die Geschichte fort.", g.Roll.Result, g.Roll.Difficulty))
	}
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
	// g.AddPlayer("user_01JXABDXJS92BR5G5CH710QE0H")
	// g.Players["user_01JXABDXJS92BR5G5CH710QE0H"].Description = PlayerData{
	// 	Name:       "Spieler 1",
	// 	Age:        "20",
	// 	Origin:     "Deutschland",
	// 	Appearance: "Männlich",
	// }
	// g.AI, _ = ai.New(context.Background())
	// g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: "Du trinkst gerade dein Bir in einer lokalen Bar, als prötzlich drei Garnoven durch die Tür schreiten. Mit Waffen auf dich und andere zielend sagen sie: \"Das ist ein Überfall, alle die Hände hoch und keiner rührt sich!\""})
	// g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "user", PlayerID: "user_01JXABDXJS92BR5G5CH710QE0H", Message: "Ich hole ein Messer aus meinem Holster am Bein und greife die Garnoven an."})
	// g.AI.ChatHistory = append(g.AI.ChatHistory, ai.ChatMessage{Role: "model", Message: "Ok, würfle, ob du es schaffst einen anzugreifen, ohne, dass er dich verletzt."})
	// g.AcceptingInput = false
	// g.State = GameStateRunning
	// g.Roll = &DiceRoll{Difficulty: 4, Result: 5}
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
