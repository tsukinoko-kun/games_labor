import {
  type Dispatch,
  memo,
  type SetStateAction,
  useEffect,
  useState,
} from "react";
import {
  useGameData,
  setPlayerCharacterDescription,
  startGame,
  userInput,
} from "./gamestate.ts";
import {
  chatMessageId,
  descriptionEquals,
  DiceRoll,
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
  const g = useGameData();
  return (
    <>
      <RunningGameChatHistory />
      {g.roll ? <Roll roll={g.roll} /> : null}
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
  if (g.ai.chat_history.length === 0) {
    return (
      <p className="text-xl p-8 text-stone-50">
        Kampagne wird geplant... das dauert einen Moment.
      </p>
    );
  }
  return (
    <ul className="max-w-5xl mx-auto pb-64">
      {g.ai.chat_history.map((m) => (
        <li
          key={chatMessageId(m)}
          className="chat-message block p-4 my-4 border border-stone-700 border-solid rounded-md"
        >
          {m.role === "user" ? (
            <p className="text-xl" style={{ color: stringToColor(m.player) }}>
              {g.players[m.player]?.description?.name || m.player}
            </p>
          ) : (
            <p className="text-stone-50 text-xl">
              Erzähler
              {m.audio ? (
                <audio className="h-[1em] ml-4 inline-block" controls>
                  <source src={m.audio} type="audio/ogg" />
                </audio>
              ) : (
                <span className="text-xs ml-4">Audio wird generiert...</span>
              )}
            </p>
          )}
          <p className="mt-4 text-stone-50">{m.message}</p>
        </li>
      ))}
    </ul>
  );
}

function RunningGameInput() {
  const g = useGameData();
  const [value, setValue] = useState("");
  return (
    <form
      className="flex flex-row justify-between fixed bottom-0 left-4 right-4 w-[calc(100dvw-3rem)] h-fit gap-4"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!g.accepting_input) {
          return;
        }
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
        disabled={!g.accepting_input}
        onClick={() => {
          if (!g.accepting_input) {
            return;
          }
          userInput(value);
          setValue("");
        }}
      >
        Senden
      </button>
    </form>
  );
}

const Roll = memo(
  function DiceRoll(props: { roll: DiceRoll }) {
    return (
      <div className="fixed inset-0 pb-[30vh] bg-stone-950 p-4 z-50">
        <p>{props.roll.message}</p>
        <p className="text-white text-2xl font-bold text-center">
          Schwierigkeit: {props.roll.difficulty}
        </p>
        <p
          className={
            "die-outcome text-2xl font-bold text-center " +
            (props.roll.result >= props.roll.difficulty
              ? "text-green-400"
              : "text-red-400")
          }
        >
          {props.roll.result >= props.roll.difficulty ? "Erfolg" : "Fehlschlag"}
        </p>
        <Die face={props.roll.result} className="h-[256px] w-full" />
      </div>
    );
  },
  (a, b) =>
    a.roll.message === b.roll.message &&
    a.roll.difficulty === b.roll.difficulty &&
    a.roll.result === b.roll.result,
);

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
          Länge: <span>{lengthToText(props.length)}</span>
        </p>
        <input
          className="block w-full max-w-80"
          type="range"
          min={0}
          max={2}
          value={props.length}
          onChange={(e) => {
            props.setLength(Number(e.target.value));
          }}
        />
      </label>
    </>
  );
}

function Die(props: { face: number; className?: string }) {
  return (
    <div className="die_container">
      <div className={"die " + (props.className ?? "")} data-face={props.face}>
        <figure className="face face-1"></figure>
        <figure className="face face-2"></figure>
        <figure className="face face-3"></figure>
        <figure className="face face-4"></figure>
        <figure className="face face-5"></figure>
        <figure className="face face-6"></figure>
        <figure className="face face-7"></figure>
        <figure className="face face-8"></figure>
        <figure className="face face-9"></figure>
        <figure className="face face-10"></figure>
        <figure className="face face-11"></figure>
        <figure className="face face-12"></figure>
        <figure className="face face-13"></figure>
        <figure className="face face-14"></figure>
        <figure className="face face-15"></figure>
        <figure className="face face-16"></figure>
        <figure className="face face-17"></figure>
        <figure className="face face-18"></figure>
        <figure className="face face-19"></figure>
        <figure className="face face-20"></figure>
      </div>
    </div>
  );
}
