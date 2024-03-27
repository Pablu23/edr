package server

import (
	"bufio"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
)

type ProcessInfo struct {
	Name string
	Hash string
}

// OsInfo Placeholder
type OsInfo struct {
	OS      string // win / linux / osx / etc..
	Version string
	x64     bool
}

// PcInfo Placeholder
type PcInfo struct {
	GPU    string
	CPU    string
	Memory string
	// All Pc Details
}

type ClientProcess struct {
	Proc ProcessInfo
	PIDs []uint32
}

type Client struct {
	Id               uuid.UUID
	Ip               string
	Os               OsInfo
	Pc               PcInfo
	Users            []string
	RunningProcesses []ClientProcess
	Connected        bool
}

type Server struct {
	Clients          []Client
	AllowedProcesses []ProcessInfo
	router           *http.ServeMux
	logger           *zap.SugaredLogger
}

func (s *Server) ListenAndServe() {
	server := http.Server{
		Addr:    ":8080",
		Handler: s.router,
	}
	s.logger.Infow("Starting server", "Addr", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("Could not start Server")
	}
}

func (s *Server) addRoutes() {
	s.logger.Debug("Adding routes")
	s.router.HandleFunc("GET /allowed/", s.getAllowedProcesses)
	s.router.HandleFunc("POST /client/", s.postClientInfo)
	s.router.HandleFunc("GET /clients/", s.getClients)
}

func New(logger *zap.Logger) Server {
	s := Server{
		Clients:          make([]Client, 0),
		AllowedProcesses: make([]ProcessInfo, 0),
		router:           http.NewServeMux(),
		logger:           logger.Sugar(),
	}
	s.addRoutes()
	return s
}

func NewFromFile(filePath string, logger *zap.Logger) Server {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	allowed := make([]ProcessInfo, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := scanner.Text()
		s := strings.Split(l, ";")

		a := ProcessInfo{
			Name: path.Base(s[0]),
			Hash: s[1],
		}

		if !slices.Contains(allowed, a) {
			allowed = append(allowed, a)
		}
	}
	file.Close()

	s := Server{
		Clients:          make([]Client, 0),
		AllowedProcesses: allowed,
		router:           http.NewServeMux(),
		logger:           logger.Sugar(),
	}
	s.addRoutes()
	return s
}
