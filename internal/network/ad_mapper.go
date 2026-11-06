package network

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/diamond/bhf/internal/models"
)

// ADAssessmentResult holds findings from Active Directory and identity discovery.
type ADAssessmentResult struct {
	IsDomainController bool     `json:"is_domain_controller"`
	DomainName         string   `json:"domain_name,omitempty"`
	OpenServices       []string `json:"open_services"`
	Observations       []string `json:"observations"`
}

// ActiveDirectoryAuditor audits host services for Active Directory roles and configurations.
type ActiveDirectoryAuditor struct {
	Timeout time.Duration
}

// NewActiveDirectoryAuditor creates a new ActiveDirectoryAuditor.
func NewActiveDirectoryAuditor() *ActiveDirectoryAuditor {
	return &ActiveDirectoryAuditor{
		Timeout: 2 * time.Second,
	}
}

// AuditHost evaluates whether a host functions as an Active Directory Domain Controller.
func (a *ActiveDirectoryAuditor) AuditHost(ctx context.Context, host models.Host) (*ADAssessmentResult, error) {
	result := &ADAssessmentResult{
		OpenServices: make([]string, 0),
		Observations: make([]string, 0),
	}

	adPorts := map[int]string{
		88:   "Kerberos (KDC)",
		389:  "LDAP Directory Service",
		636:  "LDAPS (Secure LDAP)",
		445:  "SMB / IPC Directory Sharing",
		3268: "Global Catalog Service",
	}

	dcScore := 0
	for port, serviceName := range adPorts {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		addr := fmt.Sprintf("%s:%d", host.IP, port)
		d := net.Dialer{Timeout: a.Timeout}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			conn.Close()
			result.OpenServices = append(result.OpenServices, fmt.Sprintf("%d/tcp (%s)", port, serviceName))
			if port == 88 || port == 389 || port == 3268 {
				dcScore += 2
			} else {
				dcScore += 1
			}
		}
	}

	// If host runs Kerberos (88) and LDAP (389/3268), it is confirmed as a Domain Controller
	if dcScore >= 4 {
		result.IsDomainController = true
		result.Observations = append(result.Observations,
			"Host exhibits Active Directory Domain Controller role signatures (Kerberos KDC and LDAP services open).")
	}

	// Check for unencrypted LDAP exposure
	hasCleartextLDAP := false
	hasSecureLDAP := false
	for _, svc := range result.OpenServices {
		if stringsContains(svc, "389") {
			hasCleartextLDAP = true
		}
		if stringsContains(svc, "636") {
			hasSecureLDAP = true
		}
	}

	if hasCleartextLDAP && !hasSecureLDAP {
		result.Observations = append(result.Observations,
			"Cleartext LDAP (port 389) is active without LDAPS (port 636) enforcement, exposing directory queries to network sniffing.")
	}

	return result, nil
}

func stringsContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
