// m36residency measures the frozen chat/mutation/chat residency sequence once.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"gopkg.in/yaml.v3"
)

type identity struct {
	Name, Digest string
}

type profile struct {
	Version  int
	Status   string
	Provider struct {
		Version string `yaml:"version"`
		BaseURL string `yaml:"base_url"`
	}
	Models struct {
		DirectChat         identity `yaml:"direct_chat"`
		ControlledMutation identity `yaml:"controlled_mutation"`
	}
	Generation struct {
		Cycles          int    `yaml:"cycles"`
		ContextWindow   int    `yaml:"context_window"`
		MaxOutputTokens int    `yaml:"max_output_tokens"`
		KeepAlive       string `yaml:"keep_alive"`
	}
	LatencyLimits map[string]int `yaml:"latency_limits_ms"`
	Reference     struct {
		WSLConfigPath   string `yaml:"wslconfig_path"`
		WSLConfigSHA256 string `yaml:"wslconfig_sha256"`
	} `yaml:"reference_environment"`
}

type runningModel struct {
	Name          string    `json:"name"`
	Model         string    `json:"model"`
	Size          int64     `json:"size"`
	SizeVRAM      int64     `json:"size_vram"`
	Digest        string    `json:"digest"`
	ExpiresAt     time.Time `json:"expires_at"`
	ContextLength int       `json:"context_length"`
}

type hostProcess struct {
	ID           int    `json:"Id"`
	Name         string `json:"Name"`
	Command      string `json:"Command"`
	WorkingSet   int64  `json:"WorkingSet"`
	Start        string `json:"Start"`
	ProviderRoot bool   `json:"ProviderRoot"`
}

type hostState struct {
	TotalRAMKiB, FreeRAMKiB, TotalSwapKiB, SwapUsedKiB int64
	Processes                                          []hostProcess
}

type snapshot struct {
	Cycle, Step                                                    int
	Label                                                          string
	Models                                                         []runningModel
	GPUName                                                        string
	GPUTotalMiB, GPUUsedMiB, GPUFreeMiB                            int64
	HostTotalRAMKiB, HostFreeRAMKiB, HostSwapUsedKiB               int64
	WSLTotalRAMKiB, WSLFreeRAMKiB, WSLTotalSwapKiB, WSLSwapUsedKiB int64
	ProviderProcesses                                              []hostProcess
	HostProcesses                                                  []hostProcess
	ResidencyPattern                                               string
}

type requestResult struct {
	Cycle, Step                                                            int
	Kind, Model, ObservedModel, Terminal, OutputSHA256, Error              string
	ClientLatencyMS, TotalDurationMS, LoadDurationMS, PromptEvalDurationMS int64
	EvalDurationMS                                                         int64
	PromptTokens, OutputTokens                                             int
	Correct, WithinLatencyLimit                                            bool
	ExplicitUnload                                                         bool
	UnloadedModel                                                          string
}

type gates struct {
	PatternsIdentical, ExpectedModelsLoaded, LatenciesWithinLimits bool
	ProviderProcessObserved                                        bool
	CrossProfileFallbacks, UnexpectedCPUOffloads, SwapGrowthEvents int
	OOMEvents, ProviderRestarts, FailedTransitions                 int
	Passed                                                         bool
}

type report struct {
	Version, Cycles                                                                 int
	ExecutedAt, ProviderVersion, ProfileSHA256, PromptSHA256, SchemaSHA256, Verdict string
	DirectChat, ControlledMutation                                                  identity
	Snapshots                                                                       []snapshot
	Requests                                                                        []requestResult
	Gates                                                                           gates
}

type apiClient struct {
	base string
	http *http.Client
}

