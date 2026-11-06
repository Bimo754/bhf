package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/network"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NetworkHandler manages network discovery, service enumeration, and web recon handoffs.
type NetworkHandler struct {
	hostRepo    *database.HostRepo
	serviceRepo *database.ServiceRepo
	assetRepo   *database.AssetRepo
	findingRepo *database.FindingRepo
	targetRepo  *database.TargetRepo
	scanRepo    *database.ScanRepo
}

// NewNetworkHandler creates a new NetworkHandler.
func NewNetworkHandler(
	hostRepo *database.HostRepo,
	serviceRepo *database.ServiceRepo,
	assetRepo *database.AssetRepo,
	findingRepo *database.FindingRepo,
	targetRepo *database.TargetRepo,
	scanRepo *database.ScanRepo,
) *NetworkHandler {
	return &NetworkHandler{
		hostRepo:    hostRepo,
		serviceRepo: serviceRepo,
		assetRepo:   assetRepo,
		findingRepo: findingRepo,
		targetRepo:  targetRepo,
		scanRepo:    scanRepo,
	}
}

// ListHosts returns all network hosts discovered for a target.
func (h *NetworkHandler) ListHosts(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	hosts, _, err := h.hostRepo.ListByTarget(targetID, 100, 0)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if hosts == nil {
		hosts = []models.Host{}
	}
	JSON(w, http.StatusOK, hosts)
}

// ListServices returns all network services discovered for a target.
func (h *NetworkHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	services, _, err := h.serviceRepo.ListByTarget(targetID, 100, 0)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if services == nil {
		services = []models.NetworkService{}
	}
	JSON(w, http.StatusOK, services)
}

// ScanRequest parameters for network discovery.
type ScanRequest struct {
	CIDR  string `json:"cidr"`
	Ports []int  `json:"ports"`
}

// ScanSubnet executes host discovery, port probing, and web service handoff.
func (h *NetworkHandler) ScanSubnet(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	target, err := h.targetRepo.GetByID(targetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	var req ScanRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	cidr := req.CIDR
	if cidr == "" {
		// Resolve target domain to IP and scan single host (/32)
		ips, err := net.LookupHost(target.Domain)
		if err != nil || len(ips) == 0 {
			ErrorJSON(w, http.StatusBadRequest, "no valid CIDR provided and target domain could not be resolved: "+err.Error())
			return
		}
		cidr = ips[0] + "/32"
	}

	ipList, err := network.ParseCIDR(cidr, 256)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid CIDR: "+err.Error())
		return
	}

	ports := req.Ports
	if len(ports) == 0 {
		ports = []int{21, 22, 80, 443, 389, 3389, 6379, 8080, 8443, 9200, 27017}
	}

	now := time.Now()
	scan := &models.Scan{
		ID:        uuid.New().String(),
		TargetID:  targetID,
		ToolName:  "network_discovery",
		Status:    models.StatusRunning,
		StartedAt: now,
		Arguments: []string{cidr},
	}
	if err := h.scanRepo.Create(scan); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "failed to initialize scan: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	discoveredHosts := make([]models.Host, 0)
	discoveredServices := make([]models.NetworkService, 0)
	adAuditor := network.NewActiveDirectoryAuditor()

	for _, ip := range ipList {
		if !network.ProbeHost(ctx, ip, 500*time.Millisecond) && len(ipList) > 1 {
			continue
		}

		hostModel := &models.Host{
			TargetID: targetID,
			IP:       ip,
			Status:   "up",
		}
		if err := h.hostRepo.Upsert(hostModel); err != nil {
			continue
		}
		discoveredHosts = append(discoveredHosts, *hostModel)

		// Probe requested ports on live host
		for _, port := range ports {
			svc, err := network.ProbeService(ctx, hostModel.ID, targetID, ip, port, 500*time.Millisecond)
			if err == nil && svc != nil && svc.Status == "open" {
				_ = h.serviceRepo.Upsert(svc)
				discoveredServices = append(discoveredServices, *svc)
				_ = network.AuditNetworkService(ctx, *svc)
			}
		}

		// Active Directory audit
		_, _ = adAuditor.AuditHost(ctx, *hostModel)
	}

	// Automatic Web Reconnaissance Handoff
	_ = network.HandoffWebServices(ctx, discoveredServices, h.assetRepo, scan.ID)
	completedAt := time.Now()
	_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusCompleted, &completedAt, "")

	JSON(w, http.StatusOK, map[string]any{
		"target_id":        targetID,
		"scan_id":          scan.ID,
		"hosts_discovered": len(discoveredHosts),
		"services_open":    len(discoveredServices),
		"hosts":            discoveredHosts,
		"services":         discoveredServices,
	})
}

