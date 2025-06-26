Projektname: KI Game Master (Virtuelles Pen and Paper RPG)

Namen der Teammitglieder: Frank Mayer

Besonderheiten des Projekts:
    Benötigte Software: Go, Node.js, Pnpm, Air, Just (optional), Templ
    Umgebungsvariablen: GOOGLE_API_KEY Google Cloud API Key (Cloud Text-to-Speech API + Generative Language API)
    Dev Server starten: `just dev` oder `air`
    Build: `just build`
    Binaries sind im Ordner `bin/` abgelegt

Besondere Leistungen:
    - Mehr Details in ./DOCUMENTATION.md
    - Text-to-Speech
      Ich habe verschiedenste Modelle ausprobiert und hatte auch versucht
      live Audio Unterhaltung mit https://www.agora.io/ zu bauen,
      aber das war zu dem Zeitpunkt noch nicht bereit zu verwenden.
    - Story Generative
      Das Problem ist nicht eine Kampagne zu planen,
      sondern diese konsistent durchzuführen.
      Ein normaler Chat mit einem LLM ist dafür ungeeignet.
    - Datenmodell
      Ich habe eine Datenbank gebaut, die es dem LLM ermöglicht,
      relevante Informationen für später zu speichern.
      Ich wollte das auch mal auf eine Vektor-Datenbank umstellen,
      das funktioniert aber in diesem Fall nicht,
      weil diese eine Query benötigt, um andere Informationen zu finden,
      die damit zusammenhängen.
    - Sessions (multiplayer) und mehrere Kampagnen gleichzeitig
      Jeder Spieler, bekommt eine UUID, welche diesen identifiziert.
      Die UUID wird einem Spieler-Charakter zugeordnet, damit das LLM weiß,
      welcher Charakter etwas tut, wenn ein Spieler eine Nachricht schreibt.
      Auch jede Kampagne verfügt über eine eigene UUID,
      so können mehrere Spieler gleichzeitig mehrere Kampagnen spielen.
    - Da die Daten im Laufe der Kampagne immer größer werden,
      werden Updates zum Game-State inkrementell übertragen.
      Die React Oberfläche verwendet einen eigenen React-Hook,
      der Updaten über ein WebSocket synchronisiert und die Teile der
      Oberfläche neu rendert, deren Daten sich geändert haben.
      Wenn ein Spieler seine Seite neu lädt, wird einmalig der gesamte
      Game-State neu übertragen, bevor die inkrementellen Updates
      weitergegeben werden.
Alles was ich nicht selbst gemacht habt:
    First-Party-Libraries:
    - Google Cloud Generative AI SDK google.golang.org/genai v1.6.0
    - Text to speech Google Cloud SDK cloud.google.com/go/texttospeech v1.13.0
    - HTML Template Engine github.com/a-h/templ v0.3.898
    - WebSocket implementation github.com/gorilla/websocket v1.5.3
    Bilder für Szenarien: OpenAI GPT ImageGen (via https://t3.chat/)
    Favicon: https://www.flaticon.com/free-icons/magic-book

GitHub: https://github.com/tsukinoko-kun/games_labor

Video: https://youtu.be/_UT_s3CVBF0
