import { useEffect, useState } from "react";
import type { GameData } from "./types.ts";

const websockets = new Map<string, WebSocket>();

export function useWebSocket(url: string) {
  let ws = websockets.get(url);
  if (ws === undefined) {
    ws = new WebSocket(url);
    ws.addEventListener(
      "close",
      () => {
        websockets.delete(url);
      },
      { capture: true, passive: true },
    );
    websockets.set(url, ws);
  }
  if (
    ws.readyState === WebSocket.CLOSED ||
    ws.readyState === WebSocket.CLOSING
  ) {
    ws = new WebSocket(url);
    websockets.set(url, ws);
  }
  return ws;
}

export function useGameState(url: string) {
  const [serverData, setServerData] = useState<GameData | null>(null);
  const ws = useWebSocket(url);

  useEffect(() => {
    if (!url) {
      console.error("WebSocket URL is not provided.");
      return;
    }

    const handleMessage = (event: MessageEvent) => {
      const newData = JSON.parse(event.data);
      setServerData(newData);
    };

    ws.addEventListener("message", handleMessage, {
      capture: false,
      passive: true,
    });

    return () => {
      ws.removeEventListener("message", handleMessage);
    };
  }, [url, ws]);

  (globalThis as any).serverData = serverData;
  return serverData;
}
