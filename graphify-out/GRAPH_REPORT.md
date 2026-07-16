# Graph Report - nwcli  (2026-07-16)

## Corpus Check
- 54 files · ~28,599 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 530 nodes · 1106 edges · 24 communities (22 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 180 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b6b57456`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- root.go
- stubDNSResolver
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
1. `AppModel` - 34 edges
2. `NewAppModel()` - 26 edges
3. `executeForTest()` - 24 edges
4. `newCommandError()` - 22 edges
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

## Communities (24 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.08
Nodes (24): CancelFunc, Context, appView(), Cmd, Context, Model, Msg, View (+16 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (43): ErrICMPPermission, PingAttempt, PingOptions, pingSession, pingSocket, PingSummary, pingTracker, PingUpdate (+35 more)

### Community 2 - "Plan"
Cohesion: 0.17
Nodes (20): validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes(), Plan(), T (+12 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.14
Nodes (22): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpOutput(), httpResultError() (+14 more)

### Community 4 - "root.go"
Cohesion: 0.40
Nodes (9): fieldValidator(), T, TestInteractiveActionFieldsValidateBeforeRunning(), TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(), TestInteractiveActionsUseTogglesAndPlaceholders(), TestInteractiveContinuousPingReportsLiveLoss(), TestInteractivePingCountOnlyAppearsForFiniteMode(), TestInteractiveSubnetBuildsStructuredAllocationResult() (+1 more)

### Community 5 - "stubDNSResolver"
Cohesion: 0.31
Nodes (5): stubDNSResolver, Context, IPAddr, MX, NS

### Community 6 - "NewAppModel"
Cohesion: 0.23
Nodes (26): NewAppModel(), Context, T, testActions(), TestAppBooleanFieldUsesToggleControl(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQEditsInputInsteadOfQuitting() (+18 more)

### Community 11 - "executeForTest"
Cohesion: 0.06
Nodes (42): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+34 more)

### Community 12 - "newCommandError"
Cohesion: 0.11
Nodes (38): BatchItem, addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand() (+30 more)

### Community 13 - "runInteractiveTLS"
Cohesion: 0.13
Nodes (23): CertPool, CertificateInfo, TLSOptions, TLSResult, certificateJSON, validateExtendedDuration(), Command, Duration (+15 more)

### Community 14 - "runInteractiveTCP"
Cohesion: 0.14
Nodes (20): ContextDialer, contextDialFunc, TCPOptions, TCPResult, Command, Writer, newTCPCommand(), renderTCP() (+12 more)

### Community 15 - "calc.go"
Cohesion: 0.22
Nodes (20): validateCIDR(), Int, addressCount(), addrToBig(), bigToAddr(), ContainsCIDR(), Addr, InspectCIDR() (+12 more)

### Community 16 - "Network Toolbox Design"
Cohesion: 0.11
Nodes (18): Acceptance Criteria, AGENTS.md Content, Architecture, Command Design, Current Project State, Dependencies, Error Handling, Goal (+10 more)

### Community 17 - "Result"
Cohesion: 0.08
Nodes (40): Builder, boolResultField(), Context, pingProgressEvent(), reportSingle(), runInteractiveCIDR(), runInteractiveDNS(), runInteractiveHTTP() (+32 more)

### Community 18 - "Usage"
Cohesion: 0.12
Nodes (15): Bounded batch checks, CIDR calculator, Development, DNS diagnostics, Features, HTTP/S check, Install, Interactive mode (+7 more)

### Community 19 - "CheckDNS"
Cohesion: 0.16
Nodes (21): DNSOptions, DNSResolver, DNSResult, dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand() (+13 more)

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

- **Why does `Result` connect `Result` to `AppModel`?**
  _High betweenness centrality (0.182) - this node is a cross-community bridge._
- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `executeForTest`, `runInteractiveTLS`, `runInteractiveTCP`, `CheckDNS`?**
  _High betweenness centrality (0.166) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `executeForTest` to `ping.go`, `CheckHTTP`, `newCommandError`, `runInteractiveTLS`, `runInteractiveTCP`, `Result`, `CheckDNS`?**
  _High betweenness centrality (0.134) - this node is a cross-community bridge._
- **Are the 22 inferred relationships involving `NewAppModel()` (e.g. with `TestAppBooleanFieldUsesToggleControl()` and `TestAppFormCtrlCQuits()`) actually correct?**
  _`NewAppModel()` has 22 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose` to the rest of the system?**
  _55 weakly-connected nodes found - possible documentation gaps or missing edges._