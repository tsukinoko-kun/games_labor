package main

import (
	"context"
	"fmt"
	"gameslabor/internal/llm"
)

func main() {
	ctx := context.Background()
	if llmInstalce, err := llm.New(ctx); err != nil {
		panic(err)
	} else {
		defer llmInstalce.Close()
		resp := llmInstalce.Start("Der Spieler möchte ein Action Abenteuer mit Schatzsuche-Elementen, ähnlich zu den Uncharted-Spielern. Das Setting soll im späten Mittelalter gesetzt sein mit Fantasy Elementen.")
		fmt.Println(resp.JSON())

		var prompt string
		for {
			fmt.Print("Enter: ")
			for len(prompt) == 0 {
				_, err := fmt.Scanln(&prompt)
				if err != nil {
					prompt = ""
					fmt.Println("Error reading input:", err)
					continue
				}
			}

			if prompt == ".exit" {
				break
			}

			resp := llmInstalce.Text(prompt)
			fmt.Println(resp.JSON())

			prompt = ""
		}
	}
}
