import { GameDataShema, PlayerData, type GameData } from "./types.ts";
import { Sync } from "./sync.ts";
import z from "zod";
import { zodErr } from "./util.ts";
import { useSyncExternalStore } from "react";

const gameWsUri = new URL(location.href);
gameWsUri.pathname = "/api/game_state";
if (gameWsUri.protocol === "http:") {
  gameWsUri.protocol = "ws:";
  if (gameWsUri.port === "4321") {
    // this is for Air not handling websocket
    gameWsUri.port = "8080";
  }
} else if (gameWsUri.protocol === "https:") {
  gameWsUri.protocol = "wss:";
}

const ws = new WebSocket(gameWsUri.toString());

const gameSync = new Sync<GameData>({
  id: "",
  players: {},
  state: 0,
  ai: {
    event_plan: [],
    event_long_history: [],
    event_short_history: [],
    chat_history: [],
    entity_data: {},
  },
  roll: null,
  accepting_input: false,
});

const WsFullOverwrite = z.object({
  method: z.literal("full_overwrite"),
  value: GameDataShema,
});
const WsSet = z.object({
  method: z.literal("set"),
  path: z.string().nonempty(),
  value: z.any(),
});
const WsPush = z.object({
  method: z.literal("push"),
  path: z.string().nonempty(),
  value: z.any(),
});

const WsDataSchema = z.discriminatedUnion("method", [
  WsFullOverwrite,
  WsSet,
  WsPush,
]);

ws.addEventListener(
  "message",
  (ev) => {
    console.debug("from ws:", ev.data);
    const dataObject =
      typeof ev.data === "object" ? ev.data : JSON.parse(ev.data);

    const resp = WsDataSchema.safeParse(dataObject);
    if (resp.success) {
      switch (resp.data.method) {
        case "full_overwrite":
          gameSync.override(resp.data.value);
          break;
        case "set":
          gameSync.set(resp.data.path, resp.data.value);
          break;
        case "push":
          gameSync.push(resp.data.path, resp.data.value);
          break;
      }
    } else {
      throw new Error(
        "invalid message format received from server:\n" +
          resp.error.issues.map(zodErr).join("\n\n"),
      );
    }
  },
  { capture: false, passive: true },
);

export function useGameData() {
  return useSyncExternalStore((onStoreChange) => {
    gameSync.subscribe(onStoreChange);
    return () => gameSync.unsubscribe(onStoreChange);
  }, gameSync.getSnapshot.bind(gameSync)).data;
}

function error(message: string) {
  console.error(message);
  alert(message);
}

export function setPlayerCharacterDescription(description: PlayerData) {
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

export function userInput(input: string) {
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

export function continueAfterRoll() {
  if (ws.readyState !== WebSocket.OPEN) {
    error("can't send user input, WebSocket is not open");
    return;
  }

  ws.send(
    JSON.stringify({
      action: "continue_after_roll",
    }),
  );
}
