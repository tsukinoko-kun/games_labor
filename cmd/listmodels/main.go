package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gameslabor/internal/env"
	"os"
	"strings"

	"google.golang.org/genai"
)

var (
	filters []string
)

func main() {
	if flag.NArg() > 0 {
		filters = flag.Args()
	}

	ctx := context.Background()
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  env.GOOGLE_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}

	mp, err := c.Models.List(ctx, nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
	for {
	items:
		for _, m := range mp.Items {
			for _, filter := range filters {
				if !strings.Contains(m.Name, filter) && !strings.Contains(m.DisplayName, filter) {
					continue items
				}
			}
			fmt.Printf("%s (%s)\n", m.Name, m.DisplayName)
		}
		mp, err = mp.Next(ctx)
		if err != nil {
			if errors.Is(err, genai.ErrPageDone) {
				break
			}
			fmt.Println(err.Error())
			os.Exit(1)
			return
		}
	}
}