func main() {
	profilePath := flag.String("profile", "docs/milestone-36-residency-profile-v4.yaml", "frozen profile")
	out := flag.String("output", "docs/reports/milestone-36-residency-runs-v4.json", "exclusive report")
	preflight := flag.Bool("preflight", false, "validate without changing residency")
	flag.Parse()
	if flag.NArg() != 0 {
		panic("unexpected arguments")
	}
	encoded := read(*profilePath)
	var p profile
	must(yaml.Unmarshal(encoded, &p))
	validateProfile(p)
	prompt := read("docs/prompts/mutation-host-bound-model-selection-v1.txt")
	schema := read("docs/schemas/host-bound-mutation-decision-v1.schema.json")
	client := apiClient{base: p.Provider.BaseURL, http: &http.Client{Timeout: 2 * time.Minute}}
	version := client.version()
	if version != p.Provider.Version {
		panic("provider version mismatch")
	}
	tags := client.tags()
	for _, model := range []identity{p.Models.DirectChat, p.Models.ControlledMutation} {
		if tags[model.Name] != model.Digest {
			panic("model identity mismatch: " + model.Name)
		}
	}
	baseline := capture(0, 0, "preflight")
	if baseline.GPUName != "NVIDIA GeForce RTX 5070" || baseline.GPUTotalMiB != 12227 {
		panic("reference GPU mismatch")
	}
	if baseline.WSLTotalSwapKiB != 0 || baseline.WSLSwapUsedKiB != 0 {
		panic("reference WSL swap is enabled")
	}
	if *preflight {
		fmt.Printf("PASS cycles=%d profile=%s prompt=%s schema=%s gpu=%s/%dMiB\n", p.Generation.Cycles, hash(encoded), hash(prompt), hash(schema), baseline.GPUName, baseline.GPUTotalMiB)
		return
	}
	f, err := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	must(err)
	must(f.Close())
	r := report{Version: 1, Cycles: p.Generation.Cycles, ExecutedAt: time.Now().UTC().Format(time.RFC3339), ProviderVersion: version, ProfileSHA256: hash(encoded), PromptSHA256: hash(prompt), SchemaSHA256: hash(schema), Verdict: "in_progress", DirectChat: p.Models.DirectChat, ControlledMutation: p.Models.ControlledMutation}
	write(*out, r)
	for cycle := 1; cycle <= p.Generation.Cycles; cycle++ {
		client.unload(p.Models.DirectChat.Name)
		client.unload(p.Models.ControlledMutation.Name)
		client.waitEmpty()
		r.Snapshots = append(r.Snapshots, capture(cycle, 0, "baseline"))
		steps := []struct {
			kind, limit, unload string
			model               identity
		}{
			{"direct_chat_cold", "direct_chat_cold", "", p.Models.DirectChat},
			{"direct_chat_warm", "direct_chat_warm", "", p.Models.DirectChat},
			{"chat_to_mutation", "chat_to_mutation", p.Models.DirectChat.Name, p.Models.ControlledMutation},
			{"mutation_warm", "mutation_warm", "", p.Models.ControlledMutation},
			{"mutation_to_chat", "mutation_to_chat", p.Models.ControlledMutation.Name, p.Models.DirectChat},
			{"returned_chat_warm", "returned_chat_warm", "", p.Models.DirectChat},
		}
		for index, step := range steps {
			if step.unload != "" {
				client.unload(step.unload)
				client.waitEmpty()
			}
			result := client.generate(cycle, index+1, step.kind, step.model, p, prompt, schema)
			result.ExplicitUnload = step.unload != ""
			result.UnloadedModel = step.unload
			result.WithinLatencyLimit = result.ClientLatencyMS <= int64(p.LatencyLimits[step.limit])
			r.Requests = append(r.Requests, result)
			r.Snapshots = append(r.Snapshots, capture(cycle, index+1, "after_"+step.kind))
			write(*out, r)
			fmt.Printf("M36 cycle=%d step=%s terminal=%s latency_ms=%d\n", cycle, step.kind, result.Terminal, result.ClientLatencyMS)
		}
		client.unload(p.Models.DirectChat.Name)
		client.unload(p.Models.ControlledMutation.Name)
		client.waitEmpty()
		r.Snapshots = append(r.Snapshots, capture(cycle, 7, "final_baseline"))
		write(*out, r)
	}
	r.Gates = evaluate(r)
	r.Verdict = "dual_model_residency_rejected"
	if r.Gates.Passed {
		r.Verdict = "dual_model_residency_qualified"
	}
	write(*out, r)
	fmt.Printf("verdict=%s passed=%t\n", r.Verdict, r.Gates.Passed)
}

func validateProfile(p profile) {
	if p.Version != 4 || p.Status != "swap_disabled_reference_frozen_not_run" || p.Generation.Cycles != 3 || p.Generation.ContextWindow != 4096 || p.Generation.MaxOutputTokens != 64 || p.Generation.KeepAlive != "5m" {
		panic("invalid frozen profile")
	}
	if p.Reference.WSLConfigPath != "/mnt/c/Users/cafeo/.wslconfig" || p.Reference.WSLConfigSHA256 != "f5e1a679dbeca06a712c4f098ca3d17ae3f3c2eb19d20a224904b500e23e4cd6" || hash(read(p.Reference.WSLConfigPath)) != p.Reference.WSLConfigSHA256 {
		panic("reference WSL configuration mismatch")
	}
	if p.Models.DirectChat.Name != "qwen3.5:9b" || p.Models.DirectChat.Digest != "6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7" || p.Models.ControlledMutation.Name != "qwen2.5-coder:14b" || p.Models.ControlledMutation.Digest != "9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849" {
		panic("qualified identity changed")
	}
	for _, key := range []string{"direct_chat_cold", "direct_chat_warm", "chat_to_mutation", "mutation_warm", "mutation_to_chat", "returned_chat_warm"} {
		if p.LatencyLimits[key] <= 0 {
			panic("missing latency limit: " + key)
		}
	}
}

