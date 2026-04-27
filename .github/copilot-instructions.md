# Copilot Instructions for serverless-proxy

<!-- security-checklist-managed -->

## Security Checklist

Apply this checklist to every code change. If a control is not applicable, briefly say why in the PR description.

### Authentication
- Never trust client-supplied identity (user IDs, org IDs, tenant IDs) from request bodies, query strings, or headers other than verified auth tokens.
- All authenticated endpoints must use the shared auth middleware from `commons` (JWT/OAuth verification, signature + expiry + audience checks). Do not re-implement token parsing.
- Service-to-service calls must use mTLS or signed service tokens — no anonymous internal endpoints exposed beyond `localhost`/in-cluster.
- API keys, bootstrap tokens, and refresh tokens must be hashed at rest (bcrypt/argon2 or HMAC with a server-side pepper) — never stored in plaintext.

### Authorization
- Enforce RBAC at the handler boundary, not only in the UI or upstream caller. Every mutating handler must call the authorization helper before acting.
- Check resource ownership (tenant/org/service-provider) on every read and write — `GET /resource/{id}` must verify the caller owns `{id}`.
- Default-deny: unknown roles, missing scopes, or missing tenant context must reject with 403, not fall through.
- Admin/internal endpoints must be gated by an explicit "internal" role and, where possible, network-restricted.

### Tenant Isolation
- Every database query, cache lookup, queue operation, and S3/object-store path must be scoped by `tenantID` / `orgID` / `serviceProviderID`. Treat a missing tenant scope as a bug.
- Never look up a resource by a primary key alone if that key is user-supplied — always combine with the caller's tenant scope in the WHERE clause.
- Cross-tenant operations (admin tooling, migration jobs) must be explicitly marked, audited, and require an internal role.
- Workflow inputs (Temporal activities) must include and re-validate tenant context — do not rely on the caller having validated it.

### Input Validation
- Validate all external input at the handler boundary: type, length, charset, range, and allowed-value set.
- Use parameterized queries / prepared statements for SQL. Never concatenate or `fmt.Sprintf` user input into SQL, shell, regex, template, or URL strings.
- Reject unknown JSON fields where strict shape matters; cap request body sizes and array lengths.
- Validate file paths against traversal (`..`, absolute paths, symlinks) before any filesystem access.
- Treat YAML and Helm value inputs as untrusted; use safe loaders and forbid arbitrary Go template execution on user data.

### Secrets Handling
- No secrets in source, tests, fixtures, sample configs, or commit history. Run `gitleaks` / `git-secrets` mentally before committing.
- Read secrets from environment variables, AWS Secrets Manager, or the platform secret store — never from disk paths a user can influence.
- Do not log secret values, tokens, signed URLs, connection strings, or full `Authorization` headers — even at debug level.
- Redact secrets from error messages before returning them to the caller; wrap upstream errors instead of bubbling raw output.
- Rotate any secret that touches a logged code path; treat accidental log exposure as a security incident.

### Logging Hygiene
- Use the structured logger from `commons`. No `fmt.Println` / `log.Printf` in production code paths.
- Never log: passwords, tokens, API keys, license keys, full request/response bodies, customer PII, billing data, or full SQL with values.
- Log identifiers (tenantID, requestID, resourceID), not contents. Use stable correlation IDs across services.
- Set log levels deliberately: `error` for actionable failures, `warn` for recoverable anomalies, `info` for lifecycle, `debug` only for development.
- Ensure new log lines do not break existing redaction filters or SIEM parsers.

### Dependency & Supply Chain
- Run `govulncheck ./...` for any change that touches `go.mod` or adds imports. Address any new findings or document the exception.
- Pin module versions; avoid `replace` directives pointing at forks unless reviewed.
- Prefer stdlib and existing internal packages over new third-party deps. Justify every new dependency in the PR description.
- Keep Dockerfiles on minimal base images (`distroless`, `alpine` with explicit pinned digest); never pull `:latest`.
- For new GitHub Actions, pin to a commit SHA, not a tag.

### Crypto & Transport
- Use stdlib `crypto/*` and well-vetted libraries; do not implement custom crypto, custom JWT parsing, or custom session schemes.
- Enforce TLS 1.2+ on all outbound clients; never set `InsecureSkipVerify: true` outside of explicitly-flagged test code.
- Use `crypto/rand` for any security-relevant randomness (tokens, IDs, nonces). `math/rand` is for non-security use only.

### SSRF & Outbound Egress
- For any HTTP client whose target host is influenced by user input (webhooks, callback URLs, proxy targets, image fetchers, OIDC discovery URLs), block requests to private (`10/8`, `172.16/12`, `192.168/16`), loopback, link-local (`169.254/16`, including the cloud metadata service), and `0.0.0.0` ranges, plus internal control-plane hostnames. Resolve the host and re-check the resolved IP, not just the input string.
- Prefer an explicit allowlist of destination hosts where the use case permits it.
- Do not follow redirects to addresses that fail the above check.

### Cross-Boundary Credentials
- Calls into customer data planes / customer cloud accounts must use per-tenant scoped credentials (assumed roles, workload identity), never a shared high-privilege control-plane credential. Audit-log every such call with the caller identity, target tenant, and action.

### What to do when unsure
- If a change touches authn, authz, tenant scoping, crypto, or secret handling and the right approach is not obvious, stop and ask in the PR description rather than guessing.
- Prefer adding a regression test that proves the security property (e.g., "tenant A cannot read tenant B's resource") over a comment claiming it.
