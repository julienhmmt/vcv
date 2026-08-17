#!/usr/bin/env python3
"""Query local SonarQube results from the command line.

Designed for LLM agents: outputs structured, concise summaries that fit
in a tool response without overwhelming the context window.

Usage:
    python3 tools/sonar-query.py projects
    python3 tools/sonar-query.py summary
    python3 tools/sonar-query.py issues <project-key> [--severity CRITICAL] [--limit 10]
    python3 tools/sonar-query.py metrics <project-key>
    python3 tools/sonar-query.py gate <project-key>
    python3 tools/sonar-query.py file <project-key> <file-path>

Credentials are read from .sonar/admin-password (created by make sonar-bootstrap).
Override the URL with SONAR_URL env var (default http://localhost:9000).
"""
import argparse
import json
import os
import sys
import urllib.request
import urllib.parse
import base64


SONAR_URL = os.environ.get("SONAR_URL", "http://localhost:9000")
PASS_FILE = ".sonar/admin-password"
ADMIN_USER = "admin"


def get_password():
    if os.environ.get("SONAR_ADMIN_PASSWORD"):
        return os.environ["SONAR_ADMIN_PASSWORD"]
    if os.path.exists(PASS_FILE):
        with open(PASS_FILE) as f:
            return f.read().strip()
    print("Error: no admin password found. Run 'make sonar-bootstrap' first.", file=sys.stderr)
    sys.exit(1)


def api_get(path, params=None):
    url = f"{SONAR_URL}/api/{path}"
    if params:
        url += "?" + urllib.parse.urlencode(params)
    password = get_password()
    credentials = base64.b64encode(f"{ADMIN_USER}:{password}".encode()).decode()
    req = urllib.request.Request(url, headers={"Authorization": f"Basic {credentials}"})
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        print(f"HTTP {e.code}: {e.read().decode()[:200]}", file=sys.stderr)
        sys.exit(1)
    except urllib.error.URLError as e:
        print(f"Connection error: {e.reason}", file=sys.stderr)
        print("Is SonarQube running? Try 'make sonar-up'.", file=sys.stderr)
        sys.exit(1)


def cmd_projects(_args):
    data = api_get("projects/search", {"ps": 100})
    components = data.get("components", [])
    print(f"SonarQube projects ({len(components)}):")
    for c in sorted(components, key=lambda x: x["key"]):
        print(f"  {c['key']:25s}  {c['name']}")


def cmd_summary(_args):
    data = api_get("projects/search", {"ps": 100})
    components = data.get("components", [])
    print(f"{'PROJECT':25s} {'GATE':8s} {'BUGS':6s} {'VULN':6s} {'SMELLS':7s} {'COV':6s} {'FILES':6s} {'NCLOC':7s}")
    print("-" * 80)
    for c in sorted(components, key=lambda x: x["key"]):
        key = c["key"]
        gate = api_get("qualitygates/project_status", {"projectKey": key})
        gate_status = gate.get("projectStatus", {}).get("status", "?")
        metrics = api_get("measures/component", {
            "component": key,
            "metricKeys": "bugs,vulnerabilities,code_smells,coverage,files,ncloc",
        })
        m = {x["metric"]: x["value"] for x in metrics.get("component", {}).get("measures", [])}
        print(f"{key:25s} {gate_status:8s} {m.get('bugs','?'):6s} {m.get('vulnerabilities','?'):6s} "
              f"{m.get('code_smells','?'):7s} {m.get('coverage','?'):6s} {m.get('files','?'):6s} {m.get('ncloc','?'):7s}")


def cmd_issues(args):
    params = {
        "componentKeys": args.project,
        "resolved": "false",
        "ps": args.limit,
    }
    if args.severity:
        params["severities"] = args.severity
    if args.rule:
        params["rules"] = args.rule
    data = api_get("issues/search", params)
    total = data.get("total", 0)
    issues = data.get("issues", [])
    print(f"{args.project}: {total} open issues (showing {len(issues)})")
    for i in issues:
        comp = i.get("component", "").split(":")[-1]
        line = i.get("line", "?")
        sev = i.get("severity", "?")
        rule = i.get("rule", "?")
        msg = i.get("message", "")[:100]
        print(f"  [{sev}] {rule} → {comp}:{line}")
        print(f"         {msg}")


