const websockets = new Map<string, WebSocket>();

export function useWebSocket(url: string) {
  let ws = websockets.get(url);
  if (ws === undefined) {
    ws = new WebSocket(url);
    websockets.set(url, ws);
  }
  if (ws.readyState === WebSocket.CLOSED) {
    console.log("WebSocket closed");
    ws = new WebSocket(url);
    websockets.set(url, ws);
  }
  return ws;
}
