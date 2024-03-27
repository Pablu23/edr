package server

import (
	"edr/pkg/common"
	"encoding/json"
	"net/http"
	"slices"
)

func (s *Server) getAllowedProcesses(w http.ResponseWriter, _ *http.Request) {
	s.logger.Debug("GET /allowed")
	res, err := json.Marshal(s.AllowedProcesses)
	if err != nil {
		s.logger.Errorw("Could not Marshal allowed Processes to Json", "Error", err)
		w.WriteHeader(500)
		return
	}
	_, err = w.Write(res)
	if err != nil {
		s.logger.Errorw("Could not Write to response", "Error", err)
		w.WriteHeader(500)
		return
	}
}

func (s *Server) getClients(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(s.Clients)
	if err != nil {
		s.logger.Errorw("Could not Write to response", "Error", err)
		w.WriteHeader(500)
	}
}

func (s *Server) postClientInfo(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("POST /client")
	var cInfo common.ClientInfo
	err := json.NewDecoder(r.Body).Decode(&cInfo)
	if err != nil {
		s.logger.Errorw("Could not Marshal allowed Processes to Json", "Error", err)

		// Write alert for operator to see
		w.WriteHeader(400)
		// Write error json
		return
	}

	cProcs := make([]ClientProcess, 0)
	pInfos := make([]ProcessInfo, 0)
	kill := make([]uint32, 0)
	for _, proc := range cInfo.Running {
		contains := slices.ContainsFunc(s.AllowedProcesses, func(process ProcessInfo) bool {
			return process.Hash == proc.HashHex
		})

		if !slices.ContainsFunc(cProcs, func(p ClientProcess) bool {
			return p.Proc.Hash == proc.HashHex
		}) {
			cProc := ClientProcess{
				Proc: ProcessInfo{
					Name: proc.ExePath,
					Hash: proc.HashHex,
				},
				PIDs: make([]uint32, 1),
			}
			cProc.PIDs[0] = proc.PID
			cProcs = append(cProcs, cProc)
		} else {
			i := slices.IndexFunc(cProcs, func(p ClientProcess) bool {
				return p.Proc.Hash == proc.HashHex
			})
			cProcs[i].PIDs = append(cProcs[i].PIDs, proc.PID)
		}

		if !contains {
			p := ProcessInfo{
				Name: proc.ExePath,
				Hash: proc.HashHex,
			}
			pInfos = append(pInfos, p)
			kill = append(kill, proc.PID)
		}
	}

	i := slices.IndexFunc(s.Clients, func(client Client) bool {
		return client.Id == cInfo.ID
	})
	if i != -1 {
		s.Clients[i].RunningProcesses = cProcs
	} else {
		s.Clients = append(s.Clients, Client{
			Id:               cInfo.ID,
			Ip:               r.RemoteAddr,
			Os:               OsInfo{},
			Pc:               PcInfo{},
			Users:            nil,
			RunningProcesses: cProcs,
			Connected:        true,
		})
	}

	if len(kill) > 0 {
		s.logger.Infow("Forbidden Processes", "Client", cInfo.ID, "Processes", pInfos)
	}
	err = json.NewEncoder(w).Encode(kill)
	if err != nil {
		s.logger.Errorw("Could not Write to response", "Error", err)
		// alert and more
		return
	}
}
