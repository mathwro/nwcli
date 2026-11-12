package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestTraceCommandRendersHopsAndJSON(t *testing.T) {
	orig := traceCheck
	t.Cleanup(func() { traceCheck = orig })
	traceCheck = func(_ context.Context, opts checks.TraceOptions) (checks.TraceResult, error) {
		return checks.TraceResult{
			Host:    opts.Host,
			Target:  "198.51.100.10",
			Reached: true,
			Hops: []checks.TraceHop{
				{TTL: 1, Address: "192.0.2.1", Latency: 2 * time.Millisecond},
				{TTL: 2, Timeout: true},
				{TTL: 3, Address: "198.51.100.10", Latency: 9 * time.Millisecond, Reached: true},
			},
		}, nil
	}
	stdout, _, err := executeForTest("trace", "example.com", "--max-hops", "8")
	if err != nil {
		t.Fatalf("trace returned error: %v", err)
	}
	for _, want := range []string{"192.0.2.1", "*", "198.51.100.10"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("trace stdout missing %q:\n%s", want, stdout)
		}
	}

	jsonOut, _, err := executeForTest("--output", "json", "trace", "example.com")
	if err != nil {
		t.Fatalf("trace json returned error: %v", err)
	}
	var env struct {
		Command string `json:"command"`
		Result  struct {
			Reached bool `json:"reached"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &env); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if env.Command != "trace" || !env.Result.Reached {
		t.Fatalf("envelope = %#v", env)
	}
}

func TestInterfacesAndCompletionCommands(t *testing.T) {
	orig := listInterfaces
	t.Cleanup(func() { listInterfaces = orig })
	listInterfaces = func(includeDown bool) ([]checks.InterfaceInfo, error) {
		return []checks.InterfaceInfo{{Name: "eth0", Index: 2, MTU: 1500, Up: true, HardwareAddr: "02:42:ac:11:00:02", Flags: []string{"up", "broadcast"}, Addresses: []string{"10.0.0.15/24"}}}, nil
	}

	stdout, _, err := executeForTest("interfaces")
	if err != nil {
		t.Fatalf("interfaces returned error: %v", err)
	}
	if !strings.Contains(stdout, "eth0") || !strings.Contains(stdout, "10.0.0.15/24") {
		t.Fatalf("interfaces stdout = %q", stdout)
	}

	compOut, _, err := executeForTest("completion", "bash")
	if err != nil {
		t.Fatalf("completion bash returned error: %v", err)
	}
	if !strings.Contains(compOut, "nwcli") {
		t.Fatalf("completion output missing nwcli script")
	}
}
