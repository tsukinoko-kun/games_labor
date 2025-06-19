package server

import (
	"fmt"
	"gameslabor/internal/ai"
	"gameslabor/internal/env"
	"gameslabor/internal/server/api"
	"gameslabor/internal/server/pages"
	"gameslabor/internal/server/public"
	"net"
	"net/http"
	"strconv"
)

type Server struct {
	ai         *ai.AI
	mux        *http.ServeMux
	httpServer *http.Server
}

func NewServer() *Server {
	mux := http.DefaultServeMux
	mux.HandleFunc("/public/", public.Handler)
	mux.HandleFunc("/ai/", ai.Handler)
	mux.HandleFunc("/api/", api.Handler)
	mux.HandleFunc("/", pages.Handler)
	return &Server{
		mux: mux,
	}
}

func (s *Server) Start() error {
	addr := ":" + strconv.Itoa(env.PORT)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	if ip, err := getLocalIP(); err == nil {
		fmt.Printf("http://%s:%d\n", ip, env.PORT)
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Close() {
	if s.httpServer != nil {
		_ = s.httpServer.Close()
		s.httpServer = nil
	}
	if s.ai != nil {
		s.ai.Close()
		s.ai = nil
	}
}

// getLocalIP retrieves the local IP address of the computer.
func getLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no connected network interface found")
}
