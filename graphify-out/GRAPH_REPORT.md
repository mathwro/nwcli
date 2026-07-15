# Graph Report - .  (2026-07-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 516 nodes · 1062 edges · 25 communities (23 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 171 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `16ecb2e5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- Action
- ProgressModel
- NewAppModel
- TestNWAliasDefinesMain
- github.com/mathwro/nwcli
- executeForTest
- newCommandError
- newTLSCommand
- newTCPCommand
- calc.go
- Network Toolbox Design
- root.go
- Usage
- CheckDNS
- Compact TUI Redesign
- stubDNSResolver
- File Structure
- AGENTS.md
- Compact TUI Redesign Implementation Plan

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 29 edges
2. `executeForTest()` - 24 edges
3. `newCommandError()` - 22 edges
4. `NewAppModel()` - 20 edges
5. `Result` - 17 edges
6. `testActions()` - 15 edges
7. `emitCommandResult()` - 14 edges
8. `NewRootCommand()` - 14 edges
9. `ProgressEvent` - 14 edges
10. `Action` - 13 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  main.go → cmd/root.go
- `runInteractiveMenu()` --calls--> `RunApp()`  [INFERRED]
  cmd/root.go → internal/tui/app.go
- `runInteractiveSubnet()` --calls--> `Render()`  [INFERRED]
  cmd/root.go → internal/subnet/render.go
- `runInteractiveSubnet()` --calls--> `ParsePrefixes()`  [INFERRED]
  cmd/root.go → internal/subnet/subnet.go
- `runInteractiveSubnet()` --calls--> `Plan()`  [INFERRED]
  cmd/root.go → internal/subnet/subnet.go

## Import Cycles
- None detected.

## Communities (25 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.12
Nodes (16): appView(), Cmd, Context, Model, Msg, View, isCtrlC(), keyName() (+8 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (43): ErrICMPPermission, PingAttempt, PingOptions, pingSession, pingSocket, PingSummary, pingTracker, PingUpdate (+35 more)

### Community 2 - "Plan"
Cohesion: 0.17
Nodes (20): validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes(), Plan(), T (+12 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.14
Nodes (22): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpOutput(), httpResultError() (+14 more)

### Community 4 - "Action"
Cohesion: 0.18
Nodes (11): fieldValidator(), T, TestInteractiveActionFieldsValidateBeforeRunning(), TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(), TestInteractiveContinuousPingReportsLiveLoss(), TestInteractiveSubnetBuildsStructuredAllocationResult(), interactiveActions(), Context (+3 more)

### Community 5 - "ProgressModel"
Cohesion: 0.19
Nodes (10): Cmd, Context, Model, Msg, View, NewProgressModel(), RunProgress(), waitProgressEvent() (+2 more)

### Community 6 - "NewAppModel"
Cohesion: 0.30
Nodes (20): NewAppModel(), Context, T, testActions(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQEditsInputInsteadOfQuitting(), TestAppFormValidationKeepsFocusAndShowsGuidance() (+12 more)

### Community 11 - "executeForTest"
Cohesion: 0.06
Nodes (40): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+32 more)

### Community 12 - "newCommandError"
Cohesion: 0.10
Nodes (40): BatchItem, addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand() (+32 more)

### Community 13 - "newTLSCommand"
Cohesion: 0.13
Nodes (23): CertPool, CertificateInfo, TLSOptions, TLSResult, certificateJSON, validateExtendedDuration(), Command, Duration (+15 more)

### Community 14 - "newTCPCommand"
Cohesion: 0.14
Nodes (20): ContextDialer, contextDialFunc, TCPOptions, TCPResult, Command, Writer, newTCPCommand(), renderTCP() (+12 more)

### Community 15 - "calc.go"
Cohesion: 0.22
Nodes (20): validateCIDR(), Int, addressCount(), addrToBig(), bigToAddr(), ContainsCIDR(), Addr, InspectCIDR() (+12 more)

### Community 16 - "Network Toolbox Design"
Cohesion: 0.11
Nodes (18): Acceptance Criteria, AGENTS.md Content, Architecture, Command Design, Current Project State, Dependencies, Error Handling, Goal (+10 more)

### Community 17 - "root.go"
Cohesion: 0.11
Nodes (32): Builder, boolResultField(), Context, pingProgressEvent(), reportSingle(), runInteractiveCIDR(), runInteractiveDNS(), runInteractiveHTTP() (+24 more)

### Community 18 - "Usage"
Cohesion: 0.12
Nodes (15): Bounded batch checks, CIDR calculator, Development, DNS diagnostics, Features, HTTP/S check, Install, Interactive mode (+7 more)

### Community 19 - "CheckDNS"
Cohesion: 0.16
Nodes (21): DNSOptions, DNSResolver, DNSResult, dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand() (+13 more)

### Community 20 - "Compact TUI Redesign"
Cohesion: 0.15
Nodes (12): Acceptance Criteria, Architecture, Compact TUI Redesign, Error Handling, Goal, In Scope, Interaction Design, Out of Scope (+4 more)

### Community 21 - "stubDNSResolver"
Cohesion: 0.31
Nodes (5): stubDNSResolver, Context, IPAddr, MX, NS

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

- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `newTLSCommand`, `newTCPCommand`, `CheckDNS`?**
  _High betweenness centrality (0.174) - this node is a cross-community bridge._
- **Why does `Result` connect `root.go` to `AppModel`, `Action`?**
  _High betweenness centrality (0.171) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `executeForTest` to `ping.go`, `CheckHTTP`, `newCommandError`, `newTLSCommand`, `newTCPCommand`, `root.go`, `CheckDNS`?**
  _High betweenness centrality (0.137) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 16 inferred relationships involving `NewAppModel()` (e.g. with `TestAppFormCtrlCQuits()` and `TestAppFormEscapeReturnsToMenuWithoutQuitting()`) actually correct?**
  _`NewAppModel()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose` to the rest of the system?**
  _55 weakly-connected nodes found - possible documentation gaps or missing edges._