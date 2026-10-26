# BHF (Bug Hunting Framework) - Development Workflow and Operational Runbook

This document defines the standard engineering lifecycle for adding features, integrating security tools, managing dependencies, debugging defects, and tracking progress.

---

## 1. Feature Addition Lifecycle

Every code modification or feature addition must follow this 7-step sequence:

```
[1. Recon Value Assessment] 
       |
       v
[2. Architecture & SRP Placement]
       |
       v
[3. Test Definition]
       |
       v
[4. Clean Implementation]
       |
       v
[5. Rigorous Verification (go test & npm build)]
       |
       v
[6. Changelog Documentation]
       |
       v
[7. Atomic Conventional Git Commit]
```

### Step 1: Reconnaissance Value and Traffic Safety Assessment
Before writing any code, answer these three mandatory questions:
1. **Does this directly assist a bug bounty hunter?** Features that do not yield high-signal intelligence or speed up analysis are rejected.
2. **Will this trigger excessive network traffic or WAF bans?** Any active network requests must include rate limiting (`-rl`), concurrency bounds (`-c`), and timeout controls.
3. **Is scope strictly maintained?** Ensure that no third-party assets can leak into the database.

### Step 2: Architecture and Single Responsibility Placement
Determine exactly which files need to be created or modified according to Single Responsibility Principle:
- New security tool? Add exactly one file in `internal/tools/<tool_name>.go` implementing `tools.Tool`.
- New database entity? Add `internal/models/<entity>.go`, `internal/database/<entity>_repo.go`, update `internal/database/schema.go`, and add `web/src/types/<entity>.ts`.
- New UI panel? Create a dedicated view in `web/src/views/` and granular components in `web/src/components/`.

### Step 3: Test Definition
Define tests for the new functionality before or alongside code implementation:
- If adding a tool, prepare sample output lines and test its `Run()` parser.
- If modifying database logic, add a test in `internal/database/<entity>_repo_test.go`.

### Step 4: Clean Implementation
- Write idiomatic, clean Go code using typed structs, explicit error checking (`if err != nil`), and clean interfaces.
- Write modern TypeScript React code using functional components with explicit types and dedicated service clients.
- Adhere strictly to the GitHub Dark UI theme (`#0d1117`) with zero neon colors, zero fluff text, and zero emojis.

### Step 5: Rigorous Verification
Run all automated test suites to ensure zero regressions:
```bash
go test -v ./...
go test -race ./...
cd web && npm run build && cd ..
```

### Step 6: Progress Tracking
Record the change in `CHANGELOG.md` under the `[Unreleased]` section following the standard categories (`Added`, `Changed`, `Fixed`, `Removed`, `Security`).

### Step 7: Atomic Conventional Git Commit
Commit the changes using a descriptive Conventional Commit message:
```bash
git add <specific-files>
git commit -m "feat(tools): add katana crawler integration with domain scoping"
```

---

## 2. Python Tool Isolation Runbook

To add or update a Python-based security tool (e.g., `theHarvester`, `wafw00f`):

1. **Never use global pip**: Installing packages into the host OS or global container environment is strictly forbidden.
2. **Use uv tool isolation**:
   Inside the Dockerfile, install tools into their own virtual environment using `uv`:
   ```dockerfile
   RUN uv tool install --python 3.14 git+https://github.com/laramies/theHarvester.git
   ```
3. **Subprocess Execution in Go**:
   Configure the tool's `BinaryName()` in `internal/tools/<tool>.go` to check for the installed executable:
   ```go
   func (t *Tool) BinaryName() string {
       if _, err := exec.LookPath("tool-name"); err == nil {
           return "tool-name"
       }
       return "uv"
   }
   ```
4. **Local Development Virtual Environment**:
   If developing Python helper scripts locally, create a project-local virtual environment:
   ```bash
   uv venv .venv
   source .venv/bin/activate
   ```
   Ensure `.venv` is ignored by `.gitignore`.

---

## 3. Methodical Investigation Runbook (Zero Guesswork)

When a bug occurs (e.g., scan fails unexpectedly, output is not parsed, database error):

### Step 1: Capture the Exact Output and Command
- Review the terminal logs for the scan session via the Console tab or SQLite:
  ```bash
  sqlite3 data/bhf.db "SELECT stream, line FROM logs WHERE scan_id = 'TARGET_SCAN_ID';"
  ```
- Identify the exact command line arguments and whether the binary returned a non-zero exit code.

### Step 2: Validate the Subprocess and Binary Environment
- Test whether the binary is in `$PATH`:
  ```bash
  which <tool-name>
  ```
- Run the tool directly with the same arguments against a test domain to observe raw stdout and stderr behavior.

### Step 3: Inspect Database Records and Constraints
- Check SQLite for table schema mismatches or foreign key violations:
  ```bash
  sqlite3 data/bhf.db ".schema scans"
  sqlite3 data/bhf.db "SELECT * FROM scans WHERE id = 'TARGET_SCAN_ID';"
  ```

### Step 4: Write an Isolating Unit Test
- Create a test case reproducing the failing output or invalid input inside the corresponding `_test.go` file.
- Verify that the test fails as expected.

### Step 5: Implement the Fix and Confirm Resolution
- Fix the root cause in the source code.
- Run `go test ./...` and verify the reproducing test now passes.
- Test in the running application to confirm end-to-end resolution.

---

## 4. Starter Prompt for New Conversations

When opening a new session or thread with the AI assistant, use the following starter prompt to immediately establish the project context, architectural guidelines, and operating rules:

```text
You are working on BHF (Bug Hunting Framework) located at /home/diamond/Desktop/Github/bhf.

Before proposing or making any changes, you must read:
1. ARCHITECTURE.md (system design, models, directory structure)
2. RULES.md (engineering rules, UI standards, Single Responsibility Principle, zero emojis, zero guesswork)
3. TESTING.md (test suite architecture and table-driven test patterns)
4. WORKFLOW.md (feature addition lifecycle, dependency isolation, debugging runbook)
5. CHANGELOG.md (current progress and recent modifications)

Guidelines for your response:
- Follow the Single Responsibility Principle: each file has one single job.
- Use Conventional Commits naming methodology (feat, fix, refactor, test, docs, chore).
- UI must remain clean, minimal, developer-grade GitHub Dark (#0d1117). Strictly NO AI glow, NO neon tropes, NO emojis.
- Every feature must directly serve the bug hunting cause without generating excessive or reckless traffic.
- Never guess causes of bugs; investigate methodically using logs and reproduction.
- Ensure all Go tests pass via 'go test ./...' before completing your task.
- Track all changes in CHANGELOG.md.

Current Task:
<DESCRIBE YOUR TASK HERE>
```
