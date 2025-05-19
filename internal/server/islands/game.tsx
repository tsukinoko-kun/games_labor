import { type Dispatch, type SetStateAction, useEffect, useState } from "react";
import {
  useGameData,
  setPlayerCharacterDescription,
  startGame,
  userInput,
} from "./gamestate.ts";
import {
  chatMessageId,
  descriptionEquals,
  GameState,
  type PlayerData,
} from "./types.ts";
import { myUserId, seededRandomCharacter, stringToColor } from "./util.ts";

interface Props {
  scenarios: { title: string; id: string; image: string }[];
}

export function Game(props: Props) {
  const g = useGameData();
  switch (g.state) {
    case GameState.INIT:
      return <Init {...props} />;
    case GameState.RUNNING:
      return <RunningGame />;
  }
}

function RunningGame() {
  return (
    <>
      <RunningGameChatHistory />
      <RunningGameInput />
    </>
  );
}

function RunningGameChatHistory() {
  const g = useGameData();
  useEffect(() => {
    const chatMessages = document.getElementsByClassName("chat-message");
    chatMessages
      .item(chatMessages.length - 1)
      ?.scrollIntoView({ behavior: "smooth" });
  });
  return (
    <ul className="max-w-5xl mx-auto pb-64">
      {g.ai.chat_history.map((m) => (
        <li
          key={chatMessageId(m)}
          className="chat-message block p-4 my-4 border border-stone-700 border-solid rounded-md"
        >
          {m.role === "user" ? (
            <p style={{ color: stringToColor(m.player) }}>
              {g.players[m.player]?.description?.name || m.player}
            </p>
          ) : (
            <p className="text-stone-50">Erzähler</p>
          )}
          {m.audio ? (
            <audio controls>
              <source src={m.audio} type="audio/ogg" />
            </audio>
          ) : null}
          <p className="mt-4 text-stone-50">{m.message}</p>
        </li>
      ))}
    </ul>
  );
}

