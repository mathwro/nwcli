# Graph Report - .  (2026-07-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 527 nodes · 1097 edges · 23 communities (21 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 178 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b0a83936`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- root.go
- NewAppModel
- TestNWAliasDefinesMain
- github.com/mathwro/nwcli
- executeForTest
- newCommandError
- runInteractiveTLS
- runInteractiveTCP
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
1. `AppModel` - 33 edges
2. `executeForTest()` - 24 edges
3. `NewAppModel()` - 24 edges
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
- `pingResultError()` --calls--> `IsICMPPermissionError()`  [INFERRED]
  cmd/ping.go → internal/checks/ping.go
- `main()` --calls--> `Execute()`  [INFERRED]
  main.go → cmd/root.go
- `runInteractiveSubnet()` --calls--> `Render()`  [INFERRED]
  cmd/root.go → internal/subnet/render.go
- `runInteractiveSubnet()` --calls--> `ParsePrefixes()`  [INFERRED]
  cmd/root.go → internal/subnet/subnet.go

## Import Cycles
- None detected.

## Communities (23 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.08
Nodes (24): CancelFunc, runInteractiveMenu(), Context, appView(), Cmd, Context, Model, Msg (+16 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (42): ErrICMPPermission, PingAttempt, PingOptions, pingSession, pingSocket, PingSummary, pingTracker, PingUpdate (+34 more)

### Community 2 - "Plan"
Cohesion: 0.17
Nodes (20): validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes(), Plan(), T (+12 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.13
Nodes (23): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpOutput(), httpResultError() (+15 more)

### Community 4 - "root.go"
Cohesion: 0.09
Nodes (30): dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand(), renderDNS(), supportedDNSType(), dnsResultJSON (+22 more)

### Community 6 - "NewAppModel"
Cohesion: 0.25
Nodes (24): NewAppModel(), Context, T, testActions(), TestAppBooleanFieldUsesToggleControl(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQEditsInputInsteadOfQuitting() (+16 more)

### Community 11 - "executeForTest"
Cohesion: 0.06
Nodes (42): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+34 more)

### Community 12 - "newCommandError"
Cohesion: 0.11
Nodes (38): BatchItem, addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand() (+30 more)

### Community 13 - "runInteractiveTLS"
Cohesion: 0.13
Nodes (23): CertPool, CertificateInfo, TLSOptions, TLSResult, certificateJSON, runInteractiveTLS(), Command, Duration (+15 more)

### Community 14 - "runInteractiveTCP"
Cohesion: 0.14
Nodes (21): ContextDialer, contextDialFunc, TCPOptions, TCPResult, runInteractiveTCP(), Command, Writer, newTCPCommand() (+13 more)

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

- **Why does `Result` connect `Result` to `AppModel`, `CheckHTTP`, `root.go`, `runInteractiveTLS`, `runInteractiveTCP`?**
  _High betweenness centrality (0.180) - this node is a cross-community bridge._
- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `root.go`, `executeForTest`, `runInteractiveTLS`, `runInteractiveTCP`?**
  _High betweenness centrality (0.168) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `executeForTest` to `ping.go`, `CheckHTTP`, `root.go`, `newCommandError`, `runInteractiveTLS`, `runInteractiveTCP`?**
  _High betweenness centrality (0.135) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `NewAppModel()` (e.g. with `TestAppBooleanFieldUsesToggleControl()` and `TestAppFormCtrlCQuits()`) actually correct?**
  _`NewAppModel()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose` to the rest of the system?**
  _55 weakly-connected nodes found - possible documentation gaps or missing edges._