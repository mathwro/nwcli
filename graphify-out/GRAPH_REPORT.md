# Graph Report - .  (2026-07-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 524 nodes · 1085 edges · 24 communities (22 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 176 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `82ffcc84`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- root.go
- RunBatch
- NewAppModel
- TestNWAliasDefinesMain
- github.com/mathwro/nwcli
- executeForTest
- newCommandError
- runInteractiveTLS
- CheckTCP
- calc.go
- Network Toolbox Design
- Result
- Usage
- CheckDNS
- Compact TUI Redesign
- File Structure
- AGENTS.md
- Compact TUI Redesign Implementation Plan

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 32 edges
2. `executeForTest()` - 24 edges
3. `NewAppModel()` - 23 edges
4. `newCommandError()` - 22 edges
5. `Result` - 17 edges
6. `testActions()` - 15 edges
7. `emitCommandResult()` - 14 edges
8. `NewRootCommand()` - 14 edges
9. `ProgressEvent` - 14 edges
10. `Action` - 13 edges

## Surprising Connections (you probably didn't know these)
- `validateCIDR()` --calls--> `InspectCIDR()`  [INFERRED]
  cmd/root.go → internal/subnet/calc.go
- `newBatchTCPCommand()` --calls--> `RunBatch()`  [INFERRED]
  cmd/batch.go → internal/checks/batch.go
- `newBatchDNSCommand()` --calls--> `RunBatch()`  [INFERRED]
  cmd/batch.go → internal/checks/batch.go
- `newBatchTLSCommand()` --calls--> `RunBatch()`  [INFERRED]
  cmd/batch.go → internal/checks/batch.go
- `main()` --calls--> `Execute()`  [INFERRED]
  main.go → cmd/root.go

## Import Cycles
- None detected.

## Communities (24 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.07
Nodes (30): CancelFunc, fieldValidator(), T, TestInteractiveActionFieldsValidateBeforeRunning(), TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(), TestInteractiveActionsUseTogglesAndPlaceholders(), interactiveActions(), runInteractiveMenu() (+22 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (43): ErrICMPPermission, PingAttempt, PingOptions, pingSession, pingSocket, PingSummary, pingTracker, PingUpdate (+35 more)

### Community 2 - "Plan"
Cohesion: 0.16
Nodes (22): TestInteractiveSubnetBuildsStructuredAllocationResult(), runInteractiveSubnet(), validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes() (+14 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.13
Nodes (23): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpOutput(), httpResultError() (+15 more)

### Community 4 - "root.go"
Cohesion: 0.10
Nodes (23): dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand(), renderDNS(), supportedDNSType(), dnsResultJSON (+15 more)

### Community 5 - "RunBatch"
Cohesion: 0.33
Nodes (7): BatchItem, Context, T, RunBatch(), T, TestRunBatchBoundsConcurrencyAndPreservesOrder(), TestRunBatchUsesSafeDefaultConcurrency()

### Community 6 - "NewAppModel"
Cohesion: 0.26
Nodes (23): NewAppModel(), Context, T, testActions(), TestAppBooleanFieldUsesToggleControl(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQEditsInputInsteadOfQuitting() (+15 more)

### Community 11 - "executeForTest"
Cohesion: 0.06
Nodes (42): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+34 more)

### Community 12 - "newCommandError"
Cohesion: 0.12
Nodes (36): addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand(), validateHostPorts() (+28 more)

### Community 13 - "runInteractiveTLS"
Cohesion: 0.13
Nodes (23): CertPool, CertificateInfo, TLSOptions, TLSResult, certificateJSON, runInteractiveTLS(), Command, Duration (+15 more)

### Community 14 - "CheckTCP"
Cohesion: 0.21
Nodes (13): ContextDialer, contextDialFunc, TCPOptions, TCPResult, Conn, CheckTCP(), Context, Duration (+5 more)

### Community 15 - "calc.go"
Cohesion: 0.24
Nodes (19): Int, addressCount(), addrToBig(), bigToAddr(), ContainsCIDR(), Addr, InspectCIDR(), OverlapsCIDR() (+11 more)

### Community 16 - "Network Toolbox Design"
Cohesion: 0.11
Nodes (18): Acceptance Criteria, AGENTS.md Content, Architecture, Command Design, Current Project State, Dependencies, Error Handling, Goal (+10 more)

### Community 17 - "Result"
Cohesion: 0.11
Nodes (28): Builder, Cmd, Context, Model, Msg, View, NewProgressModel(), RunProgress() (+20 more)

### Community 18 - "Usage"
Cohesion: 0.12
Nodes (15): Bounded batch checks, CIDR calculator, Development, DNS diagnostics, Features, HTTP/S check, Install, Interactive mode (+7 more)

### Community 19 - "CheckDNS"
Cohesion: 0.15
Nodes (17): DNSOptions, DNSResolver, DNSResult, stubDNSResolver, CheckDNS(), dnsResolver(), Context, Duration (+9 more)

### Community 20 - "Compact TUI Redesign"
Cohesion: 0.15
Nodes (12): Acceptance Criteria, Architecture, Compact TUI Redesign, Error Handling, Goal, In Scope, Interaction Design, Out of Scope (+4 more)

### Community 22 - "File Structure"
Cohesion: 0.20
Nodes (9): File Structure, Global Constraints, Network Toolbox Implementation Plan, Task 1: Subnet Allocator and Renderer, Task 2: HTTP/S Checker, Task 3: Ping Summary and ICMP Checker, Task 4: Bubble Tea Menu, Prompts, and Progress Models, Task 5: Cobra Command Wiring (+1 more)

### Community 23 - "AGENTS.md"
Cohesion: 0.29
Nodes (5): Commands, Development Rules, Package Boundaries, Project Purpose, Terminal UI Guidelines

### Community 24 - "Compact TUI Redesign Implementation Plan"
Cohesion: 0.33
Nodes (5): Compact TUI Redesign Implementation Plan, File Structure, Global Constraints, Task 1: TUI App State and Navigation, Task 2: Command Integration and Smoke Tests

## Knowledge Gaps
- **55 isolated node(s):** `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose`, `Commands`, `Package Boundaries` (+50 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Result` connect `Result` to `AppModel`, `Plan`, `CheckHTTP`, `root.go`, `runInteractiveTLS`?**
  _High betweenness centrality (0.178) - this node is a cross-community bridge._
- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `root.go`, `executeForTest`, `runInteractiveTLS`?**
  _High betweenness centrality (0.170) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `executeForTest` to `ping.go`, `CheckHTTP`, `root.go`, `newCommandError`, `runInteractiveTLS`?**
  _High betweenness centrality (0.135) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `NewAppModel()` (e.g. with `TestAppBooleanFieldUsesToggleControl()` and `TestAppFormCtrlCQuits()`) actually correct?**
  _`NewAppModel()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose` to the rest of the system?**
  _55 weakly-connected nodes found - possible documentation gaps or missing edges._