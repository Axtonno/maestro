# Maestro Mutation Qualification Report

- Schema: `mutation-qualification-report/1.0.0`
- Run: `mutation-deterministic-a64b7557ccd24f32`
- Gate: `deterministic`
- Scenario: `mutation_matrix`
- State: **PASSED**
- Candidate: `linux_amd64/ollama`, `ibm/granite4.1:8b`
- Profile SHA-256: `a64b7557ccd24f32bb4fb7cee7d64b630e16ec017c0776b2549d86bcd8480cac`

| Attempt | State | Terminal | Turns | Calls | Mutations | Approval | Freshness | Cleanup |
|---:|---|---|---:|---:|---:|---|---|---|
| 1 | passed | completed | 0 | 0 | 1 | allowed_once | fresh | clean |
| 2 | passed | tool_failure | 0 | 0 | 0 | not_requested | not_applicable | clean |
| 3 | passed | tool_failure | 0 | 0 | 0 | not_requested | not_applicable | clean |
| 4 | passed | tool_failure | 0 | 0 | 0 | not_requested | not_applicable | clean |
| 5 | passed | permission_denied | 0 | 0 | 0 | denied | not_applicable | clean |
| 6 | passed | permission_denied | 0 | 0 | 0 | denied | not_applicable | clean |
| 7 | passed | permission_denied | 0 | 0 | 0 | unavailable | not_applicable | clean |
| 8 | passed | permission_denied | 0 | 0 | 0 | denied | not_applicable | clean |
| 9 | passed | canceled | 0 | 0 | 1 | allowed_once | stale | clean |
| 10 | passed | canceled | 0 | 0 | 1 | allowed_once | stale | clean |
| 11 | passed | tool_failure | 0 | 0 | 1 | allowed_once | stale | clean |
| 12 | passed | tool_failure | 0 | 0 | 1 | allowed_once | stale | clean |
| 13 | passed | tool_failure | 0 | 0 | 0 | not_requested | not_applicable | clean |
| 14 | passed | tool_failure | 0 | 0 | 1 | allowed_once | stale | clean |
| 15 | passed | tool_failure | 0 | 0 | 1 | allowed_once | stale | clean |

## Workspace evidence

- Attempt 1 (`positive_exact_patch`, `workspace_tools_positive_patch`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d`, workspace `2f34c1b466ff81ed00c1cf1d9f23517a1a42cc012989a153f9820b8e2c1980e5`.
- Attempt 2 (`stale_digest`, `workspace_patch_precondition`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 3 (`traversal`, `workspace_tools_containment`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 4 (`symlink`, `workspace_tools_containment`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 5 (`user_deny`, `terminal_approver_exact_patch`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 6 (`approval_eof`, `terminal_approver_input_unavailable`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 7 (`approval_no_tty`, `terminal_approver_non_interactive`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 8 (`approval_invalid_input`, `terminal_approver_input_invalid`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 9 (`cancellation_before_commit`, `atomic_replace_precommit_faults`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 10 (`cancellation_after_commit`, `atomic_replace_postcommit_cancel`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d`, workspace `2f34c1b466ff81ed00c1cf1d9f23517a1a42cc012989a153f9820b8e2c1980e5`.
- Attempt 11 (`filesystem_fault_before_commit`, `atomic_replace_fault_matrix`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 12 (`refresh_failure_after_commit`, `agent_refresh_failure`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d`, workspace `2f34c1b466ff81ed00c1cf1d9f23517a1a42cc012989a153f9820b8e2c1980e5`.
- Attempt 13 (`undeclared_tool`, `agent_unknown_tool`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 14 (`approval_replay`, `permit_one_shot_replay`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, workspace `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`.
- Attempt 15 (`second_mutation_attempt`, `agent_second_mutation_rejected`): initial `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`, final `509b566bd04a17d567248a721885ac5af0d623f9f505288548c7c302628bac5d`, workspace `2f34c1b466ff81ed00c1cf1d9f23517a1a42cc012989a153f9820b8e2c1980e5`.
