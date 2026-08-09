package runtimebench

import (
	"runtime"
	"testing"
)

func TestProcParsersAndCPUCalculation(t *testing.T) {
	processTicks, err := parseProcessTicks("123 (maestro worker) S 1 2 3 4 5 6 7 8 9 10 20 5 0")
	if err != nil || processTicks != 25 {
		t.Fatalf("process ticks=%d err=%v", processTicks, err)
	}
	systemTicks, err := parseSystemTicks("cpu  10 20 30 40 5\ncpu0 1 2 3 4\n")
	if err != nil || systemTicks != 105 {
		t.Fatalf("system ticks=%d err=%v", systemTicks, err)
	}
	rss, err := parseRSSBytes("Name:\tmaestro\nVmRSS:\t2048 kB\n")
	if err != nil || rss != 2*1024*1024 {
		t.Fatalf("rss=%d err=%v", rss, err)
	}
	percent, ok := cpuPercent(
		processSnapshot{processTicks: 10, systemTicks: 100},
		processSnapshot{processTicks: 20, systemTicks: 200},
	)
	want := float64(runtime.NumCPU()) * 10
	if !ok || percent != want {
		t.Fatalf("cpu=%f ok=%t want=%f", percent, ok, want)
	}
}

func TestCPUCalculationOmitsUnavailableDelta(t *testing.T) {
	if _, ok := cpuPercent(processSnapshot{systemTicks: 10}, processSnapshot{systemTicks: 10}); ok {
		t.Fatal("zero system delta should be unavailable")
	}
}
