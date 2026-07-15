# Graph Report - .  (2026-07-15)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 190 nodes · 396 edges · 11 communities (9 shown, 2 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 55 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7c1ceaa6`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- AppModel
- ping.go
- Plan
- CheckHTTP
- NewRootCommand
- ProgressModel
- NewAppModel
- TestNWAliasDefinesMain
- github.com/mathwro/nwcli

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 27 edges
2. `NewAppModel()` - 14 edges
3. `CheckPing()` - 12 edges
4. `testActions()` - 12 edges
5. `CheckHTTP()` - 11 edges
6. `Plan()` - 11 edges
7. `Action` - 10 edges
8. `ProgressModel` - 10 edges
9. `NewRootCommand()` - 9 edges
10. `listenPingSocket()` - 8 edges

## Surprising Connections (you probably didn't know these)
- `newPingCommand()` --calls--> `CheckPing()`  [INFERRED]
  cmd/ping.go → internal/checks/ping.go
- `runInteractiveMenu()` --calls--> `RunApp()`  [INFERRED]
  cmd/root.go → internal/tui/app.go
- `runInteractivePing()` --calls--> `CheckPing()`  [INFERRED]
  cmd/root.go → internal/checks/ping.go
- `newHTTPCommand()` --calls--> `CheckHTTP()`  [INFERRED]
  cmd/http.go → internal/checks/http.go
- `renderPing()` --references--> `PingAttempt`  [EXTRACTED]
  cmd/ping.go → internal/checks/ping.go

## Import Cycles
- None detected.

## Communities (11 total, 2 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.12
Nodes (17): appView(), Cmd, Context, Model, Msg, View, isCtrlC(), keyName() (+9 more)

### Community 1 - "ping.go"
Cohesion: 0.16
Nodes (25): Addr, ErrICMPPermission, PingAttempt, PingOptions, pingSocket, PingSummary, CheckPing(), Context (+17 more)

### Community 2 - "Plan"
Cohesion: 0.15
Nodes (23): runInteractiveSubnet(), Command, newSubnetCarveCommand(), newSubnetCommand(), allocationBar(), Writer, Render(), ipToUint32() (+15 more)

### Community 3 - "CheckHTTP"
Cohesion: 0.14
Nodes (21): HTTPOptions, HTTPResult, roundTripFunc, Client, Command, Writer, httpResultError(), newHTTPCommand() (+13 more)

### Community 4 - "NewRootCommand"
Cohesion: 0.13
Nodes (20): executeForTest(), T, TestHTTPCommandRejectsMissingHost(), TestInteractiveActionsUseTUIAppFields(), TestRootNoArgsRunsInjectedMenu(), TestSubnetCarveCommandPrintsAllocation(), TestSubnetCarveRequiresCIDR(), main() (+12 more)

### Community 5 - "ProgressModel"
Cohesion: 0.19
Nodes (11): Cmd, Context, Model, Msg, View, NewProgressModel(), RunProgress(), waitProgressEvent() (+3 more)

### Community 6 - "NewAppModel"
Cohesion: 0.27
Nodes (17): Context, NewAppModel(), Context, T, testActions(), TestAppFormCtrlCQuits(), TestAppFormEscapeReturnsToMenuWithoutQuitting(), TestAppFormQQuits() (+9 more)

## Knowledge Gaps
- **3 isolated node(s):** `github.com/mathwro/nwcli`, `actionStartedMsg`, `actionResultMsg`
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Action` connect `NewAppModel` to `AppModel`, `NewRootCommand`, `ProgressModel`?**
  _High betweenness centrality (0.307) - this node is a cross-community bridge._
- **Why does `runInteractivePing()` connect `NewRootCommand` to `ping.go`, `ProgressModel`?**
  _High betweenness centrality (0.295) - this node is a cross-community bridge._
- **Why does `AppModel` connect `AppModel` to `ProgressModel`, `NewAppModel`?**
  _High betweenness centrality (0.253) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `NewAppModel()` (e.g. with `TestAppFormCtrlCQuits()` and `TestAppFormEscapeReturnsToMenuWithoutQuitting()`) actually correct?**
  _`NewAppModel()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `CheckPing()` (e.g. with `newPingCommand()` and `runInteractivePing()`) actually correct?**
  _`CheckPing()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `CheckHTTP()` (e.g. with `newHTTPCommand()` and `runInteractiveHTTP()`) actually correct?**
  _`CheckHTTP()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/mathwro/nwcli`, `actionStartedMsg`, `actionResultMsg` to the rest of the system?**
  _3 weakly-connected nodes found - possible documentation gaps or missing edges._