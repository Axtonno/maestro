# Maestro Benchmark Report

Generated from benchmark JSON schema `1.2.0`.

## Run

| Field | Value |
|---|---|
| Run ID | abd5f5112a80557a2766d1328a03fb32 |
| Command | bench smoke |
| Created | 2026-08-09T15:30:34.964754326Z |
| Completed | 2026-08-09T15:31:13.750946939Z |
| Duration | 38786.192618 ms |
| Manifest | v1 — milestone-3-smoke-benchmark |
| Maestro version | — |
| Maestro commit | — |

## Configuration

| Field | Value |
|---|---|
| Operating system | linux |
| Architecture | amd64 |
| CPU | Intel(R) Core(TM) i5-8365U CPU @ 1.60GHz |
| Logical CPUs | 8 |
| Memory | 15643 MiB |
| GPU | — |
| Backend | — |
| VRAM | — |
| Provider | ollama |
| Provider version | — |
| Endpoint | http://localhost:11434 |
| Primary model | — |
| Dataset | — |
| Warmup / runs | 0 / 1 |

### Models by role

| Role | Model | Format | Quantization | Context |
|---|---|---|---|---:|
| chat | qwen2.5-coder:7b | — | — | — |
| embedding | embeddinggemma | — | — | — |

## Scenario summary

| Scenario | State | Measured samples | Quality |
|---|---:|---:|---:|
| capability-instance | passed | 1 | — |
| catalog-list-discover | passed | 1 | — |
| completion-simple | passed | 1 | — |
| stream-terminal-close | passed | 1 | — |
| stream-cancel-deadline | passed | 1 | — |
| embedding | skipped | 1 | — |
| lifecycle-load-unload | skipped | 1 | — |
| acquisition-pull-remove | skipped | 1 | — |
| structured-json | passed | 1 | — |
| structured-json-schema | passed | 1 | — |
| tool-call-result | failed | 1 | — |
| tool-call-stream | failed | 1 | — |
| resilience-controlled-error | passed | 1 | — |
| observability-redaction | passed | 1 | — |

## Scenario 1

| Field | Value |
|---|---|
| ID | capability-instance |
| Capability | capability_introspection |
| Model role | none |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| capability_count | — | 1 | 11 | 11 | — | 11 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 1.341558 ms | ok | — |

## Scenario 2

| Field | Value |
|---|---|
| ID | catalog-list-discover |
| Capability | model_discovery |
| Model role | none |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| discovered_model_count | — | 1 | 2 | 2 | — | 2 | count |
| listed_model_count | — | 1 | 2 | 2 | — | 2 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 2.360574 ms | ok | — |

## Scenario 3

| Field | Value |
|---|---|
| ID | completion-simple |
| Capability | completion |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| input_tokens | — | 1 | 40 | 40 | — | 40 | count |
| output_tokens | — | 1 | 10 | 10 | — | 10 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 4133.602512 ms | ok | — |

## Scenario 4

| Field | Value |
|---|---|
| ID | stream-terminal-close |
| Capability | streaming |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| stream_chunk_count | — | 1 | 10 | 10 | — | 10 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 2897.808725 ms | ok | — |

## Scenario 5

| Field | Value |
|---|---|
| ID | stream-cancel-deadline |
| Capability | streaming |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| chunks_before_cancellation | — | 1 | 0 | 0 | — | 0 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 2291.957566 ms | ok | — |

## Scenario 6

| Field | Value |
|---|---|
| ID | embedding |
| Capability | embedding |
| Model role | embedding |
| State | skipped |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | skipped | 1.220784 ms | capability_unavailable | — |

## Scenario 7

| Field | Value |
|---|---|
| ID | lifecycle-load-unload |
| Capability | model_load |
| Model role | lifecycle |
| State | skipped |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | skipped | 0.001311 ms | model_not_configured | — |

## Scenario 8

| Field | Value |
|---|---|
| ID | acquisition-pull-remove |
| Capability | model_pull |
| Model role | acquisition_fixture |
| State | skipped |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | skipped | 0.000927 ms | catalog_mutation_not_allowed | — |

## Scenario 9

| Field | Value |
|---|---|
| ID | structured-json |
| Capability | structured_output |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| json_field_count | — | 1 | 1 | 1 | — | 1 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 4022.218292 ms | ok | — |

## Scenario 10

| Field | Value |
|---|---|
| ID | structured-json-schema |
| Capability | structured_output |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| json_field_count | — | 1 | 1 | 1 | — | 1 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 2917.605685 ms | ok | — |

## Scenario 11

| Field | Value |
|---|---|
| ID | tool-call-result |
| Capability | tool_calling |
| Model role | chat |
| State | failed |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | failed | 6999.779088 ms | smoke_gate:tool_call_missing | — |

## Scenario 12

| Field | Value |
|---|---|
| ID | tool-call-stream |
| Capability | tool_calling |
| Model role | chat |
| State | failed |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | failed | 6831.205197 ms | smoke_gate:tool_stream_terminal_missing | — |

## Scenario 13

| Field | Value |
|---|---|
| ID | resilience-controlled-error |
| Capability | completion |
| Model role | chat |
| State | passed |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 3.250104 ms | ok | — |

## Scenario 14

| Field | Value |
|---|---|
| ID | observability-redaction |
| Capability | completion |
| Model role | chat |
| State | passed |

### Aggregates

| Metric | Scope | Count | Min | Median | P95 | Max | Unit |
|---|---|---:|---:|---:|---:|---:|---|
| provider_event_count | — | 1 | 3 | 3 | — | 3 | count |

### Samples

| Iteration | Type | State | Duration | Result | Quality |
|---:|---|---|---:|---|---|
| 1 | measured | passed | 8683.396365 ms | ok | — |

