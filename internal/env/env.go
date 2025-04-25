package env

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	GOOGLE_API_KEY string
	PORT           int
)

func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	// read line by line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Split(line, "#")[0]
		parts := strings.Split(line, "=")
		key := strings.TrimSpace(parts[0])
		if len(parts) < 2 {
			os.Setenv(key, "")
			continue
		}

		var value string
		value = strings.TrimSpace(strings.Join(parts[1:], " "))
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error reading .env file: %v\n", err)
		return
	}
}

func init() {
	loadEnv()
	if googleAiApiKey := os.Getenv("GOOGLE_API_KEY"); googleAiApiKey != "" {
		GOOGLE_API_KEY = googleAiApiKey
	} else {
		fmt.Println("GOOGLE_API_KEY environment variable not set")
		os.Exit(1)
	}
	if port := os.Getenv("PORT"); port != "" {
		var err error
		PORT, err = strconv.Atoi(port)
		if err != nil {
			fmt.Printf("error parsing PORT environment variable: %v\n", err)
			os.Exit(1)
		}
	} else {
		PORT = 8080
	}

	flag.IntVar(&PORT, "port", PORT, "Port to listen on")
	flag.Parse()
}
