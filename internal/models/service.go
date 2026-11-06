package models

import "time"

// NetworkService represents an open network port and associated protocol service.
type NetworkService struct {
	ID        string    `json:"id"`
	HostID    string    `json:"host_id"`
	TargetID  string    `json:"target_id"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"` // tcp, udp
	Name      string    `json:"name"`     // http, https, ssh, smb, rdp, ldap, kerberos, etc.
	Product   string    `json:"product"`
	Version   string    `json:"version"`
	Banner    string    `json:"banner"`
	Status    string    `json:"status"` // open, filtered, closed
	CreatedAt time.Time `json:"created_at"`
}
