package ai

import (
	"google.golang.org/genai"
)

var (
	falsePtr               = genai.Ptr(false)
	llmResponseGenaiSchema = &genai.Schema{
		Type:        genai.TypeObject,
		Nullable:    falsePtr,
		Description: "Alles was die Spieler sehen ist `narrator_text` und wenn sie selbst würfeln müssen. Alles andere wird vor den Spielern verborgen.",
		Required:    []string{"narrator_text", "place", "event_plan", "event_long_history", "event_short_history", "entity_data"},
		Properties: map[string]*genai.Schema{
			"narrator_text": {
				Type:        genai.TypeString,
				Nullable:    falsePtr,
				Description: "Verwende `narrator_text` um den Spielern etwas als Game Master zu sagen. Passe deine Wortwahl so an, dass sie zum Setting der Geschichte passt. Lass dich von der Wortwahl der Spieler nicht beeinflussen. Achte bei der Formulierung der Texte darauf, dass sie sich gut lesen lassen. Dafür sollten aufeinanderfolgende Sätze unterschiedlich lang sein. Achte darauf wie eine Situation gerade für die Spieler ist und passe die Struktur der Sätze so an, dass das zusammen passt. Hektische Szenen wirken beispielsweise besser, wenn du mehr kurze Sätze verwendest. In sehr ruhigen Situationen kannst du mehr lange Sätze verwenden. Du beschreibst dem Spieler, was sein Charakter sieht, hört und fühlt. Du beschreibst auch die Umgebung, die sich um den Charakter herum befindet. Halte den Fokus dabei auf der Geschichte und kommuniziere mit dem Spieler als sein Charakter anstatt mit dem Spieler als Spieler. Alle Beschreibungen sollten das wiederspiegeln, was die Spieler-Charaktere erlegen. Es ist also keine objektive Beobachtung. Es ist okay nicht direkt jedes Detail zu erwähnen. Du kannst auch Details auslassen und später dazu generieren. Auf jeden Fall solltest du alle Details (in `narrator_text` erzählt odernicht) in `event_long_history` oder `event_short_history` speichern um sie später aufgreifen zu können. Beachte dabei den unterschied zwischen `event_long_history` und `event_short_history`. `event_long_history` ist für längere Ereignisse und Details, während `event_short_history` für kurze Ereignisse und Details verwendet wird, die später nicht mehr relevant sind.",
			},
			"place": {
				Type:        genai.TypeString,
				Nullable:    falsePtr,
				Description: "Verwende `place` um den Spielern zu vermitteln, wo sie sich gerade befinden. Änderst du den Wert von `place`, wird auch `event_short_history` geleert. Informationen, die immer noch relevant sind, musst du dann neu hinzufügen, indem du sie wieder in `event_short_history` schreibst, oder du schreibst eine Zusammenfassung davon in `event_long_history`, wenn sie auf lange Zeit relevant sind.",
			},
			"event_plan": {
				Type:     genai.TypeArray,
				Nullable: falsePtr,
				Items: &genai.Schema{
					Type:     genai.TypeString,
					Nullable: falsePtr,
				},
				Description: "Verwende `event_plan` um den Plan der Geschichte zu erweitern. Sei für Ereignisse, die weit in der Zukunft liegen wage, um flexibel zu bleiben. Wenn ein Ereignis zeitnah stattfinden soll, sollte dieses seht genau beschrieben werden. Schreibe hier alles rein, was du benötigst, um eine konsistente und geplante Geschichte erzählen zu können. Achte darauf, dass die Geschichte in ihrer Gesamtheit einem Ziel folgt. Versuche spezifisch zu sein, um die Geschichte stabiler und konsistenter zu halten.",
			},
			"event_long_history": {
				Type:     genai.TypeArray,
				Nullable: falsePtr,
				Items: &genai.Schema{
					Type:     genai.TypeString,
					Nullable: falsePtr,
				},
				Description: "Verwende `event_long_history` um ein größeres Geschehen zu erfassen. Das gilt für alle Ereignisse, die die Geschichte weiterführen und auf lange Sicht Einfluss haben (auch wenn der Einfluss klein ist). Das ist dein Langzeitgedächtnis. Gib hier auch alles an, was du an Hintergrundinformationen zur Welt, Geschichte geschrieben hast, also Orte, Religionen, Kulturen, Gegebenheiten, und sonstiges. Vor allem alles was mit der Hauptgeschichte zu tun hat. Sei hier sehr spezifisch. Es reichen klare Fakten. Hier muss nichts schön ausformuliert sein. Du kannst auch im Hintergrund Ereignisse geschehen lassen und diese nur in `event_long_history` speichern, ohne sie dem Spieler über `narrator_text` zu sagen, wenn die Spieler-Charaktere das Ereignis nichts mitbekommt. Gib alle Informationen, die du kennst auch spezifisch an.",
			},
			"event_short_history": {
				Type:     genai.TypeArray,
				Nullable: falsePtr,
				Items: &genai.Schema{
					Type:     genai.TypeString,
					Nullable: falsePtr,
				},
				Description: "Verwende `event_short_history` um ein Geschehen zu erfassen. Hierbei geht es um Ereignisse, die nur vorübergehend relevant sind. Diese Ereignisse an den aktuellen Ort der Geschichte gebunden, wenn die Spieler den Ort verlasen, kannst du eine Zusammenfassung der wichtigsten Ereignisse in `event_long_history` speichern. Sei hier sehr spezifisch. Es reichen klare Fakten. Hier muss nichts schön ausformuliert sein. Schreibe hier rein, wenn ein Kampf beginnt, ein Charakter eine Aktion durchführt, ein Charakter eine Beobachtung macht oder sich bewegt, oder wenn etwas anderes passiert. Du solltest durch diese Informationen wissen, was in den letzten Minuten passiert ist, wer wo ist, welcher nicht-Spieler-Charakter was vor hat, wie die Umgebung aufgebaut ist, ect.",
			},
			"entity_data": {
				Type:     genai.TypeArray,
				Nullable: falsePtr,
				Items: &genai.Schema{
					Type:     genai.TypeObject,
					Nullable: falsePtr,
					Required: []string{"entity", "data"},
					Properties: map[string]*genai.Schema{
						"entity": {
							Type:     genai.TypeString,
							Nullable: falsePtr,
						},
						"data": {
							Type:     genai.TypeString,
							Nullable: falsePtr,
						},
					},
				},
				Description: "Verwende `entity_data` um Daten zu Charakteren, Gruppen, Orten und Objekten zu speichern. Hierbei geht es um Daten, die für die Entität relevant sind, aber nicht für die Welt oder die Geschichte. Diese Daten können zum Beispiel die aktuelle Position des Charakters oder das aktuelle Inventar des Charakters sein. Du kannst auch Daten zu Objekten speichern, die für die Entität relevant sind, aber nicht für die Welt oder die Geschichte. Bei beweglichen Entitäten kann die aktuelle Position relevant sein. Bei fühlenden Entitäten kann die Beziehung zu anderen Entitäten relevant sein. Benenne die Entität sinnvoll und spezifisch, damit du sie später eindeutig identifizieren kannst. Die Spieler werden mit als entity player_{UUID} referenziert.",
			},
			"roll_dice": {
				Type:        genai.TypeObject,
				Nullable:    falsePtr,
				Description: "Verwende `roll_dice` um einen Spieler würfeln zu lassen. Nutze das, wenn ein Spieler etwas tun will oder muss, das für diesen nicht selbstverständlich machbar ist. Wenn es hingegen unmöglich ist, muss der Spieler nicht würfeln, er darf das dann einfach nicht tun.",
				Required:    []string{"difficulty"},
				Properties: map[string]*genai.Schema{
					"difficulty": {
						Type:     genai.TypeInteger,
						Nullable: falsePtr,
						Minimum:  genai.Ptr[float64](1),
						Maximum:  genai.Ptr[float64](20),
					},
				},
			},
		},
	}
)
