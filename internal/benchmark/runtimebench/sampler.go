package runtimebench

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

type Sampler interface {
	Start() SampleSession
}

type SampleSession interface {
	Stop() []pkgBenchmark.Measurement
}

type noopSampler struct{}
type noopSession struct{}

func NoopSampler() Sampler                           { return noopSampler{} }
func (noopSampler) Start() SampleSession             { return noopSession{} }
func (noopSession) Stop() []pkgBenchmark.Measurement { return nil }

type ProcessSampler struct {
	pid      int
	interval time.Duration
	scope    string
	read     func(int) (processSnapshot, error)
}

func NewProcessSampler(pid int, interval time.Duration) (Sampler, error) {
	if pid < 0 {
		return nil, errors.New("resource sampler PID cannot be negative")
	}
	if interval <= 0 {
		return nil, errors.New("resource sampler interval must be positive")
	}
	scope := "provider_process"
	if pid == 0 {
		pid = os.Getpid()
		scope = "maestro_process"
	}
	if runtime.GOOS != "linux" {
		return NoopSampler(), nil
	}
	if _, err := readProcessSnapshot(pid); err != nil {
		return nil, fmt.Errorf("initialize resource sampler for PID %d: %w", pid, err)
	}
	return &ProcessSampler{pid: pid, interval: interval, scope: scope, read: readProcessSnapshot}, nil
}

type processSnapshot struct {
	processTicks uint64
	systemTicks  uint64
	rssBytes     uint64
}

type processSession struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once

	mu       sync.Mutex
	first    processSnapshot
	last     processSnapshot
	have     bool
	peakRSS  uint64
	peakCPU  float64
	scope    string
	read     func(int) (processSnapshot, error)
	pid      int
	interval time.Duration
}

func (s *ProcessSampler) Start() SampleSession {
	session := &processSession{
		stop: make(chan struct{}), done: make(chan struct{}), scope: s.scope,
		read: s.read, pid: s.pid, interval: s.interval,
	}
	session.sample()
	go session.loop()
	return session
}

func (s *processSession) loop() {
	defer close(s.done)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.sample()
		case <-s.stop:
			s.sample()
			return
		}
	}
}

func (s *processSession) sample() {
	snapshot, err := s.read(s.pid)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.have {
		s.first = snapshot
		s.last = snapshot
		s.have = true
		s.peakRSS = snapshot.rssBytes
		return
	}
	if cpu, ok := cpuPercent(s.last, snapshot); ok && cpu > s.peakCPU {
		s.peakCPU = cpu
	}
	if snapshot.rssBytes > s.peakRSS {
		s.peakRSS = snapshot.rssBytes
	}
	s.last = snapshot
}

func (s *processSession) Stop() []pkgBenchmark.Measurement {
	s.once.Do(func() { close(s.stop); <-s.done })
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.have {
		return nil
	}
	measurements := []pkgBenchmark.Measurement{{
		Name: "peak_memory_mb", Value: float64(s.peakRSS) / (1024 * 1024),
		Unit: "MiB", Scope: s.scope, Method: "linux_procfs",
	}}
	if average, ok := cpuPercent(s.first, s.last); ok {
		measurements = append(measurements,
			pkgBenchmark.Measurement{Name: "cpu_avg_percent", Value: average, Unit: "percent", Scope: s.scope, Method: "linux_procfs"},
			pkgBenchmark.Measurement{Name: "cpu_peak_percent", Value: s.peakCPU, Unit: "percent", Scope: s.scope, Method: "linux_procfs"},
		)
	}
	return measurements
}

func cpuPercent(previous, current processSnapshot) (float64, bool) {
	if current.systemTicks <= previous.systemTicks || current.processTicks < previous.processTicks {
		return 0, false
	}
	processDelta := current.processTicks - previous.processTicks
	systemDelta := current.systemTicks - previous.systemTicks
	return float64(processDelta) / float64(systemDelta) * float64(runtime.NumCPU()) * 100, true
}

func readProcessSnapshot(pid int) (processSnapshot, error) {
	processStat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return processSnapshot{}, err
	}
	status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return processSnapshot{}, err
	}
	systemStat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return processSnapshot{}, err
	}
	processTicks, err := parseProcessTicks(string(processStat))
	if err != nil {
		return processSnapshot{}, err
	}
	systemTicks, err := parseSystemTicks(string(systemStat))
	if err != nil {
		return processSnapshot{}, err
	}
	rssBytes, err := parseRSSBytes(string(status))
	if err != nil {
		return processSnapshot{}, err
	}
	return processSnapshot{processTicks: processTicks, systemTicks: systemTicks, rssBytes: rssBytes}, nil
}

func parseProcessTicks(value string) (uint64, error) {
	closing := strings.LastIndex(value, ")")
	if closing < 0 {
		return 0, errors.New("process stat has no command terminator")
	}
	fields := strings.Fields(value[closing+1:])
	if len(fields) <= 12 {
		return 0, errors.New("process stat is incomplete")
	}
	userTicks, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse process user ticks: %w", err)
	}
	systemTicks, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse process system ticks: %w", err)
	}
	return userTicks + systemTicks, nil
}

func parseSystemTicks(value string) (uint64, error) {
	line, _, _ := strings.Cut(value, "\n")
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "cpu" {
		return 0, errors.New("system stat has no aggregate CPU row")
	}
	var total uint64
	for _, field := range fields[1:] {
		ticks, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse system CPU ticks: %w", err)
		}
		total += ticks
	}
	return total, nil
}

func parseRSSBytes(value string) (uint64, error) {
	for _, line := range strings.Split(value, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "VmRSS:" {
			kilobytes, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse resident memory: %w", err)
			}
			return kilobytes * 1024, nil
		}
	}
	return 0, errors.New("process status has no VmRSS")
}
