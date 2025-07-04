# Games Labor

Frank Mayer

## UI

Die UI ist mit React und TailwindCSS gebaut.
Die React Components werden als Islands auf die HTML Seite gerendert.
Das HTML kommt von [Templ](https://templ.dev/).

Das Index-Seite ist statisch (kein React).
Die Seite für das Spiel ist eine React single page application.

Ein eigener React-Hook `useGameData` wird verwendet, um den aktuellen Game-State über das WebSocket zum Server zu synchronisieren.

![](docs/screensho_setup.jpg)

![](docs/screenshot_ingame.jpg)

## Models

In meinem Test hat die Gemini Familie allgemein die beste Qualität an Text-Output.

Dabei habe ich getestet, wie gut das LLM einen Ansprechenden Text generieren kann. Ich habe Instruktionen angegeben, wie ein Text gut lesbar ist, interessanten Rhythmus hat ect.
Ich habe getestet:

- Gemini 2.0 Flash
- Gemini 2.0 Flash Lite
- Gemini 2.5 Flash
- Gemini 2.5 Pro
- GPT o3-mini
- GPT o4-mini
- Claude 3.5 Sonnet
- Claude 4 Sonnet
- DeepSeek R1
- DeepSeek v3
- Llama 4 Scout
- Llama 4 Maverick
- Grok 3
- Qwen qwq-32b

Dabei waren die Ergebnisse von Gemini mit großem Abstand am besten.

Ursprünglich hatte ich Gemini 2.5 Pro als Modell für die initiale Generierung der Story verwendet, weil dieses Modell bessere Story-Plane generiert. Das dauert aber zu lange.
Vor kurzem wurde Gemini 2.5 Flash mit Reasoning ausgestattet, was es gut genug macht, um die Story-Plane zu generieren.

Ich hatte Gemini 2.5 Flash Lite als Haupt-Modell vorgesehen, weil es noch schneller ist. Haupt-Modell heißt, es generiert alles nach dem initialen Story-Plan. Flash Lite hat sich aber schlecht and die Vorgaben gehalten. Beispielsweise hat es meine Anweisung komplett ignoriert dem Spieler zu sagen wofür er gerade würfelt.

Jetzt verwende ich Gemini 2.5 Flash für alles.

Die Geschwindigkeit variiert auch nach Tageszeit, was die Wahl des Modells erschwert.

## Eigene Datenbank

### Architektur

![](docs/architecture.jpg)

Normalerweise kennt ein LLM einen bestimmten Bereich aus der Chat-Historie, den es als Kontext verwenden kann, um seine Antworten zu generieren. Da das sehr unstrukturiert ist, kann es vorkommen, dass das LLM bei längeren Chats vergisst, worum es eigentlich geht. Er vergisst, dass wir ein Pen and Paper spielen und was der ursprüngliche Plan für die Kampagne war.

Um das zu lösen, habe ich die Daten in verschiedene Gebiete eingeteilt.

Ein Bereich für den groben Plan der Kampagne. Hier wird zu Beginn ein Pfad geplant, den die Geschichte nehmen soll.

Weite und nahe Vergangenheit, um zu speichern, was passiert ist. Die nahe Vergangenheit ist detaillierter, aber deckt nur den Teil der Geschehnisse ab, die erst vor kurzem passiert sind.

Anfangs hatte ich extra Speicher für verschiedene Arten von Entitäten, wie Orte, Personen, Gegenstände, etc. Diese sind jetzt aber als Entitäten zusammengefasst, da es teilweise Überschneidungen zwischen den Arten gab und ich keine Lust mehr hatte immer noch eine neue Art von Entität anzulegen.

Die Chat Historie wird auch teilweise mitgegeben. Das dient dazu, das Pacing zu verbessert. Wenn das LLM weiß, was es zuletzt gesagt hat, ist es wahrscheinlicher, dass es sich nicht ständig wiederholt.

Das LLM ist angewiesen, neue Einträge für diese Speicher zurückzugeben, und es bekommt bei jeder neuen Anfrage die gesamten Daten mit, anstatt der Chat Historie. Da diese Daten so organisierter sind, ist es für das LLM sehr einfach, sie sinnvoll zu nutzen.

### JSON Schema

Die Gemini API kann so konfiguriert werden, dass sie eine JSON Schema bei der Antwort verwendet.
So kann sichergestellt werden, dass die Antwort vom Server auch gelesen werden kann.

Schema ist in der Datei `internal/ai/schema.go` definiert.

### Prompts

System Prompt `internal/ai/system.txt`

Start Prompt `internal/ai/start.txt`

`%s` ist der Platzhalter für das Szenario, das vom Spieler ausgewählt wird.
Diese sind in `internal/games/scenarios/` abgelegt.

## Embeddings

Ich habe mich wärend der Implementierung gegen eine Vektor Datenbank entschieden. Im Rest dieses Kapitels erkläre ich trotzdem, was ich darüber herausgefunden habe und was genau ich für meine Tests verwendet habe.

Ich bin von der eigenen Datenbank zur Vektor Datenbank Qdrant gewechselt um dem LLM mehr Informationen zur Verfügung zu stellen.

Die Antwort vom LLM (Gemini) ist bereits im JSON Format, was durch seine Struktur gut geeignet ist.
Dieser Text kann in ein Embedding umgeformt werden.
Das sind dann ein Haufen Vektoren, die die Relation zwischen den einzelnen Teilen des Textes darstellen.

Diese Vektoren können dann in einer Vektor Datenbank wie Qdrant gespeichert werden.

Das Embedding Model `text-embedding-004` ist das neueste allgemein verfügbare Texteinbettungsmodell von Google. Es ist für die hochwertige semantische Suche konzipiert. Dieses Modell gibt Embeddings in 768 Dimensionen aus.

Die von `text-embedding-004` erzeugten Embeddings sind L2-normalisiert. Wenn Vektoren normalisiert sind, misst die cosine similarity (Kosinusähnlichkeit) die semantische Nähe, indem sie den Winkel zwischen ihnen betrachtet. Das ist wohl die empfohlene Metrik für solche Embeddings.

what is L2 normalisation?

> For the L2 norm, you square all the data, add them up and then square root them. Then, to normalize the data, you divide each data point by this value. The L2 norm is the most common, but there are others (such as L1) that are slightly different, but do similar things. This is common when the magnitude doesnt matter. Directions are common.

lethal_rads (Reddit) https://www.reddit.com/r/explainlikeimfive/comments/18smr7z/eli5_what_is_l2_normalisation_how_is_it_different/

Jetzt kommt das Problem: Wie kommt man wieder an die Daten?

Ich dachte, dass ich die Vektor Datenbank einfach dem LLM zur Verfügung stellen könnte und dieses kann die Daten zur Generierung der Antwort nutzen kann.
Das funktioniert so nicht.
Man benötigt eine Query.
Das ist eine Information, mit deren Hilfe man in der Datenbank nach semantisch ähnlichen Daten suchen kann.
Da ich das LLM aber in einem rein kreativen Kontext nutze, es also "aus dem nichts" Content generieren soll, ist es nicht möglich eine sinnvolle Query zu formulieren.
Für die Generierung sind immer alle Daten relevant.

Ein kleines Detail, das vor Stunden in der Story passiert ist, kann wieder relevant werden.
Bzw. es kann relevant gemacht werden.
Genau hier liegt das der Ursprung des Problems.
Durch die kreative Art wie das LLM generiert, entscheidet es eher willkürlich, welche Informationen für das Fortsetzen der Story genutzt werden.
Um eine Query für die nötigen Daten erstellen zu können, werden genau diese Daten benötigt.

## Sessions

Jeder Browser, der noch keinen ID-Cookie hat, bekommt eine neue UUID als Cookie gesetzt.
Diese wird verwendet, um den Spieler (bzw. seinen Browser) zu identifizieren.

![](docs/session.jpg)

## Synchronisation

Zu Beginn bekommt jeder Browser den kompletten, aktuellen Game-State.
Danach werden alle Updates des Game-State inkrementell übertragen um die Datenübertragungsrate zwischen dem Server und dem Browser zu minimieren.

Der Browser kann jederzeit die Seite neu laden oder zwischen verschiedenen Kampagnen wechseln,
ohne, dass es dadurch zu Datenverlust oder Unterbrechungen bei den anderen Browsern in der Kampagne kommt.

Die React Oberfläche rendert die updates auch inkrementell und scrollt automatisch zur neuesten Nachricht usw.

![](docs/sync.jpg)

## Multiplayer

In einer Kampagne können theoretisch beliebig viele Spieler gleichzeitig teilnehmen.
Jeder Spieler wird über seine UUID identifiziert.
Die Daten zu seinem Spieler-Charakter werden auch dieser UUID zugeordnet, damit das LLM weiß, welcher Charakter zu welchem Spieler gehört.