def cmd_metrics(args):
    data = api_get("measures/component", {
        "component": args.project,
        "metricKeys": "bugs,vulnerabilities,code_smells,coverage,security_hotspots,files,ncloc,duplicated_lines_density,sqale_rating,reliability_rating,security_rating",
    })
    m = {x["metric"]: x["value"] for x in data.get("component", {}).get("measures", [])}
    print(f"Metrics for {args.project}:")
    for k in ["files", "ncloc", "coverage", "bugs", "vulnerabilities", "code_smells",
              "security_hotspots", "duplicated_lines_density", "sqale_rating",
              "reliability_rating", "security_rating"]:
        if k in m:
            print(f"  {k:30s} {m[k]}")


def cmd_gate(args):
    data = api_get("qualitygates/project_status", {"projectKey": args.project})
    ps = data.get("projectStatus", {})
    print(f"Quality gate for {args.project}: {ps.get('status', '?')}")
    for cond in ps.get("conditions", []):
        print(f"  {cond.get('metricKey','?'):25s} {cond.get('status','?'):8s} "
              f"{cond.get('actual','?')} (op={cond.get('operator','?')}, threshold={cond.get('errorThreshold','?')})")


def cmd_file(args):
    params = {
        "component": f"{args.project}:{args.file_path}",
        "metricKeys": "bugs,vulnerabilities,code_smells,coverage,ncloc",
    }
    try:
        data = api_get("measures/component", params)
        m = {x["metric"]: x["value"] for x in data.get("component", {}).get("measures", [])}
        print(f"Metrics for {args.project}:{args.file_path}")
        for k, v in m.items():
            print(f"  {k}: {v}")
    except SystemExit:
        # Try issues for this file
        pass
    # Also show issues for this file
    data = api_get("issues/search", {
        "componentKeys": args.project,
        "resolved": "false",
        "ps": 50,
    })
    file_issues = [i for i in data.get("issues", []) if args.file_path in i.get("component", "")]
    if file_issues:
        print(f"\nIssues in {args.file_path} ({len(file_issues)}):")
        for i in file_issues:
            sev = i.get("severity", "?")
            rule = i.get("rule", "?")
            line = i.get("line", "?")
            msg = i.get("message", "")[:100]
            print(f"  [{sev}] {rule} line {line}: {msg}")
    elif not file_issues:
        print(f"\nNo issues found in {args.file_path}")


def main():
    parser = argparse.ArgumentParser(description="Query local SonarQube results.")
    sub = parser.add_subparsers(dest="command", required=True)

    sub.add_parser("projects", help="List all SonarQube projects").set_defaults(func=cmd_projects)
    sub.add_parser("summary", help="Print summary table of all projects").set_defaults(func=cmd_summary)

    p_issues = sub.add_parser("issues", help="List open issues for a project")
    p_issues.add_argument("project", help="Project key (e.g. pvmss-server)")
    p_issues.add_argument("--severity", help="Filter by severity (CRITICAL, MAJOR, MINOR, INFO)")
    p_issues.add_argument("--rule", help="Filter by rule ID (e.g. go:S1192)")
    p_issues.add_argument("--limit", type=int, default=20, help="Max issues to show (default 20)")
    p_issues.set_defaults(func=cmd_issues)

    p_metrics = sub.add_parser("metrics", help="Show metrics for a project")
    p_metrics.add_argument("project", help="Project key")
    p_metrics.set_defaults(func=cmd_metrics)

    p_gate = sub.add_parser("gate", help="Show quality gate status for a project")
    p_gate.add_argument("project", help="Project key")
    p_gate.set_defaults(func=cmd_gate)

    p_file = sub.add_parser("file", help="Show issues and metrics for a specific file")
    p_file.add_argument("project", help="Project key")
    p_file.add_argument("file_path", help="File path relative to project (e.g. server/internal/httpapi/auth.go)")
    p_file.set_defaults(func=cmd_file)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
