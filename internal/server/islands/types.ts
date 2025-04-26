export type GameData = {
  id: string;
  players: Record<string, Player>;
  state: (typeof GameState)[keyof typeof GameState];
  ai: AI;
};

export type Player = {
  id: string;
  description: string;
};

export const GameState = {
  INIT: 0,
  RUNNING: 1,
} as const;

export type AI = {
  event_plan: string[];
  event_long_history: string[];
  event_short_history: string[];
  chat_history: ChatMessage[];
  character_data: Record<string, string[]>;
  place_data: Record<string, string[]>;
  group_data: Record<string, string[]>;
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
