package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gameslabor/internal/llm"
	"strings"
)

func main() {
	ctx := context.Background()
	if llmInstalce, err := llm.New(ctx); err != nil {
		panic(err)
	} else {
		defer llmInstalce.Close()

		prompt := "Der Spieler möchte ein Action Abenteuer mit Schatzsuche-Elementen, ähnlich zu den Uncharted-Spielern. Das Setting soll im späten Mittelalter gesetzt sein mit Fantasy Elementen."
		for prompt != ".end" {
			resp := llmInstalce.Text(prompt)
			sb := strings.Builder{}
			je := json.NewEncoder(&sb)
			je.SetIndent("", "  ")
			je.Encode(&resp)
			fmt.Println(sb.String())

			fmt.Println("Press Enter to continue...")
			fmt.Scanln(&prompt)
		}
	}
}
