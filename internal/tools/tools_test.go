package tools_test

import (
	"context"
	"testing"

	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
)

func newTestContext(domain string) tools.ExecutionContext {
	target := &models.Target{
		ID:     "test-target-id",
		Name:   "Test Target",
		Domain: domain,
	}
	scan := &models.Scan{
		ID:       "test-scan-id",
		TargetID: target.ID,
		ToolName: "test-tool",
	}
	return tools.ExecutionContext{
		Ctx:    context.Background(),
		Scan:   scan,
		Target: target,
	}
}

func TestSubfinderParser(t *testing.T) {
	tool := tools.NewSubfinderTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[INF] Starting subfinder",
		"sub1.target.com",
		"sub2.target.com",
		"{\"host\":\"json.target.com\"}",
		"out-of-scope.other.com",
		"[INF] Done",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Subfinder Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(result.Assets))
	}

	expectedHosts := map[string]bool{
		"sub1.target.com": false,
		"sub2.target.com": false,
		"json.target.com": false,
	}

	for _, a := range result.Assets {
		if _, ok := expectedHosts[a.Host]; ok {
			expectedHosts[a.Host] = true
		}
	}

	for host, found := range expectedHosts {
		if !found {
			t.Errorf("expected host %s to be parsed", host)
		}
	}
}

func TestKatanaParser(t *testing.T) {
	tool := tools.NewKatanaTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"https://target.com/api/v1/users",
		"https://sub.target.com/login",
		"https://sub.target.com/login",        // Duplicate
		"https://thirdparty.com/analytics.js", // Out of scope
		"not-a-url",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Katana Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 distinct in-scope assets, got %d", len(result.Assets))
	}
}

func TestHttpxParser(t *testing.T) {
	tool := tools.NewHttpxTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"https://admin.target.com [200] [Admin Portal] [Vue.js, Nginx]",
		"https://api.target.com [404] [Not Found]",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Httpx Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 assets from raw output fallback, got %d", len(result.Assets))
	}
}

func TestNaabuParser(t *testing.T) {
	tool := tools.NewNaabuTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"sub.target.com:80",
		"sub.target.com:443",
		"sub.target.com:8443",
		"invalid-format",
		"foreign.com:80", // out of scope
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Naabu Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 3 {
		t.Fatalf("expected 3 in-scope port assets, got %d", len(result.Assets))
	}
}

func TestDNSxParser(t *testing.T) {
	tool := tools.NewDNSxTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"sub.target.com [192.168.1.1]",
		"api.target.com [192.168.1.2]",
		"other.org [10.0.0.1]", // out of scope
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("DNSx Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 in-scope assets, got %d", len(result.Assets))
	}

	if result.Assets[0].IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", result.Assets[0].IP)
	}
}

func TestAlterxParser(t *testing.T) {
	tool := tools.NewAlterxTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[INF] Generating permutations",
		"dev-api.target.com",
		"staging-api.target.com",
		"dev-api.target.com", // Duplicate
		"out-of-scope.attacker.com",
		"",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Alterx Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 distinct in-scope permutation assets, got %d", len(result.Assets))
	}
}

func TestSubzyParser(t *testing.T) {
	tool := tools.NewSubzyTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[INFO] Checking 3 targets",
		"[VULNERABLE] sub1.target.com [AWS/S3]",
		"[NOT VULNERABLE] sub2.target.com",
		"[VULNERABLE] evil.com [GitHub]", // out of scope
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Subzy Run returned unexpected error: %v", err)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 in-scope takeover finding, got %d", len(result.Findings))
	}

	if result.Findings[0].MatchedAt != "sub1.target.com" {
		t.Errorf("expected matched_at sub1.target.com, got %s", result.Findings[0].MatchedAt)
	}

	if result.Findings[0].Severity != models.SeverityHigh {
		t.Errorf("expected High severity, got %v", result.Findings[0].Severity)
	}
}

func TestAsnmapParser(t *testing.T) {
	tool := tools.NewAsnmapTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		`{"first_ip":"104.16.0.0","last_ip":"104.31.255.255","asn":13335,"as_name":"CLOUDFLARENET","as_country":"US","cidr":"104.16.0.0/12"}`,
		"192.168.1.0/24",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Asnmap Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 network allocation assets, got %d", len(result.Assets))
	}

	if result.Assets[0].IP != "104.16.0.0/12" {
		t.Errorf("expected CIDR 104.16.0.0/12, got %s", result.Assets[0].IP)
	}
}

