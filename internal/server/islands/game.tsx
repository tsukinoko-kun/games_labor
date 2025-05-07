import { useState } from "react";
import { setPlayerCharacterDescription, startGame } from "./api.ts";
import { useGameState, useWebSocket } from "./hooks.ts";
import {
  chatMessageId,
  descriptionEquals,
  type GameData,
  GameState,
  PlayerData,
} from "./types.ts";
import {
  myUserId,
  seededRandomAppearance,
  seededRandomCharacter,
  seededRandomInt,
  seededRandomName,
  seededRandomOrigin,
  stringToColor,
} from "./util.ts";

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

interface Props {
  scenarios: { title: string; id: string; image: string }[];
}

interface ExtProps extends Props {
  game: GameData;
}

export function Game(props: Props) {
  const game = useGameState(gameWsUriString);
  if (!game) {
    return <p>Loading...</p>;
  }
  switch (game.state) {
    case GameState.INIT:
      return <Init {...props} game={game} />;
    case GameState.RUNNING:
      return <RunningGame {...props} game={game} />;
    default:
      return <p>Unknown game state {JSON.stringify(game.state)}</p>;
  }
}

function RunningGame(props: ExtProps) {
  return (
    <div>
      <ul>
        {props.game.ai?.chat_history?.map((m) => {
          if (m.role === "user") {
            return (
              <li
                key={chatMessageId(m)}
                style={{ color: stringToColor(m.player) }}
              >
                {m.message}
              </li>
            );
          } else {
            return (
              <li key={chatMessageId(m)} className="text-blue-500">
                {m.message}
              </li>
            );
          }
        })}
      </ul>
      <div className="grid grid-cols-2 gap-4">
        <input type="text" placeholder="Type your message..." />
        <button type="submit" className="btn">
          Senden
        </button>
      </div>
    </div>
  );
}

function InitScenarioButton(props: {
  title: string;
  imgSrc: string;
  id: string;
  selected: boolean;
  onClick?: () => void;
}) {
  return (
    <button
      type="button"
      className={`w-72 block pointer-events-auto cursor-pointer group-hover:opacity-50 hover:opacity-100 transition-opacity border border-solid rounded-md p-4 ${props.selected ? "bg-stone-900 border-stone-500 text-white" : "border-stone-700"}`}
      onClick={props.onClick}
    >
      <img
        src={props.imgSrc}
        alt=""
        className="block rounded-md aspect-[3/2]"
        draggable={false}
      />
      <p className="block mt-4">{props.title}</p>
    </button>
  );
}

function violenceLevelToText(level: number): string {
  switch (level) {
    case 0:
      return "Gar nicht gewalttätig";
    case 1:
      return "Leicht gewalttätig";
    case 2:
      return "Gewalttätig und grausam";
    case 3:
      return "Übertrieben gewalttätig, grausam und unangenehm";
    default:
      return "Unbekannt";
  }
}

function lengthToText(length: number): string {
  switch (length) {
    case 0:
      return "Sehr kurz (30-60 Minuten)";
    case 1:
      return "Kurz (2-4 Stunden)";
    case 2:
      return "Lang (4-8 Stunden)";
    default:
      return "Unbekannt";
  }
}

function InitCharacter(props: {
  value: PlayerData;
  onChange: (newValue: PlayerData) => void;
}) {
  return (
    <div className="block">
      {Object.entries(props.value).map(([k, v], i) => (
        <label
          className={`block bg-stone-800 p-2 border border-solid rounded-md ${v ? "border-stone-700 has-focus:border-stone-400" : "border-orange-400"} ${i > 0 ? "mt-4" : ""}`}
        >
          {k[0].toUpperCase() + k.substring(1)}
          <input
            className="block w-full"
            value={v}
            onChange={(ev) =>
              props.onChange({
                ...props.value,
                [k]: ev.target.value,
              })
            }
          />
        </label>
      ))}
    </div>
  );
}

