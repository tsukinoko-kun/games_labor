export type GameData = {
  id: string;
  players: Record<string, Player>;
  state: (typeof GameState)[keyof typeof GameState];
  ai: AI;
};

export type PlayerData = {
  name: string;
  age: string;
  origin: string;
  appearance: string;
};

export type Player = {
  id: string;
  description: PlayerData;
};

export function descriptionEquals(a: PlayerData, b: PlayerData) {
  for (const key in a) {
    if (a[key as keyof PlayerData] !== b[key as keyof PlayerData]) {
      return false;
    }
  }
  return true;
}

export const GameState = {
  INIT: 0,
  RUNNING: 1,
} as const;

export type AI = {
  event_plan: string[];
  event_long_history: string[];
  event_short_history: string[];
  chat_history: ChatMessage[];
  entity_data: Record<string, string[]>;
};

export type ChatMessage =
  | {
      role: "model";
      message: string;
    }
  | {
      role: "user";
      player: string;
      message: string;
    };

function hashStr(str: string) {
  let hash = 5381;

  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash) ^ str.charCodeAt(i);
  }

  return (hash >>> 0).toString(36);
}

export function chatMessageId(m: ChatMessage): string {
  if (m.role === "user") {
    return `${m.role}-${m.player}-${hashStr(m.message)}`;
  }
  return `${m.role}-${hashStr(m.message)}`;
}
