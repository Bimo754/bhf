package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/models"
)

func TestHostAndServiceRepos(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "network_test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	defer db.Close()

	if err := MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := NewTargetRepo(db)
	hostRepo := NewHostRepo(db)
	serviceRepo := NewServiceRepo(db)

	target, err := targetRepo.Create("Network Corp", "10.0.0.0/24", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// 1. Upsert Host
	host1 := &models.Host{
		ID:        "host-1",
		TargetID:  target.ID,
		IP:        "10.0.0.10",
		Hostname:  "dc01.corp.local",
		OS:        "Windows Server 2022",
		Status:    "up",
		CreatedAt: time.Now().UTC(),
	}
	if err := hostRepo.Upsert(host1); err != nil {
		t.Fatalf("failed to upsert host: %v", err)
	}

	// 2. Query Host by ID and IP
	fetchedHost, err := hostRepo.GetByID("host-1")
	if err != nil || fetchedHost == nil {
		t.Fatalf("failed to get host by id: %v", err)
	}
	if fetchedHost.IP != "10.0.0.10" || fetchedHost.Hostname != "dc01.corp.local" {
		t.Errorf("host attributes mismatch: %+v", fetchedHost)
	}

	byIP, err := hostRepo.GetByIP(target.ID, "10.0.0.10")
	if err != nil || byIP == nil {
		t.Fatalf("failed to get host by ip: %v", err)
	}

	// 3. Upsert Services
	svc1 := &models.NetworkService{
		ID:        "svc-1",
		HostID:    host1.ID,
		TargetID:  target.ID,
		Port:      88,
		Protocol:  "tcp",
		Name:      "kerberos",
		Product:   "Microsoft Windows Kerberos",
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}
	svc2 := &models.NetworkService{
		ID:        "svc-2",
		HostID:    host1.ID,
		TargetID:  target.ID,
		Port:      443,
		Protocol:  "tcp",
		Name:      "https",
		Product:   "Microsoft-HTTPAPI/2.0",
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}
	if err := serviceRepo.Upsert(svc1); err != nil {
		t.Fatalf("failed to upsert svc1: %v", err)
	}
	if err := serviceRepo.Upsert(svc2); err != nil {
		t.Fatalf("failed to upsert svc2: %v", err)
	}

	// 4. List Services for Host
	hostServices, err := serviceRepo.ListByHost(host1.ID)
	if err != nil {
		t.Fatalf("failed to list services for host: %v", err)
	}
	if len(hostServices) != 2 {
		t.Errorf("expected 2 services on host, got %d", len(hostServices))
	}

	// 5. Automated Web Service Handoff Query
	webServices, err := serviceRepo.GetHTTPWebServices(target.ID)
	if err != nil {
		t.Fatalf("failed to query web services: %v", err)
	}
	if len(webServices) != 1 {
		t.Fatalf("expected 1 web service (port 443), got %d", len(webServices))
	}
	if webServices[0].Port != 443 {
		t.Errorf("expected web service port 443, got %d", webServices[0].Port)
	}

	// 6. Cascading Deletion
	if err := targetRepo.Delete(target.ID); err != nil {
		t.Fatalf("failed to delete target: %v", err)
	}
	hostsAfter, totalHosts, _ := hostRepo.ListByTarget(target.ID, 10, 0)
	if totalHosts != 0 || len(hostsAfter) != 0 {
		t.Errorf("expected 0 hosts after cascade delete, got %d", totalHosts)
	}
}
