# BHF - Personal Tracking and Operator Log

This document is your personal scratchpad and tracking dashboard for recording bugs, feature ideas, and operational notes while actively using the framework on targets.

---

## 1. Active Bug Backlog

Track defects, crashes, parser anomalies, or UI issues discovered during live execution.

| ID | Date | Recon Section | Severity | Summary | Status |
|---|---|---|---|---|---|
| BUG-001 | 2026-10-26 | Core | Low | Example: Terminal autoscroll pauses when clicking inside container | Closed |

*Status Values*: `Open`, `Investigating`, `In Progress`, `Resolved`, `Closed`

---

## 2. Feature & Tool Wishlist

Track capabilities, tool integrations, and workflow improvements that increase bug bounty yield.

| ID | Date | Methodology Stage | Proposed Capability / Tool | Hunter ROI / Justification | Status |
|---|---|---|---|---|---|
| FEAT-001 | 2026-10-26 | Live HTTP Probing | TLS certificate subject alternative name (SAN) extraction | Finds hidden dev/origin hostnames on single IP addresses | Backlog |
| FEAT-002 | 2026-10-26 | Content Discovery | Custom header injection in Ffuf (Authorization / Cookie) | Enables authenticated endpoint brute forcing | Planned |

*Status Values*: `Idea`, `Planned`, `In Progress`, `Implemented`, `Rejected`

---

## 3. Live Hunting Notes & Target Observations

Operator scratchpad for noting quirks, WAF behaviors, wordlist tweaks, or high-value findings during live engagements:

- **Target Notes**:
  - Example: Keep rate limit for `example.com` under 50 req/sec to avoid Akamai rate-limiting.
  - Example: `admin-api.target.com:8443` returned a 403 forbidden; test HTTP method tampering (`X-Original-URL`, `X-Rewrite-URL`, `GET /` vs `POST /`).

---

## 4. Quick Bug Log Template

Copy and paste this template below to log a new defect:

```markdown
### [BUG-XXX] Short Bug Title
- **Date**: YYYY-MM-DD
- **Recon Stage**: [e.g., Subdomain Enumeration, Port Scanning, HTTP Probe, Crawling, Fuzzing, Vuln Assessment, or New Stage]
- **Target Domain**: target.com
- **Tool Involved**: [which tool or core system]
- **Command Run**: [exact command line from terminal view]
- **Symptom**: Describe what failed (e.g., process hung, parser dropped valid subdomains, database constraint error).
- **Reproduction Steps**:
  1. Select target `target.com`.
  2. Click Run on tool.
  3. Observe error in console.
- **Log Snippet**:
  ```text
  [paste relevant stdout/stderr lines here]
  ```
- **Investigation & Resolution**:
  - Root Cause: [Leave empty for investigation]
  - Fixed in Commit: [hash]
```

---

## 5. Quick Feature Proposal Template

Copy and paste this template below to log a new feature idea:

```markdown
### [FEAT-XXX] Short Feature Title
- **Date**: YYYY-MM-DD
- **Target Stage**: [e.g., Subdomain Enumeration, Port Scanning, HTTP Probe, Crawling, Fuzzing, Vuln Assessment, or New Stage]
- **Proposed Tool or Feature**: [name of proposed tool or capability]
- **Hunter ROI Justification**: Why does this help find bounties? What high-value signal does it extract?
- **Traffic Safety & Stealth Impact**: Will this increase network volume? What rate limit or depth cap should be default?
- **Proposed Technical Implementation**:
  - Backend: [Which internal/ package and file?]
  - Frontend: [Which component or view?]
  - Database: [Any schema changes needed?]
```
