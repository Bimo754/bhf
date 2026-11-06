package network

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// ParseCIDR expands a CIDR notation into a slice of IP strings, bounded by maxHosts to prevent memory exhaustion.
func ParseCIDR(cidr string, maxHosts int) ([]string, error) {
	if maxHosts <= 0 {
		maxHosts = 1024
	}

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Test if it's a single standalone IP
		if singleIP := net.ParseIP(cidr); singleIP != nil {
			return []string{singleIP.String()}, nil
		}
		return nil, fmt.Errorf("invalid CIDR or IP address: %w", err)
	}

	ips := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
		if len(ips) >= maxHosts {
			break
		}
	}

	// Remove network address and broadcast address if subnet >= /30
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ProbeHost checks if a host is responsive on common network service ports (80, 443, 22, 445).
func ProbeHost(ctx context.Context, ip string, timeout time.Duration) bool {
	probePorts := []int{80, 443, 22, 445}
	for _, port := range probePorts {
		addr := fmt.Sprintf("%s:%d", ip, port)
		d := net.Dialer{Timeout: timeout}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

// DiscoverSubnet sweeps a CIDR range and registers responsive hosts into HostRepo.
func DiscoverSubnet(ctx context.Context, cidr string, hostRepo *database.HostRepo, targetID string) ([]models.Host, error) {
	ips, err := ParseCIDR(cidr, 1024)
	if err != nil {
		return nil, err
	}

	discovered := make([]models.Host, 0)
	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return discovered, ctx.Err()
		default:
		}

		if ProbeHost(ctx, ip, 500*time.Millisecond) {
			hostname := ""
			if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
				hostname = names[0]
			}

			host := models.Host{
				ID:        uuid.New().String(),
				TargetID:  targetID,
				IP:        ip,
				Hostname:  hostname,
				Status:    "up",
				CreatedAt: time.Now().UTC(),
			}

			if hostRepo != nil {
				_ = hostRepo.Upsert(&host)
			}
			discovered = append(discovered, host)
		}
	}

	return discovered, nil
}