func TestTlsxParser(t *testing.T) {
	tool := tools.NewTlsxTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		`{"host":"target.com:443","ip":"104.16.1.1","port":"443","subject_an":["target.com","dev.target.com","vpn.target.com","out-of-scope.net"],"tls_version":"tls13","jarm":"29d29d00029d"}`,
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Tlsx Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 3 {
		t.Fatalf("expected 3 in-scope assets (root, dev, vpn), got %d", len(result.Assets))
	}
}

func TestWafw00fParser(t *testing.T) {
	tool := tools.NewWafw00fTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"The site https://api.target.com is behind Cloudflare (Cloudflare Inc.) WAF.",
		"No WAF detected by the generic detection",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Wafw00f Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 1 {
		t.Fatalf("expected 1 WAF-detected asset, got %d", len(result.Assets))
	}

	if len(result.Assets[0].Technologies) == 0 || result.Assets[0].Technologies[0] != "WAF:Cloudflare" {
		t.Errorf("expected technology WAF:Cloudflare, got %v", result.Assets[0].Technologies)
	}
}

func TestUroParser(t *testing.T) {
	tool := tools.NewUroTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"https://target.com/api/v1/users?id=1",
		"https://sub.target.com/login",
		"https://out-of-scope.com/test",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Uro Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 in-scope deduplicated assets, got %d", len(result.Assets))
	}
}

func TestJsluiceParser(t *testing.T) {
	tool := tools.NewJsluiceTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		`{"url":"/api/v2/internal/auth","type":"relative","filename":"https://target.com/main.js"}`,
		`{"kind":"aws-access-key","match":"AKIAIOSFODNN7EXAMPLE","filename":"https://target.com/main.js","severity":"high"}`,
		`{"url":"https://evil.com/leak","type":"absolute"}`, // out of scope
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Jsluice Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 1 {
		t.Fatalf("expected 1 in-scope route asset, got %d", len(result.Assets))
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 secret finding, got %d", len(result.Findings))
	}

	if result.Findings[0].Severity != models.SeverityCritical {
		t.Errorf("expected Critical severity for AWS key, got %v", result.Findings[0].Severity)
	}
}

func TestArjunParser(t *testing.T) {
	tool := tools.NewArjunTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[+] Parameters found for https://target.com/api/users: id, debug, token",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Arjun Run returned unexpected error: %v", err)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding for discovered parameters, got %d", len(result.Findings))
	}

	if len(result.Findings[0].ExtractedResults) != 3 {
		t.Fatalf("expected 3 extracted parameters, got %d", len(result.Findings[0].ExtractedResults))
	}
}

func TestNomore403Parser(t *testing.T) {
	tool := tools.NewNomore403Tool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[200 OK] https://target.com/admin/ (Header: X-Forwarded-For: 127.0.0.1)",
		"[403 Forbidden] https://target.com/admin/login",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Nomore403 Run returned unexpected error: %v", err)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 bypass finding, got %d", len(result.Findings))
	}

	if result.Findings[0].Severity != models.SeverityHigh {
		t.Errorf("expected High severity, got %v", result.Findings[0].Severity)
	}
}

func TestCorsyParser(t *testing.T) {
	tool := tools.NewCorsyTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		`[!] CORS Misconfiguration on https://api.target.com/profile with ACAC: true`,
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Corsy Run returned unexpected error: %v", err)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 CORS finding, got %d", len(result.Findings))
	}

	if result.Findings[0].Severity != models.SeverityHigh {
		t.Errorf("expected High severity for ACAC true, got %v", result.Findings[0].Severity)
	}
}

func TestS3ScannerParser(t *testing.T) {
	tool := tools.NewS3ScannerTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"target-backup | exists | read | read",
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("S3Scanner Run returned unexpected error: %v", err)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 bucket finding, got %d", len(result.Findings))
	}

	if result.Findings[0].Severity != models.SeverityHigh {
		t.Errorf("expected High severity for readable bucket, got %v", result.Findings[0].Severity)
	}
}

func TestGowitnessParser(t *testing.T) {
	tool := tools.NewGowitnessTool()
	execCtx := newTestContext("target.com")

	rawOutput := []string{
		"[INF] Pre-flight checks complete",
		"[INF] Screenshot taken for https://api.target.com",
		"[INF] Screenshot taken for https://admin.target.com/login",
		"[INF] Screenshot failed for https://evil.com", // out of scope
	}

	result, err := tool.Run(execCtx, rawOutput)
	if err != nil {
		t.Fatalf("Gowitness Run returned unexpected error: %v", err)
	}

	if len(result.Assets) != 2 {
		t.Fatalf("expected 2 screenshot assets, got %d", len(result.Assets))
	}

	if result.Assets[0].Title != "Visual Screenshot Captured" {
		t.Errorf("expected Title 'Visual Screenshot Captured', got %s", result.Assets[0].Title)
	}
}
