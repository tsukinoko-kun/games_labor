package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gameslabor/internal/games"
	"gameslabor/internal/server/context"
	"gameslabor/internal/server/hub"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	gameState_action struct {
		Action string `json:"action"`
	}

	gameState_setPlayerCharacterDescription struct {
		Player games.PlayerData `json:"player"`
	}

	gameState_startAction struct {
		Scenario      string `json:"scenario"`
		ViolenceLevel uint8  `json:"violence_level"`
		Duration      uint8  `json:"duration"`
	}

	gameState_userInput struct {
		Input string `json:"input"`
	}
)

func gameState(w http.ResponseWriter, r *http.Request) {
	ctx := context.From(w, r)
	dataID := r.URL.Query().Get("id")

	game, gameFound := games.Games[dataID]
	if !gameFound {
		http.NotFound(w, r)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("websocket upgrade error: %v\n", err)
		return
	}
	defer c.Close()

	if err := gameState_sendFullState(c, game); err != nil {
		log.Printf("error sending full state: %v", err)
		return
	}

	hubClient := hub.Register(dataID, c)
	defer hubClient.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("ws read:", err)
			break
		}

		action := gameState_action{}
		{
			jd := json.NewDecoder(bytes.NewReader(message))
			if err := jd.Decode(&action); err != nil {
				log.Println("ws read:", err)
				break
			}
		}

		switch action.Action {
		case "set_player_character_description":
			descriptionAction := gameState_setPlayerCharacterDescription{}
			jd := json.NewDecoder(bytes.NewReader(message))
			if err := jd.Decode(&descriptionAction); err != nil {
				log.Println("ws read:", err)
				break
			}
			go game.SetPlayerDescription(games.Player{ID: ctx.UserID, Description: descriptionAction.Player})
		case "start":
			startAction := gameState_startAction{}
			jd := json.NewDecoder(bytes.NewReader(message))
			if err := jd.Decode(&startAction); err != nil {
				log.Println("ws read:", err)
				break
			}
			go game.Start(startAction.Scenario, startAction.ViolenceLevel, startAction.Duration)
		case "user_input":
			inputAction := gameState_userInput{}
			jd := json.NewDecoder(bytes.NewReader(message))
			if err := jd.Decode(&inputAction); err != nil {
				log.Println("ws read:", err)
				break
			}
			go game.PlayerInput(ctx.UserID, inputAction.Input)
		}
	}
}

func gameState_sendFullState(c *websocket.Conn, game *games.Game) error {
	return c.WriteJSON(games.WsFullOverwrite{Method: "full_overwrite", Value: game})
}
