# Graph Report - nwcli  (2026-07-16)

## Corpus Check
- 58 files · ~29,310 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 551 nodes · 1137 edges · 27 communities (25 shown, 2 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 187 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `07d0f3a8`
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
- RunBatch
- File Structure
- AGENTS.md
- Compact TUI Redesign Implementation Plan
- InstallLatest
- update_test.go

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 34 edges
2. `executeForTest()` - 27 edges
3. `NewAppModel()` - 26 edges
4. `newCommandError()` - 23 edges
5. `Result` - 17 edges
6. `emitCommandResult()` - 15 edges
7. `NewRootCommand()` - 15 edges
8. `testActions()` - 15 edges
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
- `pingResultError()` --calls--> `IsICMPPermissionError()`  [INFERRED]
  cmd/ping.go → internal/checks/ping.go

## Import Cycles
- None detected.

## Communities (27 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.10
Nodes (19): CancelFunc, appView(), Cmd, Context, Model, Msg, View, isCtrlC() (+11 more)

### Community 1 - "ping.go"
Cohesion: 0.11
Nodes (42): ErrICMPPermission, PingAttempt, PingOptions, pingSession, pingSocket, PingSummary, pingTracker, PingUpdate (+34 more)

### Community 2 - "Plan"
Cohesion: 0.17
Nodes (21): runInteractiveSubnet(), validatePrefixes(), allocationBar(), Writer, Render(), ipToUint32(), ParsePrefixes(), Plan() (+13 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.13
Nodes (23): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpOutput(), httpResultError() (+15 more)

### Community 4 - "root.go"
Cohesion: 0.23
Nodes (14): fieldValidator(), T, TestInteractiveActionFieldsValidateBeforeRunning(), TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(), TestInteractiveActionsUseTogglesAndPlaceholders(), TestInteractiveContinuousPingReportsLiveLoss(), TestInteractivePingCountOnlyAppearsForFiniteMode(), TestInteractiveSubnetBuildsStructuredAllocationResult() (+6 more)

### Community 5 - "stubDNSResolver"
Cohesion: 0.08
Nodes (32): certificateJSON, dnsOutput(), dnsResultError(), Command, Writer, newDNSCommand(), renderDNS(), supportedDNSType() (+24 more)

### Community 6 - "NewAppModel"
Cohesion: 0.16
Nodes (30): Context, NewAppModel(), RunApp(), Context, T, testActions(), TestAppBooleanFieldUsesToggleControl(), TestAppFormCtrlCQuits() (+22 more)

### Community 11 - "executeForTest"
Cohesion: 0.08
Nodes (36): T, TestBatchDNSCommandWritesJSONResults(), TestBatchTCPCommandPreservesTargetOrder(), TestBatchTLSCommandReturnsUnhealthyForExpiringCertificate(), executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields() (+28 more)

### Community 12 - "newCommandError"
Cohesion: 0.09
Nodes (40): addBatchFlags(), batchFailure(), Command, newBatchCommand(), newBatchDNSCommand(), newBatchTCPCommand(), newBatchTLSCommand(), validateHostPorts() (+32 more)

### Community 13 - "runInteractiveTLS"
Cohesion: 0.23
Nodes (12): CertPool, CertificateInfo, TLSOptions, TLSResult, CheckTLS(), Context, Duration, T (+4 more)

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
Nodes (16): Bounded batch checks, CIDR calculator, Development, DNS diagnostics, Features, HTTP/S check, Install, Interactive mode (+8 more)

### Community 19 - "CheckDNS"
Cohesion: 0.15
Nodes (17): DNSOptions, DNSResolver, DNSResult, stubDNSResolver, CheckDNS(), dnsResolver(), Context, Duration (+9 more)

### Community 20 - "Compact TUI Redesign"
Cohesion: 0.15
Nodes (12): Acceptance Criteria, Architecture, Compact TUI Redesign, Error Handling, Goal, In Scope, Interaction Design, Out of Scope (+4 more)

### Community 21 - "RunBatch"
Cohesion: 0.33
Nodes (7): BatchItem, Context, T, RunBatch(), T, TestRunBatchBoundsConcurrencyAndPreservesOrder(), TestRunBatchUsesSafeDefaultConcurrency()

### Community 22 - "File Structure"
Cohesion: 0.20
Nodes (9): File Structure, Global Constraints, Network Toolbox Implementation Plan, Task 1: Subnet Allocator and Renderer, Task 2: HTTP/S Checker, Task 3: Ping Summary and ICMP Checker, Task 4: Bubble Tea Menu, Prompts, and Progress Models, Task 5: Cobra Command Wiring (+1 more)

### Community 23 - "AGENTS.md"
Cohesion: 0.29
Nodes (5): Commands, Development Rules, Package Boundaries, Project Purpose, Terminal UI Guidelines

### Community 24 - "Compact TUI Redesign Implementation Plan"
Cohesion: 0.33
Nodes (5): Compact TUI Redesign Implementation Plan, File Structure, Global Constraints, Task 1: TUI App State and Navigation, Task 2: Command Integration and Smoke Tests

### Community 25 - "InstallLatest"
Cohesion: 0.53
Nodes (5): Context, Writer, InstallLatest(), runCommand(), runner

### Community 26 - "update_test.go"
Cohesion: 0.67
Nodes (3): T, TestInstallLatestRunsGoInstallForLatestModule(), TestInstallLatestWrapsGoInstallFailure()

## Knowledge Gaps
- **57 isolated node(s):** `updateResult`, `github.com/mathwro/nwcli`, `actionStartedMsg`, `Project Purpose`, `Commands` (+52 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Result` connect `Result` to `AppModel`, `Plan`, `CheckHTTP`, `stubDNSResolver`, `NewAppModel`, `runInteractiveTCP`?**
  _High betweenness centrality (0.170) - this node is a cross-community bridge._
- **Why does `newCommandError()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `stubDNSResolver`, `executeForTest`, `runInteractiveTCP`?**
  _High betweenness centrality (0.159) - this node is a cross-community bridge._
- **Why does `NewRootCommand()` connect `newCommandError` to `ping.go`, `CheckHTTP`, `stubDNSResolver`, `executeForTest`, `runInteractiveTCP`?**
  _High betweenness centrality (0.139) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `executeForTest()` (e.g. with `TestBatchDNSCommandWritesJSONResults()` and `TestBatchTCPCommandPreservesTargetOrder()`) actually correct?**
  _`executeForTest()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **Are the 22 inferred relationships involving `NewAppModel()` (e.g. with `TestAppBooleanFieldUsesToggleControl()` and `TestAppFormCtrlCQuits()`) actually correct?**
  _`NewAppModel()` has 22 INFERRED edges - model-reasoned connections that need verification._
- **Are the 21 inferred relationships involving `newCommandError()` (e.g. with `batchFailure()` and `newBatchDNSCommand()`) actually correct?**
  _`newCommandError()` has 21 INFERRED edges - model-reasoned connections that need verification._
- **What connects `updateResult`, `github.com/mathwro/nwcli`, `actionStartedMsg` to the rest of the system?**
  _57 weakly-connected nodes found - possible documentation gaps or missing edges._