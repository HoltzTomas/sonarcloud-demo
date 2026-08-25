#!/usr/bin/env python3
"""Create one Devin session to own a failing SonarCloud quality gate.

Reads the open issues, security hotspots, and quality-gate conditions SonarCloud
reported for a pull request (or a branch), then creates one gate-owner Devin
session. That session uses the gate-owner playbook to coordinate child triage
sessions and owns the gate until it is green, committing fixes on the pull
request's own branch.

Environment:
  SONAR_TOKEN        SonarCloud token with "Browse" on the project
  SONAR_HOST_URL     defaults to https://sonarcloud.io
  SONAR_PROJECT_KEY  e.g. HoltzTomas_sonarcloud-repsol-demo
  DEVIN_API_KEY      Devin API key of the org that owns the repo
  DEVIN_API_BASE     defaults to https://api.devin.ai
  DEVIN_PLAYBOOK_ID       saved gate-owner playbook; when unset, the gate-owner
                          playbook markdown in this repo is inlined instead
  DEVIN_TRIAGE_PLAYBOOK_ID saved triage playbook ID for child sessions; passed
                           through in the gate-owner prompt
  GITHUB_REPOSITORY       owner/repo, provided by GitHub Actions
  PR_NUMBER               pull request number (omit to scan the branch)
  PR_BRANCH               head branch of that pull request, where fixes are committed
  REMEDIATION_BRANCH      fallback branch for branch scans, defaults to sonar/remediation
"""

import json
import os
import pathlib
import sys
import urllib.parse
import urllib.request

SONAR_HOST = os.environ.get("SONAR_HOST_URL", "https://sonarcloud.io").rstrip("/")
SONAR_TOKEN = os.environ["SONAR_TOKEN"]
PROJECT_KEY = os.environ["SONAR_PROJECT_KEY"]
DEVIN_API_BASE = os.environ.get("DEVIN_API_BASE", "https://api.devin.ai").rstrip("/")
DEVIN_API_KEY = os.environ["DEVIN_API_KEY"]
DEVIN_PLAYBOOK_ID = os.environ.get("DEVIN_PLAYBOOK_ID", "").strip()
DEVIN_TRIAGE_PLAYBOOK_ID = os.environ.get("DEVIN_TRIAGE_PLAYBOOK_ID", "").strip()
REPO = os.environ["GITHUB_REPOSITORY"]
PR_NUMBER = os.environ.get("PR_NUMBER", "").strip()
PR_BRANCH = os.environ.get("PR_BRANCH", "").strip()
REMEDIATION_BRANCH = os.environ.get("REMEDIATION_BRANCH", "sonar/remediation")
TARGET_BRANCH = PR_BRANCH or REMEDIATION_BRANCH

PLAYBOOK = pathlib.Path(__file__).resolve().parents[1] / "playbooks" / "sonar-gate-owner.md"


def sonar_get(path, **params):
    if PR_NUMBER:
        params["pullRequest"] = PR_NUMBER
    url = f"{SONAR_HOST}{path}?{urllib.parse.urlencode(params)}"
    req = urllib.request.Request(url)
    req.add_header("Authorization", f"Bearer {SONAR_TOKEN}")
    with urllib.request.urlopen(req) as resp:
        return json.load(resp)


def collect_context():
    issues = sonar_get(
        "/api/issues/search",
        componentKeys=PROJECT_KEY,
        resolved="false",
        types="VULNERABILITY,BUG",
        ps=100,
    ).get("issues", [])

    out = []
    for issue in issues:
        out.append(
            {
                "key": issue["key"],
                "kind": issue["type"],
                "severity": issue.get("severity", "UNKNOWN"),
                "rule": issue["rule"],
                "message": issue["message"],
                "component": issue["component"].split(":", 1)[-1],
                "line": issue.get("line"),
                "url": f"{SONAR_HOST}/project/issues?id={PROJECT_KEY}&open={issue['key']}",
            }
        )

    hotspots = sonar_get(
        "/api/hotspots/search",
        projectKey=PROJECT_KEY,
        status="TO_REVIEW",
        ps=100,
    ).get("hotspots", [])
    for hs in hotspots:
        out.append(
            {
                "key": hs["key"],
                "kind": "SECURITY_HOTSPOT",
                "severity": hs.get("vulnerabilityProbability", "UNKNOWN"),
                "rule": hs["ruleKey"],
                "message": hs["message"],
                "component": hs["component"].split(":", 1)[-1],
                "line": hs.get("line"),
                "url": f"{SONAR_HOST}/security_hotspots?id={PROJECT_KEY}&hotspots={hs['key']}",
            }
        )
    quality_gate = sonar_get(
        "/api/qualitygates/project_status",
        projectKey=PROJECT_KEY,
    ).get("projectStatus", {})
    gate_errors = []
    for condition in quality_gate.get("conditions", []):
        if condition.get("status") == "ERROR":
            gate_errors.append(
                {
                    "metric": condition.get("metricKey", "UNKNOWN"),
                    "comparator": condition.get("comparator", "UNKNOWN"),
                    "threshold": condition.get("errorThreshold", "UNKNOWN"),
                    "actual": condition.get("actualValue", "UNKNOWN"),
                }
            )
    return out, gate_errors


