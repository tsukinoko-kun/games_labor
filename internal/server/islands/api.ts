import type { PlayerData } from "./types";

function error(message: string) {
  console.error(message);
  alert(message);
}

export function setPlayerCharacterDescription(
  ws: WebSocket,
  description: PlayerData,
) {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't set player character description, WebSocket is not open");
    return;
  }

  ws.send(
    JSON.stringify({
      action: "set_player_character_description",
      player: description,
    }),
  );
}

export function startGame(
  ws: WebSocket,
  selectedScenario: string,
  violenceLevel: number,
  duration: number,
) {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't start game, WebSocket is not open");
    return;
  }

  if (!selectedScenario) {
    error("can't start game, no scenario selected");
    return;
  }

  ws.send(
    JSON.stringify({
      action: "start",
      scenario: selectedScenario,
      violence_level: violenceLevel,
      duration: duration,
    }),
  );
}

export function userInput(ws: WebSocket, input: string) {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't send user input, WebSocket is not open");
    return;
  }

  input = input.trim();

  if (!input) {
    return;
  }

  ws.send(
    JSON.stringify({
      action: "user_input",
      input,
    }),
  );
}
