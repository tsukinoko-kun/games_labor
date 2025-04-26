function error(message: string) {
  console.error(message)
  alert(message)
}

export function setPlayerCharacterDescription(ws: WebSocket, description: string) {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't set player character description, WebSocket is not open")
    return
  }

  ws.send(JSON.stringify({
    action: "set_player_character_description",
    value: description,
  }))
}

export function startGame(ws: WebSocket, selectedScenario: string, violenceLevel: number, duration: number) {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't start game, WebSocket is not open")
    return
  }

  if (!selectedScenario) {
    error("can't start game, no scenario selected")
    return
  }

  ws.send(JSON.stringify({
    action: "start",
    scenario: selectedScenario,
    violenceLevel: violenceLevel,
    duration: duration,
  }))
}
