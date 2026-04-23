## 2025-05-15 - [CRITICAL] Path Traversal in Proxy Handler Bypass RBAC
**Vulnerability:** The `HandleProxy` function used `url.JoinPath` with user-controlled `name` and `namespace` parameters without validation. This allowed attackers to traverse up the URL path (e.g., `../services/myservice`) and access resources they didn't have RBAC permissions for (e.g., checking permission for `pods` but accessing `services`).
**Learning:** `url.JoinPath` resolves `..` elements. When constructing URLs for backend requests based on user input, validating that the input doesn't contain path traversal sequences is critical, especially when RBAC checks are performed on the input parameters before URL construction.
**Prevention:** Strictly validate all URL parameters used in backend requests. Ensure `name` and `namespace` do not contain `/` or `..`. Reject `path` parameters that contain `..`.

## 2026-02-17 - [HIGH] Insecure Default File Permissions for Credentials
**Vulnerability:** The `GetUserGlabConfigDir` and `WriteUserAWSCredentials` helper functions created directories and files with `0777` (world-writable/readable) and `0666` (world-writable/readable) permissions, exposing sensitive credentials to other users on the system.
**Learning:** Copy-pasting code or using overly permissive defaults (like `0777`) for convenience is a dangerous anti-pattern. Even in containerized environments, "least privilege" should apply to file system permissions to prevent privilege escalation or data leakage if a volume is shared or the container is compromised.
**Prevention:** Always use `0700` (user-only) for private directories and `0600` (user-only) for private files. Enforce this in code reviews and static analysis.
