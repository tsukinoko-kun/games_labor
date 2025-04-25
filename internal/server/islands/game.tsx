import { useWebSocket } from "./hooks.ts";

type GameData = {
  id: string;
};

const gameWsUri = new URL(location.href);
gameWsUri.pathname = "/api/game_state";
if (gameWsUri.protocol === "http:") {
  gameWsUri.protocol = "ws:";
} else if (gameWsUri.protocol === "https:") {
  gameWsUri.protocol = "wss:";
}
if (gameWsUri.port === "4321") {
  gameWsUri.port = "8080";
}
const gameWsUriString = gameWsUri.toString();

const ws = useWebSocket(gameWsUriString);

ws.onmessage = (ev) => {
  console.log("Received message:", ev.data);
};

ws.onopen = () => {
  console.log("WebSocket connection opened");
  ws.send(JSON.stringify({ action: "join" }));
};

export function Game(game: GameData) {
  return (
    <h1>Game with ID: {game.id}</h1>
  );
}