func (c apiClient) version() string {
	var response struct{ Version string }
	c.get("/api/version", &response)
	return response.Version
}

func (c apiClient) tags() map[string]string {
	var response struct{ Models []runningModel }
	c.get("/api/tags", &response)
	result := map[string]string{}
	for _, model := range response.Models {
		name := model.Name
		if name == "" {
			name = model.Model
		}
		result[name] = model.Digest
	}
	return result
}

func (c apiClient) unload(model string) {
	var response struct {
		Done  bool   `json:"done"`
		Error string `json:"error"`
	}
	c.post("/api/generate", map[string]any{"model": model, "stream": false, "keep_alive": 0}, &response)
	if response.Error != "" || !response.Done {
		panic("unload failed: " + model)
	}
}

func (c apiClient) waitEmpty() {
	deadline := time.Now().Add(20 * time.Second)
	for {
		if len(c.running()) == 0 {
			return
		}
		if time.Now().After(deadline) {
			panic("models did not unload")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (c apiClient) running() []runningModel {
	var response struct{ Models []runningModel }
	c.get("/api/ps", &response)
	return response.Models
}

func (c apiClient) generate(cycle, step int, kind string, model identity, p profile, prompt, schema []byte) (result requestResult) {
	result = requestResult{Cycle: cycle, Step: step, Kind: kind, Model: model.Name, Terminal: "failed"}
	messages := []map[string]string{{"role": "system", "content": "Reply with exactly OK."}, {"role": "user", "content": "OK"}}
	request := map[string]any{
		"model": model.Name, "messages": messages, "stream": false, "think": false,
		"keep_alive": p.Generation.KeepAlive,
		"options":    map[string]any{"num_ctx": p.Generation.ContextWindow, "num_predict": p.Generation.MaxOutputTokens, "temperature": 0},
	}
	mutationRequest := model.Name == p.Models.ControlledMutation.Name
	if mutationRequest {
		payload, _ := json.Marshal(map[string]any{"Request": "Porta workers da 6 a 8 nella selezione.", "SelectedText": "$workers = 6;", "StartLine": 2, "EndLine": 2})
		request["messages"] = []map[string]string{{"role": "system", "content": string(prompt)}, {"role": "user", "content": string(payload)}}
		var format any
		must(json.Unmarshal(schema, &format))
		request["format"] = format
	}
	var response struct {
		Model      string `json:"model"`
		Done       bool   `json:"done"`
		DoneReason string `json:"done_reason"`
		Error      string `json:"error"`
		Message    struct {
			Content string `json:"content"`
		} `json:"message"`
		TotalDuration      int64 `json:"total_duration"`
		LoadDuration       int64 `json:"load_duration"`
		PromptEvalDuration int64 `json:"prompt_eval_duration"`
		EvalDuration       int64 `json:"eval_duration"`
		PromptEvalCount    int   `json:"prompt_eval_count"`
		EvalCount          int   `json:"eval_count"`
	}
	started := time.Now()
	err := c.postError("/api/chat", request, &response)
	result.ClientLatencyMS = time.Since(started).Milliseconds()
	result.ObservedModel = response.Model
	result.TotalDurationMS = response.TotalDuration / int64(time.Millisecond)
	result.LoadDurationMS = response.LoadDuration / int64(time.Millisecond)
	result.PromptEvalDurationMS = response.PromptEvalDuration / int64(time.Millisecond)
	result.EvalDurationMS = response.EvalDuration / int64(time.Millisecond)
	result.PromptTokens, result.OutputTokens = response.PromptEvalCount, response.EvalCount
	result.OutputSHA256 = hash([]byte(response.Message.Content))
	if err != nil || response.Error != "" {
		result.Error = strings.TrimSpace(fmt.Sprint(err, " ", response.Error))
		if strings.Contains(strings.ToLower(result.Error), "out of memory") || strings.Contains(strings.ToLower(result.Error), "oom") {
			result.Terminal = "oom"
		}
		return
	}
	if !response.Done || response.DoneReason != "stop" || response.Model != model.Name || strings.TrimSpace(response.Message.Content) == "" {
		result.Error = "invalid completion response"
		return
	}
	result.Correct = true
	if mutationRequest {
		decision, err := mutation.DecodeHostBoundDecision([]byte(response.Message.Content))
		result.Correct = err == nil && decision.Decision == mutation.BinaryPropose && decision.NewText == "$workers = 8;"
	}
	if !result.Correct {
		result.Error = "semantic response mismatch"
		return
	}
	result.Terminal = "completed"
	return
}

func (c apiClient) get(path string, destination any) {
	response, err := c.http.Get(c.base + path)
	must(err)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic("HTTP GET failed")
	}
	must(json.NewDecoder(response.Body).Decode(destination))
}

func (c apiClient) post(path string, source, destination any) {
	must(c.postError(path, source, destination))
}

func (c apiClient) postError(path string, source, destination any) error {
	encoded, err := json.Marshal(source)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("HTTP %d: %s", response.StatusCode, body)
	}
	return json.NewDecoder(response.Body).Decode(destination)
}

func capture(cycle, step int, label string) snapshot {
	client := apiClient{base: "http://127.0.0.1:11434", http: &http.Client{Timeout: 10 * time.Second}}
	s := snapshot{Cycle: cycle, Step: step, Label: label, Models: client.running()}
	sort.Slice(s.Models, func(i, j int) bool { return s.Models[i].Name < s.Models[j].Name })
	names := make([]string, len(s.Models))
	for i, model := range s.Models {
		names[i] = model.Name
	}
	s.ResidencyPattern = strings.Join(names, ",")
	gpu := command("nvidia-smi", "--query-gpu=name,memory.total,memory.used,memory.free", "--format=csv,noheader,nounits")
	parts := strings.Split(strings.TrimSpace(gpu), ",")
	if len(parts) != 4 {
		panic("invalid nvidia-smi output")
	}
	s.GPUName = strings.TrimSpace(parts[0])
	s.GPUTotalMiB = integer(parts[1])
	s.GPUUsedMiB = integer(parts[2])
	s.GPUFreeMiB = integer(parts[3])
	host := windowsHostState()
	s.HostTotalRAMKiB, s.HostFreeRAMKiB, s.HostSwapUsedKiB = host.TotalRAMKiB, host.FreeRAMKiB, host.SwapUsedKiB
	s.HostProcesses = host.Processes
	wsl := linuxHostState()
	s.WSLTotalRAMKiB, s.WSLFreeRAMKiB, s.WSLTotalSwapKiB, s.WSLSwapUsedKiB = wsl.TotalRAMKiB, wsl.FreeRAMKiB, wsl.TotalSwapKiB, wsl.SwapUsedKiB
	s.ProviderProcesses = wsl.Processes
	return s
}

func linuxHostState() hostState {
	encoded := read("/proc/meminfo")
	values := map[string]int64{}
	for _, line := range strings.Split(string(encoded), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err == nil {
			values[strings.TrimSuffix(fields[0], ":")] = value
		}
	}
	state := hostState{TotalRAMKiB: values["MemTotal"], FreeRAMKiB: values["MemAvailable"], TotalSwapKiB: values["SwapTotal"], SwapUsedKiB: values["SwapTotal"] - values["SwapFree"]}
	entries, err := os.ReadDir("/proc")
	must(err)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || !entry.IsDir() {
			continue
		}
		comm, err := os.ReadFile("/proc/" + entry.Name() + "/comm")
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		if name != "ollama" && name != "llama-server" {
			continue
		}
		status, _ := os.ReadFile("/proc/" + entry.Name() + "/status")
		var rss int64
		for _, line := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					rss = integer(fields[1]) * 1024
				}
			}
		}
		stat, _ := os.ReadFile("/proc/" + entry.Name() + "/stat")
		cmdline, _ := os.ReadFile("/proc/" + entry.Name() + "/cmdline")
		commandLine := strings.ReplaceAll(strings.TrimSpace(string(cmdline)), "\x00", " ")
		start := "unknown"
		if end := strings.LastIndex(string(stat), ") "); end >= 0 {
			fields := strings.Fields(string(stat)[end+2:])
			if len(fields) > 19 {
				start = fields[19]
			}
		}
		state.Processes = append(state.Processes, hostProcess{ID: pid, Name: name, Command: commandLine, WorkingSet: rss, Start: start, ProviderRoot: name == "ollama" && strings.Contains(commandLine, "serve")})
	}
	sort.Slice(state.Processes, func(i, j int) bool { return state.Processes[i].ID < state.Processes[j].ID })
	return state
}

