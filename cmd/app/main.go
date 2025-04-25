package main

import (
	"fmt"
	"gameslabor/internal/ai"
	"gameslabor/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	defer ai.Cleanup()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	s := server.NewServer()
	defer s.Close()
	go func() {
		err := s.Start()
		if err != nil {
			fmt.Println("error while http serving:", err.Error())
		}
	}()

	<-closeChan
}
