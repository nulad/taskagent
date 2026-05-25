# TASK-009: Implement API Request Error Mapping

## Token Safety Verification

- **Task ID & Title:** TASK-009: Implement API Request Error Mapping
- **Objective:** Map HTTP and network failures from `api_request` to stable CLI exit codes and stderr messages.
- **Target Files/Scope:** `cli/lib/common.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this changes one helper function in one small module.
  * Dependencies: Successful request wrapper from TASK-008; JSON error contract from the HTTP API server.

## Step-by-step Instructions

1. Extend `api_request()` to handle non-2xx responses.
2. On 4xx responses, extract `.error` from a JSON body with `jq -r`, print it to stderr, and return exit code 3.
3. On 5xx responses, extract `.error` when available, otherwise print the raw body to stderr, and return exit code 4.
4. On curl network errors, print `network error: <message>` to stderr and return exit code 5.
5. Keep stdout empty for all error paths.
6. Do not use `curl -f`, because the response body is needed for server errors.
7. Preserve 2xx behavior from TASK-008.
8. Ensure malformed JSON error bodies do not crash the shell script with confusing `jq` output.

## Definition of Done

- 4xx API errors exit 3.
- 5xx API errors exit 4.
- Network failures exit 5.
- Error text is written to stderr only.