func windowsHostState() hostState {
	const script = `$ErrorActionPreference='Stop'; $os=Get-CimInstance Win32_OperatingSystem; $page=@(Get-CimInstance Win32_PageFileUsage); $swap=($page | Measure-Object -Property CurrentUsage -Sum).Sum; $listener=Get-NetTCPConnection -LocalPort 11434 -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1; $ps=@(Get-Process -ErrorAction SilentlyContinue | Where-Object {$_.ProcessName -like '*ollama*'} | ForEach-Object {[PSCustomObject]@{Id=$_.Id;Name=$_.ProcessName;WorkingSet=$_.WorkingSet64;Start=$_.StartTime.ToUniversalTime().ToString('o');ProviderRoot=($listener -and $_.Id -eq $listener.OwningProcess)}}); [PSCustomObject]@{TotalRAMKiB=[int64]$os.TotalVisibleMemorySize;FreeRAMKiB=[int64]$os.FreePhysicalMemory;SwapUsedKiB=[int64]$swap*1024;Processes=$ps} | ConvertTo-Json -Compress -Depth 4`
	encoded := command("/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe", "-NoProfile", "-Command", script)
	var state hostState
	must(json.Unmarshal([]byte(strings.TrimSpace(encoded)), &state))
	return state
}

func evaluate(r report) (g gates) {
	g.PatternsIdentical, g.ExpectedModelsLoaded, g.LatenciesWithinLimits = true, true, true
	patterns := map[int][]string{}
	rootIdentity := ""
	baselineSwap := map[int]int64{}
	for _, s := range r.Snapshots {
		if s.Label == "baseline" {
			baselineSwap[s.Cycle] = s.WSLSwapUsedKiB
		}
		if s.Step >= 1 && s.Step <= 6 {
			patterns[s.Cycle] = append(patterns[s.Cycle], s.ResidencyPattern)
		}
		if s.WSLSwapUsedKiB > baselineSwap[s.Cycle] {
			g.SwapGrowthEvents++
		}
		for _, model := range s.Models {
			if model.SizeVRAM < model.Size {
				g.UnexpectedCPUOffloads++
			}
		}
		for _, process := range s.ProviderProcesses {
			if !process.ProviderRoot {
				continue
			}
			id := fmt.Sprintf("%d/%s", process.ID, process.Start)
			if rootIdentity == "" {
				rootIdentity = id
				g.ProviderProcessObserved = true
			} else if id != rootIdentity {
				g.ProviderRestarts++
			}
		}
	}
	for cycle := 2; cycle <= r.Cycles; cycle++ {
		if strings.Join(patterns[cycle], "|") != strings.Join(patterns[1], "|") {
			g.PatternsIdentical = false
		}
	}
	for index, request := range r.Requests {
		if request.Terminal != "completed" || !request.Correct {
			g.FailedTransitions++
		}
		if request.Terminal == "oom" {
			g.OOMEvents++
		}
		if request.ObservedModel != request.Model {
			g.CrossProfileFallbacks++
		}
		if !request.WithinLatencyLimit {
			g.LatenciesWithinLimits = false
		}
		snapshotIndex := index + 1 + index/6*2
		if snapshotIndex >= len(r.Snapshots) || !containsModel(r.Snapshots[snapshotIndex].Models, request.Model) {
			g.ExpectedModelsLoaded = false
		}
	}
	g.Passed = g.PatternsIdentical && g.ExpectedModelsLoaded && g.LatenciesWithinLimits && g.ProviderProcessObserved && g.CrossProfileFallbacks == 0 && g.UnexpectedCPUOffloads == 0 && g.SwapGrowthEvents == 0 && g.OOMEvents == 0 && g.ProviderRestarts == 0 && g.FailedTransitions == 0
	return
}

func containsModel(models []runningModel, name string) bool {
	for _, model := range models {
		if model.Name == name || model.Model == name {
			return true
		}
	}
	return false
}

func command(name string, arguments ...string) string {
	cmd := exec.Command(name, arguments...)
	encoded, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("command %s failed: %v: %s", name, err, encoded))
	}
	return string(encoded)
}

func integer(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	must(err)
	return parsed
}

func read(path string) []byte { data, err := os.ReadFile(path); must(err); return data }
func hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func write(path string, value any) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	must(err)
	must(os.WriteFile(path, append(encoded, '\n'), 0o600))
}