// ADAuditResultWithHost wraps the ADAssessmentResult with the host IP.
type ADAuditResultWithHost struct {
	HostIP string `json:"host_ip"`
	network.ADAssessmentResult
}

// AuditActiveDirectory evaluates all hosts for target or specific host for Active Directory configuration.
func (h *NetworkHandler) AuditActiveDirectory(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "targetID")
	target, err := h.targetRepo.GetByID(targetID)
	if err != nil || target == nil {
		ErrorJSON(w, http.StatusNotFound, "target not found")
		return
	}

	var req struct {
		HostIP string `json:"host_ip"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	hosts, _, err := h.hostRepo.ListByTarget(targetID, 100, 0)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If no hosts in DB, check if target domain itself resolves to an IP
	if len(hosts) == 0 {
		ips, err := net.LookupHost(target.Domain)
		if err == nil && len(ips) > 0 {
			hostModel := &models.Host{
				TargetID: targetID,
				IP:       ips[0],
				Status:   "up",
			}
			_ = h.hostRepo.Upsert(hostModel)
			hosts = append(hosts, *hostModel)
		}
	}

	adAuditor := network.NewActiveDirectoryAuditor()
	results := make([]ADAuditResultWithHost, 0)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var scan *models.Scan
	if h.scanRepo != nil {
		now := time.Now().UTC()
		scan = &models.Scan{
			ID:        uuid.New().String(),
			TargetID:  targetID,
			ToolName:  "active_directory_auditor",
			Status:    models.StatusRunning,
			StartedAt: now,
			Arguments: []string{},
		}
		_ = h.scanRepo.Create(scan)
	}

	for _, host := range hosts {
		if req.HostIP != "" && host.IP != req.HostIP {
			continue
		}

		res, err := adAuditor.AuditHost(ctx, host)
		if err == nil && res != nil {
			results = append(results, ADAuditResultWithHost{
				HostIP:             host.IP,
				ADAssessmentResult: *res,
			})

			// If unencrypted LDAP found or domain controller, persist to findings
			if h.findingRepo != nil && scan != nil {
				for _, obs := range res.Observations {
					sev := models.SeverityMedium
					if strings.Contains(obs, "Unencrypted") {
						sev = models.SeverityHigh
					}
					finding := &models.Finding{
						ID:          uuid.New().String(),
						TargetID:    targetID,
						ScanID:      scan.ID,
						ToolName:    "active_directory_auditor",
						Severity:    sev,
						Name:        "Active Directory: " + obs,
						Description: fmt.Sprintf("Active Directory audit on host %s identified: %s", host.IP, obs),
						MatchedAt:   fmt.Sprintf("ldap://%s:389", host.IP),
						CurlCommand: fmt.Sprintf("ldapsearch -x -H ldap://%s -s base", host.IP),
						CreatedAt:   time.Now().UTC(),
					}
					_ = h.findingRepo.Insert(finding)
				}
			}
		}
	}

	if h.scanRepo != nil && scan != nil {
		now := time.Now().UTC()
		_ = h.scanRepo.UpdateStatus(scan.ID, models.StatusCompleted, &now, "")
	}

	JSON(w, http.StatusOK, results)
}
