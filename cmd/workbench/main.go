package main

import (
	"context"
	"fmt"
	"gameslabor/internal/ai"
	"os"
	"os/exec"
)

func main() {
	aiInstalce, err := ai.New(context.Background())
	if err != nil {
		panic(err)
	}
	defer aiInstalce.Close()

	defer ai.Cleanup()

	audioFileName, err := aiInstalce.TTS(`Ihr stoßt die knarrende Tür zur Spelunke "Zum salzigen Seeteufel" auf und eine Welle aus abgestandenem Rum, Schweiß und dem salzigen Geruch geteerter Taue schlägt euch entgegen.`)
	if err != nil {
		panic(err)
	}

	audioFileName = ai.FullFilename(audioFileName)
	fmt.Println(audioFileName)

	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", "-nodisp", "-autoexit", audioFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
