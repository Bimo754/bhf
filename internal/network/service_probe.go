package network

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// ProbeService probes an open port to fingerprint the running protocol, service name, and banner.
func ProbeService(ctx context.Context, hostID, targetID, ip string, port int, timeout time.Duration) (*models.NetworkService, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	serviceName := inferDefaultService(port)
	product := ""
	version := ""
	banner := ""

	// Probe standard text/SSH banner first (listening services send banner on connect)
	if port == 22 || strings.Contains(serviceName, "ssh") {
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSH-") {
			serviceName = "ssh"
			banner = line
			parts := strings.Split(line, "-")
			if len(parts) >= 3 {
				product = parts[2]
			}
		}
	} else if port == 80 || port == 443 || port == 8080 || port == 8443 || port == 8000 || port == 3000 {
		// Send HTTP probe
		_, _ = conn.Write([]byte("HEAD / HTTP/1.0\r\nHost: " + ip + "\r\nUser-Agent: BHF-NetworkProbe/1.0\r\n\r\n"))
		reader := bufio.NewReader(conn)
		statusLine, _ := reader.ReadString('\n')
		if strings.HasPrefix(statusLine, "HTTP/") {
			serviceName = "http"
			if port == 443 || port == 8443 {
				serviceName = "https"
			}
			banner = strings.TrimSpace(statusLine)
			// Read headers for Server header
			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil || headerLine == "\r\n" || headerLine == "\n" || headerLine == "" {
					break
				}
				if strings.HasPrefix(strings.ToLower(headerLine), "server:") {
					product = strings.TrimSpace(strings.TrimPrefix(headerLine, "Server:"))
				}
			}
		}
	} else if port == 6379 {
		// Redis probe: send PING
		_, _ = conn.Write([]byte("PING\r\n"))
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		if strings.HasPrefix(line, "+PONG") || strings.Contains(line, "-NOAUTH") {
			serviceName = "redis"
			product = "Redis Key-Value Store"
			banner = strings.TrimSpace(line)
		}
	}

	return &models.NetworkService{
		ID:        uuid.New().String(),
		HostID:    hostID,
		TargetID:  targetID,
		Port:      port,
		Protocol:  "tcp",
		Name:      serviceName,
		Product:   product,
		Version:   version,
		Banner:    banner,
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}, nil
}

// HandoffWebServices bridges network discovery to Layer B web assessment by creating Asset entities.
func HandoffWebServices(ctx context.Context, services []models.NetworkService, assetRepo *database.AssetRepo, scanID string) error {
	if assetRepo == nil {
		return nil
	}

	for _, svc := range services {
		isWeb := strings.ToLower(svc.Name) == "http" || strings.ToLower(svc.Name) == "https" ||
			svc.Port == 80 || svc.Port == 443 || svc.Port == 8080 || svc.Port == 8443 ||
			svc.Port == 8000 || svc.Port == 9000 || svc.Port == 3000 || svc.Port == 5000

		if !isWeb {
			continue
		}

		scheme := "http"
		if svc.Port == 443 || svc.Port == 8443 || strings.ToLower(svc.Name) == "https" {
			scheme = "https"
		}

		hostHeader := svc.HostID // HostID contains IP or host address
		url := fmt.Sprintf("%s://%s:%d", scheme, hostHeader, svc.Port)
		if (scheme == "http" && svc.Port == 80) || (scheme == "https" && svc.Port == 443) {
			url = fmt.Sprintf("%s://%s", scheme, hostHeader)
		}

		asset := &models.Asset{
			ID:            uuid.New().String(),
			TargetID:      svc.TargetID,
			ScanID:        scanID,
			Host:          hostHeader,
			URL:           url,
			Port:          svc.Port,
			PriorityScore: 20, // Baseline for network discovered web services
			CreatedAt:     time.Now().UTC(),
		}

		_ = assetRepo.Upsert(asset)
	}

	return nil
}

func inferDefaultService(port int) string {
	switch port {
	case 21:
		return "ftp"
	case 22:
		return "ssh"
	case 23:
		return "telnet"
	case 25:
		return "smtp"
	case 53:
		return "dns"
	case 80, 8080, 8000, 3000:
		return "http"
	case 88:
		return "kerberos"
	case 389:
		return "ldap"
	case 443, 8443:
		return "https"
	case 445:
		return "smb"
	case 636:
		return "ldaps"
	case 3306:
		return "mysql"
	case 3389:
		return "rdp"
	case 5432:
		return "postgresql"
	case 6379:
		return "redis"
	case 9200:
		return "elasticsearch"
	case 27017:
		return "mongodb"
	default:
		return "unknown"
	}
}
