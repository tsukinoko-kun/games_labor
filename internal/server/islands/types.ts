import { z } from "zod";

export const PlayerDataShema = z.object({
  name: z.string(),
  age: z.string(),
  origin: z.string(),
  appearance: z.string(),
});
export type PlayerData = z.infer<typeof PlayerDataShema>;

export const PlayerShema = z.object({
  id: z.string(),
  description: PlayerDataShema,
});
export type Player = z.infer<typeof PlayerShema>;

export function descriptionEquals(a: PlayerData, b: PlayerData) {
  for (const key in a) {
    if (a[key as keyof PlayerData] !== b[key as keyof PlayerData]) {
      return false;
    }
  }
  return true;
}
export const ChatMessageShema = z.discriminatedUnion("role", [
  z.object({
    role: z.literal("model"),
    message: z.string(),
    audio: z.string().nullable(),
  }),
  z.object({
    role: z.literal("user"),
    player: z.string(),
    message: z.string(),
    audio: z.string().nullable(),
  }),
]);
export type ChatMessage = z.infer<typeof ChatMessageShema>;

export const AIShema = z.object({
  event_plan: z.array(z.string()),
  event_long_history: z.array(z.string()),
  event_short_history: z.array(z.string()),
  chat_history: z.array(ChatMessageShema),
  entity_data: z.record(z.array(z.string())),
});
export type AI = z.infer<typeof AIShema>;

export const DiceRollSchema = z.object({
  message: z.string(),
  difficulty: z.number(),
  result: z.number(),
});
export type DiceRoll = z.infer<typeof DiceRollSchema>;

export const GameState = { INIT: 0, RUNNING: 1 } as const;
export const GameStateShema = z.nativeEnum(GameState);
export type GameState = z.infer<typeof GameStateShema>;

export const GameDataShema = z.object({
  id: z.string(),
  players: z.record(PlayerShema),
  state: GameStateShema,
  ai: AIShema,
  roll: DiceRollSchema.nullable(),
  accepting_input: z.boolean(),
});
export type GameData = z.infer<typeof GameDataShema>;

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
