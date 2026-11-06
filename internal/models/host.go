package models

import "time"

// Host represents a network host discovered during network reconnaissance.
type Host struct {
	ID        string    `json:"id"`
	TargetID  string    `json:"target_id"`
	IP        string    `json:"ip"`
	Hostname  string    `json:"hostname"`
	OS        string    `json:"os"`
	MAC       string    `json:"mac"`
	Status    string    `json:"status"` // up, down
	CreatedAt time.Time `json:"created_at"`
}
