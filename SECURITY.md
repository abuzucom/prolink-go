# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| `main` | Yes |
| Any earlier commit | No |

The repository carries no release tags. Importers resolve a Go pseudo-version
from a commit. Only the current `main` branch receives fixes.

## Scope

Scope includes code in this repository. Scope includes configuration in this
repository. Scope includes third-party vulnerabilities when a selected version
causes the defect. Scope includes third-party vulnerabilities when a selected
pin causes the defect. Scope includes third-party vulnerabilities when
configuration causes the defect. Scope includes third-party vulnerabilities
when permissions cause the defect. Scope includes third-party vulnerabilities
when an integration causes the defect. Route purely upstream defects to the
upstream maintainers.

Scope excludes social engineering. Scope excludes physical attacks.

This library parses packets from an untrusted local network. Decoding defects
in `status.go`, `remotedb.go`, and `structs.go` fall in scope. The library
requires a live PRO DJ LINK network for reproduction.

## Reporting a Vulnerability

Report vulnerabilities through GitHub private vulnerability reporting. Open
the Security tab. Select Report a vulnerability. Never open a public issue for
a vulnerability report.

Response timing: 7 days to acknowledgment.

Remediation target: 90 days from acknowledgment.

## Disclosure Policy

Coordinate disclosure with the maintainers. Keep the report private before a
disclosure event. A shipped fix defines one disclosure event. Passage of
90 days from acknowledgment defines another disclosure event. Use the earlier
event.
