# Graph Report - nwcli  (2026-07-15)

## Corpus Check
- 53 files · ~25,411 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 497 nodes · 991 edges · 25 communities (23 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 163 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ea24770d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- root.go
- ProgressEvent
- NewAppModel
- TestNWAliasDefinesMain
- github.com/mathwro/nwcli
- executeForTest
- newCommandError
- newTLSCommand
- newTCPCommand
- calc.go
- Network Toolbox Design
- newBatchTLSCommand
- Usage
- CheckDNS
- Compact TUI Redesign
- stubDNSResolver
- File Structure
- AGENTS.md
- Compact TUI Redesign Implementation Plan

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 29 edges
2. `executeForTest()` - 22 edges
3. `newCommandError()` - 21 edges
4. `NewAppModel()` - 19 edges
5. `Result` - 17 edges
6. `testActions()` - 15 edges
7. `emitCommandResult()` - 14 edges
8. `NewRootCommand()` - 14 edges
9. `Action` - 13 edges
10. `Network Toolbox Design` - 12 edges

## Surprising Connections (you probably didn't know these)
- `validateCIDR()` --calls--> `InspectCIDR()`  [INFERRED]
  cmd/root.go → internal/subnet/calc.go
- `main()` --calls--> `Execute()`  [INFERRED]
  main.go → cmd/root.go
- `runInteractiveMenu()` --calls--> `RunApp()`  [INFERRED]
  cmd/root.go → internal/tui/app.go
- `runInteractiveCIDR()` --calls--> `InspectCIDR()`  [INFERRED]
  cmd/root.go → internal/subnet/calc.go
- `validatePrefixes()` --calls--> `ParsePrefixes()`  [INFERRED]
  cmd/root.go → internal/subnet/subnet.go

## Import Cycles
- None detected.

## Communities (25 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.12
Nodes (16): appView(), Cmd, Context, Model, Msg, View, isCtrlC(), keyName() (+8 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (35): ErrICMPPermission, PingAttempt, PingOptions, pingSocket, PingSummary, Command, Writer, newPingCommand() (+27 more)

### Community 2 - "Plan"
Cohesion: 0.17
Nodes (21): runInteractiveSubnet(), validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes(), Plan() (+13 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.20
Nodes (15): HTTPOptions, HTTPResult, roundTripFunc, Client, CheckHTTP(), Context, Duration, T (+7 more)

### Community 4 - "root.go"
Cohesion: 0.09
Nodes (26): dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand(), renderDNS(), supportedDNSType(), dnsResultJSON (+18 more)

### Community 5 - "ProgressEvent"
Cohesion: 0.19
Nodes (11): Cmd, Context, Model, Msg, View, NewProgressModel(), RunProgress(), waitProgressEvent() (+3 more)

### Community 6 - "NewAppModel"
Cohesion: 0.31
Nodes (19): NewAppModel(), Context, T, testActions(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQEditsInputInsteadOfQuitting(), TestAppFormValidationKeepsFocusAndShowsGuidance() (+11 more)

### Community 11 - "executeForTest"
Cohesion: 0.07
Nodes (39): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+31 more)

### Community 12 - "newCommandError"
Cohesion: 0.11
Nodes (38): BatchItem, addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand() (+30 more)

### Community 13 - "newTLSCommand"
Cohesion: 0.13
Nodes (23): CertPool, CertificateInfo, TLSOptions, TLSResult, certificateJSON, runInteractiveTLS(), Command, Duration (+15 more)

### Community 14 - "newTCPCommand"
Cohesion: 0.14
Nodes (21): ContextDialer, contextDialFunc, TCPOptions, TCPResult, runInteractiveTCP(), Command, Writer, newTCPCommand() (+13 more)

### Community 15 - "calc.go"
Cohesion: 0.24
Nodes (19): Int, addressCount(), addrToBig(), bigToAddr(), ContainsCIDR(), Addr, InspectCIDR(), OverlapsCIDR() (+11 more)

### Community 16 - "Network Toolbox Design"
Cohesion: 0.11
Nodes (18): Acceptance Criteria, AGENTS.md Content, Architecture, Command Design, Current Project State, Dependencies, Error Handling, Goal (+10 more)

### Community 17 - "newBatchTLSCommand"
Cohesion: 0.21
Nodes (18): Builder, renderAllocationMap(), renderAllocationRail(), renderRailEndpoints(), renderResultBody(), renderResultFields(), renderResultHeader(), resultClipboardText() (+10 more)

### Community 18 - "Usage"
Cohesion: 0.12
Nodes (15): Bounded batch checks, CIDR calculator, Development, DNS diagnostics, Features, HTTP/S check, Install, Interactive mode (+7 more)

### Community 19 - "CheckDNS"
Cohesion: 0.15
Nodes (17): DNSOptions, DNSResolver, DNSResult, stubDNSResolver, CheckDNS(), dnsResolver(), Context, Duration (+9 more)

### Community 20 - "Compact TUI Redesign"
Cohesion: 0.15
Nodes (12): Acceptance Criteria, Architecture, Compact TUI Redesign, Error Handling, Goal, In Scope, Interaction Design, Out of Scope (+4 more)

### Community 21 - "stubDNSResolver"
Cohesion: 0.19
Nodes (10): fieldValidator(), T, TestInteractiveActionFieldsValidateBeforeRunning(), TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(), TestInteractiveSubnetBuildsStructuredAllocationResult(), interactiveActions(), Context, Action (+2 more)

### Community 22 - "File Structure"
Cohesion: 0.20
Nodes (9): File Structure, Global Constraints, Network Toolbox Implementation Plan, Task 1: Subnet Allocator and Renderer, Task 2: HTTP/S Checker, Task 3: Ping Summary and ICMP Checker, Task 4: Bubble Tea Menu, Prompts, and Progress Models, Task 5: Cobra Command Wiring (+1 more)

### Community 23 - "AGENTS.md"
Cohesion: 0.33
Nodes (4): Commands, Development Rules, Package Boundaries, Project Purpose

### Community 24 - "Compact TUI Redesign Implementation Plan"
Cohesion: 0.33
Nodes (5): Compact TUI Redesign Implementation Plan, File Structure, Global Constraints, Task 1: TUI App State and Navigation, Task 2: Command Integration and Smoke Tests

## Knowledge Gaps
- **54 isolated node(s):** `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose`, `Commands`, `Package Boundaries` (+49 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Result` connect `newBatchTLSCommand` to `AppModel`, `ping.go`, `Plan`, `root.go`, `newTLSCommand`, `newTCPCommand`, `stubDNSResolver`?**
  _High betweenness centrality (0.188) - this node is a cross-community bridge._
- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `root.go`, `executeForTest`, `newTLSCommand`, `newTCPCommand`?**
  _High betweenness centrality (0.168) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `executeForTest` to `ping.go`, `root.go`, `newCommandError`, `newTLSCommand`, `newTCPCommand`?**
  _High betweenness centrality (0.133) - this node is a cross-community bridge._
- **Are the 18 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 18 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `NewAppModel()` (e.g. with `TestAppFormCtrlCQuits()` and `TestAppFormEscapeReturnsToMenuWithoutQuitting()`) actually correct?**
  _`NewAppModel()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose` to the rest of the system?**
  _54 weakly-connected nodes found - possible documentation gaps or missing edges._