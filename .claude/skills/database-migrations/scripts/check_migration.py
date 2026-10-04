#!/usr/bin/env python3
"""Flag risky statements in SQL migration files (PostgreSQL-oriented).
Usage: check_migration.py file1.sql [file2.sql ...]
Exit code 1 if any HIGH risk found."""
import re
import sys

RULES = [
    ("HIGH", r"\bDROP\s+(TABLE|COLUMN|SCHEMA|DATABASE)\b", "Destructive drop — confirm code no longer uses it and a backup exists."),
    ("HIGH", r"\bTRUNCATE\b", "TRUNCATE deletes all rows."),
    ("HIGH", r"\bDELETE\s+FROM\s+\w+\s*;", "DELETE without WHERE."),
    ("HIGH", r"\bUPDATE\s+\w+\s+SET\b(?![^;]*\bWHERE\b)[^;]*;", "UPDATE without WHERE."),
    ("HIGH", r"\bRENAME\s+(COLUMN|TO)\b", "Rename breaks running code — use expand/contract."),
    ("HIGH", r"\bALTER\s+COLUMN\s+\w+\s+(SET\s+DATA\s+)?TYPE\b", "Type change may rewrite table and take heavy locks."),
    ("MEDIUM", r"\bCREATE\s+(UNIQUE\s+)?INDEX\s+(?!CONCURRENTLY)", "Index without CONCURRENTLY locks writes on large tables."),
    ("MEDIUM", r"\bSET\s+NOT\s+NULL\b", "SET NOT NULL scans table; use CHECK NOT VALID + VALIDATE first on big tables."),
    ("MEDIUM", r"\bADD\s+(CONSTRAINT\s+\w+\s+)?FOREIGN\s+KEY\b(?![^;]*NOT\s+VALID)", "FK without NOT VALID validates immediately under lock."),
    ("MEDIUM", r"\bFLOAT\b|\bREAL\b|\bDOUBLE\s+PRECISION\b", "Floating point type — never use for money."),
    ("LOW", r"\bTIMESTAMP\b(?!\s*WITH\s+TIME\s+ZONE|TZ)", "Prefer TIMESTAMPTZ."),
    ("LOW", r"\bCASCADE\b", "CASCADE present — confirm intended blast radius."),
]

def main(paths):
    worst = 0
    order = {"LOW": 1, "MEDIUM": 2, "HIGH": 3}
    for p in paths:
        sql = open(p, encoding="utf-8").read()
        # strip comments
        clean = re.sub(r"--[^\n]*", "", sql)
        clean = re.sub(r"/\*.*?\*/", "", clean, flags=re.S)
        findings = []
        for sev, pat, msg in RULES:
            for m in re.finditer(pat, clean, flags=re.I):
                line = clean[: m.start()].count("\n") + 1
                findings.append((sev, line, msg))
                worst = max(worst, order[sev])
        has_lock_timeout = re.search(r"lock_timeout", clean, re.I)
        print(f"\n== {p} ==")
        if not findings:
            print("  No risky patterns found.")
        for sev, line, msg in sorted(findings, key=lambda f: -order[f[0]]):
            print(f"  [{sev}] line ~{line}: {msg}")
        if findings and not has_lock_timeout:
            print("  [INFO] Consider `SET lock_timeout = '5s';` at the top.")
    return 1 if worst == 3 else 0

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(2)
    sys.exit(main(sys.argv[1:]))
