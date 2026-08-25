# Playbook — Own the red SonarCloud gate until it is green

CI is red on a pull request because SonarCloud's quality gate failed. You own
that gate end to end: your job is not "fix the findings", it is **the gate is
green on this pull request and every failing cause has an explicit owner and a
written verdict**. You are done only when the gate reads `OK`.

Everything happens inside that pull request: fixes are committed on its own
branch, verdicts are comments on it. Never open another pull request, never
create another branch, never target the base branch.

## 1. Read the whole failure, not just the findings

Two different things can turn the gate red, and both are yours:

- **Findings** — issues and security hotspots Sonar raised on the new code
  (`api/issues/search` with `resolved=false`, `api/hotspots/search` with
  `status=TO_REVIEW`, both with `pullRequest=<number>`).
- **Gate conditions that are not findings** — coverage on new code, duplicated
  lines, ratings (`api/qualitygates/project_status?projectKey=...&pullRequest=<number>`).
  Nobody raises an issue for these, so they are the ones that silently keep CI
  red. Read every condition with `status=ERROR`, its threshold and its actual
  value.

Prefer the SonarQube MCP server when it is available to you
(`search_sonar_issues_in_projects`, `search_security_hotspots`,
`get_project_quality_gate_status`); fall back to the public API otherwise. Write
the full list — findings plus failing conditions — as your work plan before you
change anything.

## 2. One owner per cause

For **each finding**, create a child Devin session with the triage playbook
whose ID is given at the end of this prompt (Devin MCP: propose a session with
that `playbook_id`), passing it the repository, the pull request URL, the branch
to commit on, and the finding's key, rule, location and message. One session per
finding, so each verdict (real vulnerability vs false positive) is argued
separately.

For **each failing condition that is not a finding**, own it yourself or create
one child session for it, and fix the cause:

- `new_coverage` below the threshold: add real tests for the new code that is
  uncovered — find it with
  `go test -coverprofile=coverage.out -covermode=atomic ./... && go tool cover -func=coverage.out`,
  and cover the new handlers and helpers, exercising their behaviour (status
  codes, rendered content, error paths), not just calling them.
- `new_duplicated_lines_density`: factor out the duplicated block.
- ratings: they are driven by the findings above; recheck after those land.

Never make a condition pass by weakening the check: do not lower a threshold,
do not change the quality gate, do not add exclusions to
`sonar-project.properties`, and do not delete or neuter the planted
vulnerabilities to make numbers look better.

If creating child sessions is not available to you (no Devin MCP, or the
proposal is not approved within a few minutes), do the work yourself in this
session, finding by finding, with the same triage discipline as the playbook —
never leave a cause unowned.

## 3. Converge

The gate is a moving target: every fix that lands changes coverage and ratings.
So after the sessions report back:

1. `git pull --rebase` the pull request's branch and confirm every expected fix
   and test is in it. Children push to the same branch concurrently — retry the
   rebase-and-push if a push is rejected as non-fast-forward.
2. Wait for the new SonarCloud run on that branch, then re-read the quality gate
   for the pull request.
3. For every condition still in `ERROR`, go back to step 2 with a new owner.
4. If a false positive is keeping a rating red, make sure the finding is closed
   in SonarCloud (SonarQube MCP: `change_sonar_issue_status` with
   `status=falsepositive`, or `accept` when the risk is real but accepted) with
   the justification already written in the pull request comment.

Repeat until `projectStatus.status` is `OK` and the checks on the pull request
pass.

Before you call it done, verify the app still works with everything that landed:
`go build ./...`, `go test ./...`, then `go run ./cmd/api` and walk the UI in the
browser (home, search, contract detail, the page the pull request added), with a
screen recording of that walkthrough as proof the frontend is intact.

## 4. Report

Comment once on the pull request with the final state: every finding with its
verdict (real → what was fixed; false positive → why it is not exploitable and
that it was closed in Sonar), every gate condition that was failing with what
brought it to green, and the gate's final status. Link the child session of each
cause, and end the comment with a link to your own session
(`[Devin session](<your session URL>)`).

Before reporting a green gate, verify it yourself against Sonar — do not infer
it from your own changes.