function RunningGameInput() {
  const [value, setValue] = useState("");
  return (
    <form
      className="flex flex-row justify-between fixed bottom-0 left-4 right-4 w-[calc(100dvw-3rem)] h-fit gap-4"
      onSubmit={(ev) => {
        ev.preventDefault();
        userInput(value);
        setValue("");
      }}
    >
      <input
        type="text"
        className="w-[calc(100dvw-3rem)] p-4 bg-stone-800 rounded-md border border-solid border-transparent focus:border-stone-400"
        placeholder="Was tust du?"
        value={value}
        onChange={(ev) => {
          setValue(ev.target.value);
        }}
      />
      <button
        type="submit"
        className="btn"
        onClick={() => {
          userInput(value);
          setValue("");
        }}
      >
        Senden
      </button>
    </form>
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

const descriptionTranslations = new Map([
  ["name", "Name"],
  ["age", "Alter"],
  ["origin", "Herkunft"],
  ["appearance", "Aussehen"],
]);

function Init(props: Props) {
  const [selectedScenario, setSelectedScenario] = useState<string | null>(null);
  const [violenceLevel, setViolenceLevel] = useState<number>(1);
  const [length, setLength] = useState<number>(1);

  return (
    <div className="max-w-7xl px-4 justify-center w-fit mx-auto block my-8 pb-64">
      <InitPlayers />
      <InitScenario
        scenarios={props.scenarios}
        selectedScenario={selectedScenario}
        setSelectedScenario={setSelectedScenario}
      />
      <InitSettings
        violenceLevel={violenceLevel}
        setViolenceLevel={setViolenceLevel}
        length={length}
        setLength={setLength}
      />

      <InitStart
        selectedScenario={selectedScenario}
        violenceLevel={violenceLevel}
        length={length}
      />
    </div>
  );
}

interface InitStartProps {
  selectedScenario: string | null;
  violenceLevel: number;
  length: number;
}

function InitStart(props: InitStartProps) {
  const g = useGameData();
  const playersList = Object.values(g.players);
  const playersWithoutDescription = playersList.filter(
    (p) =>
      !p.description ||
      !p.description.name ||
      !p.description.age ||
      !p.description.appearance ||
      !p.description.origin,
  ).length;
  return (
    <>
      <p className="block text-xl font-bold mb-4 mt-16">
        {playersList.length > 1 ? "Seid ihr bereit?" : "Bist Du bereit?"}
      </p>
      {!props.selectedScenario && (
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
        disabled={!props.selectedScenario || playersWithoutDescription > 0}
        onClick={() => {
          if (props.selectedScenario) {
            startGame(
              props.selectedScenario,
              props.violenceLevel,
              props.length,
            );
          }
        }}
      >
        Los geht's!
      </button>
    </>
  );
}

function InitPlayers() {
  const g = useGameData();
  const [playerDescription, setPlayerDescription] = useState<PlayerData>(
    g.players[myUserId]?.description ?? seededRandomCharacter(),
  );

  const playersList = Object.values(g.players);

  return (
    <>
      <p className="block text-xl font-bold mb-4">Charaktere</p>
      <p className="my-4 text-stone-500">Angemeldete Spieler</p>
      <ul className="flex flex-row flex-wrap gap-4 justify-between">
        {playersList.map((player, i) =>
          player.id === myUserId ? (
            <li
              key={player.id}
              className="w-md border border-stone-500 border-solid rounded-md p-4"
            >
              <p>Charakter {i + 1} (Du)</p>
              <div className="block w-full">
                <p className="text-stone-500">Beschreibung</p>
                <div className="block">
                  {Object.entries(playerDescription).map(([k, v], i) => (
                    <label
                      className={`block bg-stone-800 p-2 border border-solid rounded-md ${v ? "border-stone-700 has-focus:border-stone-400" : "border-orange-400"} ${i > 0 ? "mt-4" : ""}`}
                    >
                      {descriptionTranslations.get(k) ??
                        k[0].toUpperCase() + k.substring(1)}
                      <input
                        className="block w-full"
                        value={v}
                        onChange={(ev) =>
                          setPlayerDescription({
                            ...playerDescription,
                            [k]: ev.target.value,
                          })
                        }
                      />
                    </label>
                  ))}
                </div>
              </div>
              <button
                type="submit"
                className={`btn ${!descriptionEquals(playerDescription, player.description) ? "outline-2 outline-solid outline-orange-400" : ""}`}
                onClick={() => {
                  setPlayerCharacterDescription(playerDescription);
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
              {player.description
                ? Object.entries(player.description).map(([k, v], i) => (
                    <p className="block mt-4">
                      <span className="text-stone-500 mr-2">
                        {descriptionTranslations.get(k) ??
                          k[0].toUpperCase() + k.substring(1)}
                      </span>
                      <span>{v || "N/A"}</span>
                    </p>
                  ))
                : player.id}
            </li>
          ),
        )}
      </ul>
    </>
  );
}

interface InitScenarioProps extends Props {
  selectedScenario: string | null;
  setSelectedScenario: Dispatch<SetStateAction<string | null>>;
}

function InitScenario(props: InitScenarioProps) {
  return (
    <>
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
            selected={props.selectedScenario === scenario.id}
            onClick={() => props.setSelectedScenario(scenario.id)}
          />
        ))}
      </div>
    </>
  );
}

interface InitSettingsProps {
  violenceLevel: number;
  setViolenceLevel: Dispatch<SetStateAction<number>>;

  length: number;
  setLength: Dispatch<SetStateAction<number>>;
}

function InitSettings(props: InitSettingsProps) {
  return (
    <>
      <p className="block text-xl font-bold mb-4 mt-16">Einstellungen</p>
      <p className="my-4 text-stone-500">
        Definiere die Feinheiten der Kampagne
      </p>
      <label className="block my-4">
        <p>
          Gewaltgrad: <span>{violenceLevelToText(props.violenceLevel)}</span>
        </p>
        <input
          className="block w-full max-w-80"
          type="range"
          min={0}
          max={3}
          value={props.violenceLevel}
          onChange={(e) => {
            props.setViolenceLevel(Number(e.target.value));
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
            props.setLength(Number(e.target.value));
          }}
        />
      </label>
    </>
  );
}
