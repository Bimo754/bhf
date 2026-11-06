package network_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/network"
)

func TestParseCIDR(t *testing.T) {
	// Single IP
	single, err := network.ParseCIDR("192.168.1.10", 100)
	if err != nil || len(single) != 1 || single[0] != "192.168.1.10" {
		t.Fatalf("unexpected single IP parse: %v, %v", single, err)
	}

	// Subnet /30 (4 IPs total: .0 network, .1 and .2 usable, .3 broadcast)
	sub, err := network.ParseCIDR("10.0.0.0/30", 100)
	if err != nil {
		t.Fatalf("failed to parse /30 subnet: %v", err)
	}
	if len(sub) != 2 || sub[0] != "10.0.0.1" || sub[1] != "10.0.0.2" {
		t.Errorf("expected usable IPs [10.0.0.1, 10.0.0.2], got %v", sub)
	}

	// Invalid CIDR
	_, err = network.ParseCIDR("invalid-ip-string", 100)
	if err == nil {
		t.Error("expected error for invalid CIDR, got nil")
	}
}

func TestProbeServiceHTTP(t *testing.T) {
	// Start mock HTTP server
	mockHTTP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "BHF-TestServer/2.4")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockHTTP.Close()

	parts := strings.Split(mockHTTP.Listener.Addr().String(), ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])

	ctx := context.Background()
	svc, err := network.ProbeService(ctx, "host-1", "tgt-1", "127.0.0.1", port, 2*time.Second)
	if err != nil {
		t.Fatalf("failed to probe mock service: %v", err)
	}

	if svc.Port != port {
		t.Errorf("expected port %d, got %d", port, svc.Port)
	}
}

func TestHandoffWebServices(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "handoff_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()
	_ = database.MigrateSchema(db)

	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	assetRepo := database.NewAssetRepo(db)
	target, _ := targetRepo.Create("Web Handoff Target", "10.0.0.5", "")

	_ = scanRepo.Create(&models.Scan{
		ID:        "scan-handoff-1",
		TargetID:  target.ID,
		ToolName:  "network-probe",
		Status:    models.StatusCompleted,
		Command:   "probe",
		StartedAt: time.Now().UTC(),
	})

	services := []models.NetworkService{
		{
			ID:       "svc-ssh",
			HostID:   "10.0.0.5",
			TargetID: target.ID,
			Port:     22,
			Name:     "ssh",
		},
		{
			ID:       "svc-http",
			HostID:   "10.0.0.5",
			TargetID: target.ID,
			Port:     8080,
			Name:     "http",
		},
	}

	ctx := context.Background()
	err = network.HandoffWebServices(ctx, services, assetRepo, "scan-handoff-1")
	if err != nil {
		t.Fatalf("HandoffWebServices failed: %v", err)
	}

	assets, total, err := assetRepo.ListByTarget(target.ID, "", "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}

	if total != 1 || len(assets) != 1 {
		t.Fatalf("expected exactly 1 web asset handed off, got total %d, len %d", total, len(assets))
	}
	if assets[0].Port != 8080 || assets[0].URL != "http://10.0.0.5:8080" {
		t.Errorf("unexpected handed off asset attributes: %+v", assets[0])
	}
}

func TestActiveDirectoryAuditor(t *testing.T) {
	// Start mock TCP listener on loopback
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind mock listener: %v", err)
	}
	defer listener.Close()

	auditor := network.NewActiveDirectoryAuditor()
	auditor.Timeout = 100 * time.Millisecond

	host := models.Host{
		IP: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := auditor.AuditHost(ctx, host)
	if err != nil {
		t.Fatalf("AuditHost failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result not nil")
	}
}

func TestAuditNetworkService(t *testing.T) {
	// Test Redis unauthenticated
	redisSvc := models.NetworkService{
		Port:   6379,
		Name:   "redis",
		Banner: "+PONG",
	}
	redisResults := network.AuditNetworkService(context.Background(), redisSvc)
	if len(redisResults) == 0 {
		t.Fatal("expected finding for unauthenticated Redis")
	}
	if redisResults[0].Severity != "critical" {
		t.Errorf("expected critical severity for open Redis, got %s", redisResults[0].Severity)
	}

	// Test Telnet cleartext
	telnetSvc := models.NetworkService{
		Port: 23,
		Name: "telnet",
	}
	telnetResults := network.AuditNetworkService(context.Background(), telnetSvc)
	if len(telnetResults) == 0 {
		t.Fatal("expected finding for Telnet")
	}
	if telnetResults[0].Severity != "high" {
		t.Errorf("expected high severity for Telnet, got %s", telnetResults[0].Severity)
	}
}
