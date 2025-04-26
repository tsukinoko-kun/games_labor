export type GameData = {
  id: string;
  players: Record<string, Player>;
  state: (typeof GameState)[keyof typeof GameState];
};

export type Player = {
  id: string;
  description: string;
};

export const GameState = {
  INIT: 0,
  RUNNING: 1,
} as const;
