package common

import "github.com/google/uuid"

type Proc struct {
	ExePath string
	HashHex string
	PID     uint32
	// More Metadata
}

type ClientInfo struct {
	ID      uuid.UUID
	Running []Proc
}