def location(finding):
    line = finding["line"] if finding["line"] is not None else "?"
    return f"{finding['component']}:{line}"


def findings_table(found):
    if not found:
        return "_No open findings._"
    rows = [
        "| Key | Type / severity | Rule | File:line | Message | Sonar link |",
        "| --- | --- | --- | --- | --- | --- |",
    ]
    for finding in found:
        rows.append(
            f"| `{finding['key']}` | `{finding['kind']} / {finding['severity']}` "
            f"| `{finding['rule']}` | `{location(finding)}` "
            f"| {finding['message']} | {finding['url']} |"
        )
    return "\n".join(rows)


def gate_errors_list(gate_errors):
    if not gate_errors:
        return "_No quality-gate conditions are in ERROR._"
    return "\n".join(
        f"- `{error['metric']}` `{error['comparator']}` "
        f"threshold `{error['threshold']}`, actual `{error['actual']}`"
        for error in gate_errors
    )


def prompt_for(found, gate_errors):
    preamble = "" if DEVIN_PLAYBOOK_ID else PLAYBOOK.read_text() + "\n\n"
    pr_url = f"https://github.com/{REPO}/pull/{PR_NUMBER}" if PR_NUMBER else "(branch scan)"
    pr_label = f"PR #{PR_NUMBER}" if PR_NUMBER else f"branch {TARGET_BRANCH}"
    triage_id = DEVIN_TRIAGE_PLAYBOOK_ID or "(not configured)"
    return (
        f"{preamble}"
        "# SonarCloud gate-owner assignment\n"
        f"- Repository: {REPO}\n"
        f"- Pull request: {pr_url}\n"
        f"- Pull request number: {PR_NUMBER or '(none; branch scan)'}\n"
        f"- Branch to commit on (do NOT open a new pull request): {TARGET_BRANCH}\n"
        f"- Use Devin triage playbook ID `{triage_id}` for every child finding session.\n"
        f"- You own {pr_label} until the SonarCloud quality gate is green.\n\n"
        "## Open Sonar findings\n"
        f"{findings_table(found)}\n\n"
        "## Quality-gate conditions in ERROR\n"
        f"{gate_errors_list(gate_errors)}\n"
    )


def create_session(found, gate_errors):
    pr_label = f"PR #{PR_NUMBER}" if PR_NUMBER else f"branch {TARGET_BRANCH}"
    payload = {
        "prompt": prompt_for(found, gate_errors),
        "title": f"Sonar gate owner — {pr_label}",
        "tags": ["sonar-remediation", "gate-owner"],
        "idempotent": True,
    }
    if DEVIN_PLAYBOOK_ID:
        payload["playbook_id"] = DEVIN_PLAYBOOK_ID
    req = urllib.request.Request(
        f"{DEVIN_API_BASE}/v1/sessions",
        data=json.dumps(payload).encode(),
        method="POST",
    )
    req.add_header("Authorization", f"Bearer {DEVIN_API_KEY}")
    req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req) as resp:
        return json.load(resp)


def write_summary(found, gate_errors, session_url):
    summary = os.environ.get("GITHUB_STEP_SUMMARY")
    if not summary:
        return
    lines = [
        "## SonarCloud gate owner",
        "",
        "### Open findings",
        findings_table(found),
        "",
        "### Quality-gate conditions in ERROR",
        gate_errors_list(gate_errors),
        "",
        f"### Orchestrator session\n{session_url}",
        "",
    ]
    with open(summary, "a") as fh:
        fh.write("\n".join(lines))


def main():
    found, gate_errors = collect_context()
    if not found and not gate_errors:
        print("No open Sonar findings and the quality gate is not in ERROR; nothing to do.")
        return

    print(
        f"{len(found)} open finding(s), "
        f"{len(gate_errors)} quality-gate condition(s) in ERROR"
    )
    session = create_session(found, gate_errors)
    session_url = session.get("url", session.get("session_id", "created"))
    print(f"Gate-owner Devin session -> {session_url}")
    write_summary(found, gate_errors, session_url)


if __name__ == "__main__":
    sys.exit(main())
