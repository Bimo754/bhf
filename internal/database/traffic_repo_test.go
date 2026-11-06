package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

func TestTrafficRepo(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "traffic_test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	defer db.Close()

	if err := MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := NewTargetRepo(db)
	trafficRepo := NewTrafficRepo(db)

	target, err := targetRepo.Create("Traffic Test Target", "traffic.example.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// 1. Insert traffic records
	rec1 := &models.TrafficRecord{
		ID:                 uuid.New().String(),
		TargetID:           target.ID,
		Method:             "POST",
		URL:                "https://traffic.example.com/api/v1/auth/login",
		RequestHeaders:     `{"Content-Type":["application/json"]}`,
		RequestBody:        `{"username":"admin"}`,
		ResponseStatusCode: 200,
		ResponseHeaders:    `{"Set-Cookie":["session=abc"]}`,
		ResponseBody:       `{"status":"ok"}`,
		ContentType:        "application/json",
		ContentLength:      15,
		DurationMs:         45,
		Category:           "auth",
		AuthContext:        "unauthenticated",
		Parameters:         `[{"name":"username","type":"body"}]`,
		CreatedAt:          time.Now().UTC(),
	}

	rec2 := &models.TrafficRecord{
		ID:                 uuid.New().String(),
		TargetID:           target.ID,
		Method:             "GET",
		URL:                "https://traffic.example.com/api/v1/users?search=test",
		RequestHeaders:     `{"Accept":["application/json"]}`,
		RequestBody:        "",
		ResponseStatusCode: 200,
		ResponseHeaders:    `{"Content-Type":["application/json"]}`,
		ResponseBody:       `[{"id":1}]`,
		ContentType:        "application/json",
		ContentLength:      10,
		DurationMs:         20,
		Category:           "api",
		AuthContext:        "session_cookie",
		Parameters:         `[{"name":"search","type":"query"}]`,
		CreatedAt:          time.Now().UTC(),
	}

	rec3 := &models.TrafficRecord{
		ID:                 uuid.New().String(),
		TargetID:           target.ID,
		Method:             "POST",
		URL:                "https://traffic.example.com/upload/avatar",
		RequestHeaders:     `{"Content-Type":["multipart/form-data"]}`,
		RequestBody:        "BINARY_DATA",
		ResponseStatusCode: 201,
		ResponseHeaders:    `{"Location":["/avatars/1.png"]}`,
		ResponseBody:       `{"url":"/avatars/1.png"}`,
		ContentType:        "application/json",
		ContentLength:      24,
		DurationMs:         120,
		Category:           "upload",
		AuthContext:        "session_cookie",
		Parameters:         `[{"name":"file","type":"file"}]`,
		CreatedAt:          time.Now().UTC(),
	}

	for _, rec := range []*models.TrafficRecord{rec1, rec2, rec3} {
		if err := trafficRepo.Insert(rec); err != nil {
			t.Fatalf("failed to insert traffic record %s: %v", rec.ID, err)
		}
	}

	// 2. Get by ID
	fetched, err := trafficRepo.GetByID(rec1.ID)
	if err != nil {
		t.Fatalf("failed to get by ID: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected fetched record to not be nil")
	}
	if fetched.URL != rec1.URL || fetched.Category != "auth" || fetched.Host != "traffic.example.com" {
		t.Errorf("fetched record mismatch: got %+v", fetched)
	}

	// 3. Count
	count, err := trafficRepo.CountByTarget(target.ID)
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	// 4. List with category filter
	authList, authTotal, err := trafficRepo.ListByTarget(target.ID, "auth", "", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list by category: %v", err)
	}
	if authTotal != 1 || len(authList) != 1 {
		t.Errorf("expected 1 auth record, got total %d, len %d", authTotal, len(authList))
	}

	// 5. List with method filter
	postList, postTotal, err := trafficRepo.ListByTarget(target.ID, "", "POST", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list by method: %v", err)
	}
	if postTotal != 2 || len(postList) != 2 {
		t.Errorf("expected 2 POST records, got total %d, len %d", postTotal, len(postList))
	}

	// 6. List with search filter
	searchList, searchTotal, err := trafficRepo.ListByTarget(target.ID, "", "", "avatar", 10, 0)
	if err != nil {
		t.Fatalf("failed to list by search: %v", err)
	}
	if searchTotal != 1 || len(searchList) != 1 {
		t.Errorf("expected 1 search result, got total %d, len %d", searchTotal, len(searchList))
	}

	// 7. Category Stats
	stats, err := trafficRepo.GetCategoryStats(target.ID)
	if err != nil {
		t.Fatalf("failed to get category stats: %v", err)
	}
	if stats["auth"] != 1 || stats["api"] != 1 || stats["upload"] != 1 {
		t.Errorf("unexpected category stats: %+v", stats)
	}

	// 8. Cascading Deletion
	if err := targetRepo.Delete(target.ID); err != nil {
		t.Fatalf("failed to delete target: %v", err)
	}
	postDeleteCount, err := trafficRepo.CountByTarget(target.ID)
	if err != nil {
		t.Fatalf("failed to count after target deletion: %v", err)
	}
	if postDeleteCount != 0 {
		t.Errorf("expected 0 records after cascade delete, got %d", postDeleteCount)
	}
}
