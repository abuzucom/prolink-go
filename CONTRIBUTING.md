# Contributing

## Development Setup

- Install the Go toolchain at version 1.13 or newer. Dependencies resolve
  through `go.mod` with `go mod download`.
- Install the policy checker dependency with
  `python -m pip install --requirement requirements-checkers.txt`.
- Build with `go build ./...`.
- Run the Go test suite with `go test ./...`. The repository carries no Go
  tests yet. The command reports `no test files`.
- Run Go lint checks with `golangci-lint run` and `go vet ./...`.
- Run policy checks with `make check`, `make lint`, and `make test`.

The library needs a live PRO DJ LINK network with real CDJs for end-to-end
verification. CI cannot supply that hardware.

## Conventions

- Name branches `<type>/<kebab-description>`. Use feat, fix, chore, docs, or
  test as the type. Never use the `claude/` prefix.
- Write commit subjects in `type: description` form. Use imperative mood.
  Limit subjects to 50 characters. Omit the trailing period.
- Read `AGENTS.md` before any change. The file states the full policy for
  human and agent contributors alike.

## Safety

- Use parameterized queries for untrusted input.
- Run subprocesses with argument arrays. Disable shell interpretation.
- Obtain explicit approval before destructive changes.
- Never weaken a test to force a pass. Never skip a test to force a pass.
  Never delete a test to force a pass. Stop after suspecting a test defect. Ask
  a maintainer before changing the test.
- Keep public APIs backward compatible. Every exported identifier in the root
  package, `bpm`, and `mixstatus` forms the public surface.
- Never commit secrets, credentials, or private user data.
- Use SHA-256 or SHA-3 for general hashing.
- Use bcrypt, scrypt, or Argon2 with a salt and work factor for passwords.
- Run containers under a non-root runtime account.
- Set `actions/checkout` `persist-credentials: false` unless the job needs the
  checked-out credential afterward. Valid needs include pushes. Valid needs
  include Git credential helper access. Valid needs include private submodule
  fetches. Valid needs include private LFS object fetches. Document each
  exception with
  `# persist-credentials: true: this job <reason> (Rule 11 exception).`
- Add commits instead of rewriting shared branch history.

## Dependencies

- Before changing a dependency, propose the name and pinned version. State the
  purpose. List alternatives. Obtain maintainer approval.
- Pin approved dependencies to exact versions.
- A reusable GitHub Actions workflow counts as a dependency. Pin every action
  and workflow to a full commit SHA.

## Licensing

The repository carries a split license. Files inherited from the initial fork
stay MIT under `LICENSE`. Every new file carries BSD-3-Clause under
`LICENSE.BSD-3-Clause`. A contribution to a new file falls under
BSD-3-Clause. A contribution to an inherited file falls under MIT.
`docs/template-drift.md` records the file boundary.

## Change Workflow

1. Create a branch following the naming convention above.
2. Add a test that exercises the real behavior. Run the test. Confirm the
   expected failure.
3. Implement the smallest change that makes the new test pass.
4. Run `go test ./...`, `golangci-lint run`, `make check`, `make lint`, and
   `make test`. Fix failures without suppressing checks.
5. Update README for substantial changes. Update CHANGELOG for all changes.
   Follow Semantic Versioning.
6. Commit using the subject convention above.
7. Open every pull request as a draft. Describe the change. Describe the risks.
   Include verification results.
