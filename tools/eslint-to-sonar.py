#!/usr/bin/env python3
"""Convert ESLint JSON output to SonarQube generic issue format.

Usage: python3 eslint-to-sonar.py <input.json> <output.json> <path_prefix>
"""
import json
import sys


SEVERITY_MAP = {
    1: "MINOR",
    2: "CRITICAL",
}

CLEAN_CODE_MAP = {
    "no-unused-vars": "CLEAR",
    "no-undef": "COMPLETE",
    "@typescript-eslint/no-explicit-any": "DISTINCT",
    "@typescript-eslint/no-unused-vars": "CLEAR",
}


def main():
    in_path, out_path, prefix = sys.argv[1], sys.argv[2], sys.argv[3]

    with open(in_path) as f:
        eslint_data = json.load(f)

    rules = {}
    issues = []

    for file_result in eslint_data:
        file_path = file_result.get("filePath", "")
        if not file_path:
            continue
        rel_path = file_path
        if file_path.startswith(prefix):
            rel_path = file_path[len(prefix):]
        rel_path = rel_path.lstrip("/")

        for msg in file_result.get("messages", []):
            rule_id = msg.get("ruleId") or "eslint-unknown"
            if rule_id not in rules:
                sev = SEVERITY_MAP.get(msg.get("severity", 1), "MINOR")
                rules[rule_id] = {
                    "id": rule_id,
                    "name": rule_id,
                    "description": f"ESLint rule: {rule_id}",
                    "engineId": "eslint",
                    "cleanCodeAttribute": CLEAN_CODE_MAP.get(rule_id, "FORMATTED"),
                    "type": "CODE_SMELL",
                    "severity": sev,
                    "impacts": [
                        {
                            "softwareQuality": "MAINTAINABILITY",
                            "severity": "MEDIUM" if sev == "MINOR" else "HIGH",
                        }
                    ],
                }

            start_line = msg.get("line", 1)
            end_line = msg.get("endLine") or start_line
            # SonarQube textRange columns are 0-based; ESLint's are 1-based.
            start_col = max(0, msg.get("column", 1) - 1)
            end_col = max(0, (msg.get("endColumn") or (msg.get("column", 1) + 1)) - 1)
            if end_col <= start_col:
                end_col = start_col + 1

            issues.append(
                {
                    "ruleId": rule_id,
                    "engineId": "eslint",
                    "primaryLocation": {
                        "message": msg.get("message", "ESLint issue"),
                        "filePath": rel_path,
                        "textRange": {
                            "startLine": start_line,
                            "endLine": end_line,
                            "startColumn": start_col,
                            "endColumn": end_col,
                        },
                    },
                }
            )

    report = {"rules": list(rules.values()), "issues": issues}
    with open(out_path, "w") as f:
        json.dump(report, f, indent=2)

    print(f"  {len(issues)} issues from {len(rules)} rules → {out_path}")


if __name__ == "__main__":
    main()