function Init(props: ExtProps) {
  const ws = useWebSocket(gameWsUriString);
  const [selectedScenario, setSelectedScenario] = useState<string | null>(null);
  const [violenceLevel, setViolenceLevel] = useState<number>(1);
  const [length, setLength] = useState<number>(1);
  const [playerDescription, setPlayerDescription] = useState<PlayerData>(
    props.game.players[myUserId]?.description ?? seededRandomCharacter(),
  );

  const playersList = Object.values(props.game.players);
  const playersWithoutDescription = playersList.filter(
    (p) => !p.description,
  ).length;

  return (
    <div className="max-w-7xl px-4 justify-center w-fit mx-auto block my-8 pb-64">
      <p className="block text-xl font-bold mb-4">Charaktere</p>
      <p className="my-4 text-stone-500">Angemeldete Spieler</p>
      <ul className="flex flex-row flex-wrap">
        {playersList.map((player, i) =>
          player.id === myUserId ? (
            <li
              key={player.id}
              className="w-md border border-stone-500 border-solid rounded-md p-4"
            >
              <p>Charakter {i + 1} (Du)</p>
              <label className="block w-full">
                <p className="text-stone-500">Beschreibung</p>
                <InitCharacter
                  value={playerDescription}
                  onChange={(x) => {
                    setPlayerDescription(x);
                  }}
                />
              </label>
              <button
                type="submit"
                className={`btn ${!descriptionEquals(playerDescription, player.description) ? "outline-2 outline-solid outline-orange-400" : ""}`}
                onClick={() => {
                  setPlayerCharacterDescription(ws, playerDescription);
                }}
              >
                Speichern
              </button>
            </li>
          ) : (
            <li
              key={player.id}
              className="w-md border border-stone-500 border-solid rounded-md"
            >
              <p>Character {i + 1}</p>
              <p className="text-stone-500">Name</p>
              <p>{player.description?.name || player.id}</p>
            </li>
          ),
        )}
      </ul>

      <p className="text-xl font-bold mb-4 mt-16">Wähle ein Szenario</p>
      <p className="my-4 text-stone-500">
        Dies bestimmt das grundlegende Setting deiner Kampagne.
      </p>
      <div className="flex flex-row flex-wrap gap-8 group pointer-events-none">
        {props.scenarios.map((scenario) => (
          <InitScenarioButton
            key={scenario.id}
            title={scenario.title}
            imgSrc={scenario.image}
            id={scenario.id}
            selected={selectedScenario === scenario.id}
            onClick={() => setSelectedScenario(scenario.id)}
          />
        ))}
      </div>

      <p className="block text-xl font-bold mb-4 mt-16">Einstellungen</p>
      <p className="my-4 text-stone-500">
        Definiere die Feinheiten der Kampagne
      </p>
      <label className="block my-4">
        <p>
          Gewaltgrad: <span>{violenceLevelToText(violenceLevel)}</span>
        </p>
        <input
          className="block w-full max-w-80"
          type="range"
          min={0}
          max={3}
          value={violenceLevel}
          onChange={(e) => {
            setViolenceLevel(Number(e.target.value));
          }}
        />
      </label>
      <label className="block my-4">
        <p>
          Länge: <span>{lengthToText(length)}</span>
        </p>
        <input
          className="block w-full max-w-80"
          type="range"
          min={0}
          max={2}
          value={length}
          onChange={(e) => {
            setLength(Number(e.target.value));
          }}
        />
      </label>

      <p className="block text-xl font-bold mb-4 mt-16">
        {playersList.length > 1 ? "Seid ihr bereit?" : "Bist Du bereit?"}
      </p>
      {!selectedScenario && (
        <p className="my-4 text-orange-400">Bitte wähle eine Szenario</p>
      )}
      {playersWithoutDescription > 0 && (
        <p className="my-4 text-orange-400">
          Es haben noch nicht alle Spieler eine Charakterbeschreibung angegeben
        </p>
      )}
      <button
        type="submit"
        className="btn"
        disabled={!selectedScenario || playersWithoutDescription > 0}
        onClick={() => {
          if (selectedScenario) {
            startGame(ws, selectedScenario, violenceLevel, length);
          }
        }}
      >
        Los geht's!
      </button>
    </div>
  );
}
