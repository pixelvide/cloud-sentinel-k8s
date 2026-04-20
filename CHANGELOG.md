# Changelog

## [1.2.1](https://github.com/pixelvide/kube-sentinel/compare/v1.2.0...v1.2.1) (2026-04-20)


### Bug Fixes

* Image cleanup issue ([#132](https://github.com/pixelvide/kube-sentinel/issues/132)) ([e2c208c](https://github.com/pixelvide/kube-sentinel/commit/e2c208c95b675407bcc9d89029f727a31b43d6f2))

## [1.2.0](https://github.com/pixelvide/kube-sentinel/compare/v1.1.2...v1.2.0) (2026-03-31)


### Features

* helm list all releases ([#111](https://github.com/pixelvide/kube-sentinel/issues/111)) ([baa6009](https://github.com/pixelvide/kube-sentinel/commit/baa60096ec5586908f35529958ab56d41cf3c0ef))


### Bug Fixes

* image pull related issues ([#126](https://github.com/pixelvide/kube-sentinel/issues/126)) ([64f4e92](https://github.com/pixelvide/kube-sentinel/commit/64f4e9240b3757931fe3381a1661a3185b908cc7))

## [1.1.2](https://github.com/pixelvide/kube-sentinel/compare/v1.1.1...v1.1.2) (2026-02-16)


### Bug Fixes

* **helm:** show failed and pending releases in list view ([#108](https://github.com/pixelvide/kube-sentinel/issues/108)) ([edc7274](https://github.com/pixelvide/kube-sentinel/commit/edc727456c870210e4bd649b2ca7f3c5a8d09519))

## [1.1.1](https://github.com/pixelvide/kube-sentinel/compare/v1.1.0...v1.1.1) (2026-02-13)


### Bug Fixes

* Encrypt AI API keys in database ([#104](https://github.com/pixelvide/kube-sentinel/issues/104)) ([ae436b3](https://github.com/pixelvide/kube-sentinel/commit/ae436b39285383e475ba2e89556e27d3402e64d4))

## [1.1.0](https://github.com/pixelvide/kube-sentinel/compare/v1.0.1...v1.1.0) (2026-02-13)


### Features

* encrypt GitLab tokens and AI API keys ([#103](https://github.com/pixelvide/kube-sentinel/issues/103)) ([00f2210](https://github.com/pixelvide/kube-sentinel/commit/00f221013fe6cbc5062ee232931a8f3ef36c6151))


### Bug Fixes

* **auth:** correct handling of empty OIDC groups during login ([#99](https://github.com/pixelvide/kube-sentinel/issues/99)) ([4a136a2](https://github.com/pixelvide/kube-sentinel/commit/4a136a2a0c496925b695d72246f0ed25fb2b112e))
* RBAC role creation defaults and input handling ([#101](https://github.com/pixelvide/kube-sentinel/issues/101)) ([3a86494](https://github.com/pixelvide/kube-sentinel/commit/3a86494538d10096143af2bfcccc9e180f2b5cd6))
* RBAC table column overflow issues ([#102](https://github.com/pixelvide/kube-sentinel/issues/102)) ([82e1af7](https://github.com/pixelvide/kube-sentinel/commit/82e1af7fe0476e9fce6678895acd8616d1a2258f))

## [1.0.1](https://github.com/pixelvide/kube-sentinel/compare/v1.0.0...v1.0.1) (2026-02-13)


### Bug Fixes

* **auth:** persist OIDCGroups in JWT for RBAC ([#97](https://github.com/pixelvide/kube-sentinel/issues/97)) ([9eb80e6](https://github.com/pixelvide/kube-sentinel/commit/9eb80e68e85c599cb6d37c121439e010d763fbf4))
* **security:** [HIGH] Fix path traversal in proxy handler ([#95](https://github.com/pixelvide/kube-sentinel/issues/95)) ([d8ab2d0](https://github.com/pixelvide/kube-sentinel/commit/d8ab2d0e7a715a8d3fa3c6b703ae7ff26cfc3598))

## [1.0.0](https://github.com/pixelvide/kube-sentinel/compare/v0.13.1...v1.0.0) (2026-02-13)


### ⚠ BREAKING CHANGES

* Renamed all CLOUD_SENTINEL_K8S_* environment variables to KUBE_SENTINEL_*. This includes ENCRYPT_KEY, BASE, USERNAME, PASSWORD, and CI/CD variables.

### Features

* add app_id to oauth_providers and move to core schema ([#86](https://github.com/pixelvide/kube-sentinel/issues/86)) ([9ab5b61](https://github.com/pixelvide/kube-sentinel/commit/9ab5b6119d16475cb10c132ee7bc679b8f7f9f75))
* Add Email field to User model and capture from OIDC ([#92](https://github.com/pixelvide/kube-sentinel/issues/92)) ([9129957](https://github.com/pixelvide/kube-sentinel/commit/912995788097c4414d87482c8b62ca7bf6d98bef))
* cleanup untagged image ([#72](https://github.com/pixelvide/kube-sentinel/issues/72)) ([a8c972d](https://github.com/pixelvide/kube-sentinel/commit/a8c972dd2368bee8666baf6105ab009c9dc46f6e))
* support custom postgres schema configuration ([#83](https://github.com/pixelvide/kube-sentinel/issues/83)) ([0b5f26f](https://github.com/pixelvide/kube-sentinel/commit/0b5f26fc155ca1fe6b5913188c78a84701e34423))
* **ui:** improve resource table accessibility ([#93](https://github.com/pixelvide/kube-sentinel/issues/93)) ([0ff4787](https://github.com/pixelvide/kube-sentinel/commit/0ff4787799f8cca69d9e7acb259a13d84ede1e09))


### Bug Fixes

* [CRITICAL] Fix hardcoded JWT secret ([#88](https://github.com/pixelvide/kube-sentinel/issues/88)) ([492102f](https://github.com/pixelvide/kube-sentinel/commit/492102f2177541c640a9aecc9c9526e135b73335))
* **security:** [HIGH] Fix CORS vulnerability by restricting allowed origins ([#90](https://github.com/pixelvide/kube-sentinel/issues/90)) ([59f1875](https://github.com/pixelvide/kube-sentinel/commit/59f1875826d026f89261ba8400a19ab47764fce4))
* **security:** prevent SSRF in image tags handler ([#89](https://github.com/pixelvide/kube-sentinel/issues/89)) ([b54230c](https://github.com/pixelvide/kube-sentinel/commit/b54230c9efbf0881c4c9cf7a9948eb01982489e6))


### Code Refactoring

* Rename CLOUD_SENTINEL_K8S_* environment variables to KUBE_SENTINEL_* ([#82](https://github.com/pixelvide/kube-sentinel/issues/82)) ([580bca1](https://github.com/pixelvide/kube-sentinel/commit/580bca1c7b37f04435de83ac119e2db1fd9e08d3))

## [0.13.1](https://github.com/pixelvide/kube-sentinel/compare/v0.13.0...v0.13.1) (2026-01-28)


### Bug Fixes

* ensure namespace refresh on cluster switch ([43c0d52](https://github.com/pixelvide/kube-sentinel/commit/43c0d5215c62d5f3f40e306a9cc93639a4fb16c7))

## [0.13.0](https://github.com/pixelvide/kube-sentinel/compare/v0.12.0...v0.13.0) (2026-01-27)


### Features

* add AI knowledge base and debug tool ([ef34310](https://github.com/pixelvide/kube-sentinel/commit/ef34310b398dec8bb869183c1c0dcd620cf48413))
* add debug_app_connection tool and improve AI chat handling ([d5b6970](https://github.com/pixelvide/kube-sentinel/commit/d5b6970ea5204dbba1cb250e75d5496599499d4c))

## [0.12.0](https://github.com/pixelvide/kube-sentinel/compare/v0.11.0...v0.12.0) (2026-01-27)


### Features

* enhance security dashboard with comprehensive report  ([#64](https://github.com/pixelvide/kube-sentinel/issues/64)) ([0091eea](https://github.com/pixelvide/kube-sentinel/commit/0091eea588f39c1b2b131add2d258ad6a3e51e71))

## [0.11.0](https://github.com/pixelvide/kube-sentinel/compare/v0.10.0...v0.11.0) (2026-01-26)


### Features

* Expand DescribeResourceTool support ([#62](https://github.com/pixelvide/kube-sentinel/issues/62)) ([8d721c3](https://github.com/pixelvide/kube-sentinel/commit/8d721c39d969e5a00e4c449338d070eb23579d4f))


### Bug Fixes

* **cr_handler:** support multi-namespace listing for CRDs ([0146168](https://github.com/pixelvide/kube-sentinel/commit/01461682859952bd66c2f202c8f572a9edb8f188))

## [0.10.0](https://github.com/pixelvide/kube-sentinel/compare/v0.9.0...v0.10.0) (2026-01-23)


### Features

* ⚡ Optimize AWS config restoration performance ([#58](https://github.com/pixelvide/kube-sentinel/issues/58)) ([5b494e3](https://github.com/pixelvide/kube-sentinel/commit/5b494e36a294b1b9977f9553fa163cb6682b7ee9))
* **ai:** integrate AI assistant with chat history and k8s tools ([#59](https://github.com/pixelvide/kube-sentinel/issues/59)) ([5a8f4de](https://github.com/pixelvide/kube-sentinel/commit/5a8f4decffee77067010a4cffc621c441d137336))
* Replace Bubble Sort with efficient sort.Slice ([#55](https://github.com/pixelvide/kube-sentinel/issues/55)) ([aea4d4f](https://github.com/pixelvide/kube-sentinel/commit/aea4d4f41fca9ed5d063d41a514027aff1b7d828))

## [0.9.0](https://github.com/pixelvide/kube-sentinel/compare/v0.8.0...v0.9.0) (2026-01-22)


### Features

* upgrade helm release UI and backend ([#54](https://github.com/pixelvide/kube-sentinel/issues/54)) ([f1f9402](https://github.com/pixelvide/kube-sentinel/commit/f1f9402cd005bbc01b5c0a519980b2b264677086))


### Bug Fixes

* **ui:** correct pluralization for k8s resources using pluralize library ([#50](https://github.com/pixelvide/kube-sentinel/issues/50)) ([f8dcc1a](https://github.com/pixelvide/kube-sentinel/commit/f8dcc1af1849fcd3a4ad51d294e9be731d90147b))

## [0.8.0](https://github.com/pixelvide/kube-sentinel/compare/v0.7.0...v0.8.0) (2026-01-21)


### Features

* add Endpoints, IngressClasses, and NetworkPolicies to Traffic menu ([#46](https://github.com/pixelvide/kube-sentinel/issues/46)) ([528ddc4](https://github.com/pixelvide/kube-sentinel/commit/528ddc42bb5cf8a18e88b6f5fe8b8513dd549cff))
* add ReplicaSets and ReplicationControllers to Workloads menu ([#48](https://github.com/pixelvide/kube-sentinel/issues/48)) ([d182af4](https://github.com/pixelvide/kube-sentinel/commit/d182af40508acba01d942d7b3e0826ffafd9d700))
* add Resource Quotas and Limit Ranges menus ([#44](https://github.com/pixelvide/kube-sentinel/issues/44)) ([58a6f10](https://github.com/pixelvide/kube-sentinel/commit/58a6f10b1cdbfdd09042a78c5f83cc82d8d37750))


### Bug Fixes

* **ui:** correct empty state message for all namespaces ([ac9e15d](https://github.com/pixelvide/kube-sentinel/commit/ac9e15dc03d6c0427f586eaa137b7c15cf6d44f9))

## [0.7.0](https://github.com/pixelvide/kube-sentinel/compare/v0.6.2...v0.7.0) (2026-01-21)


### Features

* Add Mutating and Validating Webhook menus ([#42](https://github.com/pixelvide/kube-sentinel/issues/42)) ([c0439b9](https://github.com/pixelvide/kube-sentinel/commit/c0439b98207e7a766a812dc4bdba428774081e84))
* Add Pod Disruption Budgets menu item ([#39](https://github.com/pixelvide/kube-sentinel/issues/39)) ([196f520](https://github.com/pixelvide/kube-sentinel/commit/196f520fdf7f0a5d18673199f00829f9769d9f35))
* Add PriorityClasses, RuntimeClasses, and Leases to sidebar ([#41](https://github.com/pixelvide/kube-sentinel/issues/41)) ([a2e0331](https://github.com/pixelvide/kube-sentinel/commit/a2e0331092ad43f510c13e807cde3908afe167bd))


### Performance Improvements

* Debounce user activity updates to prevent writes more frequently than every 5 minutes. ([d21afdf](https://github.com/pixelvide/kube-sentinel/commit/d21afdfd4c43e31680f3ffd2dbf011a2a0fa4e6c))

## [0.6.2](https://github.com/pixelvide/kube-sentinel/compare/v0.6.1...v0.6.2) (2026-01-21)


### Bug Fixes

* **ui:** enable password reset for admin users ([#34](https://github.com/pixelvide/kube-sentinel/issues/34)) ([#37](https://github.com/pixelvide/kube-sentinel/issues/37)) ([376052b](https://github.com/pixelvide/kube-sentinel/commit/376052bdcb12485544a5ab7f35248085bfa870e9))

## [0.6.1](https://github.com/pixelvide/kube-sentinel/compare/v0.6.0...v0.6.1) (2026-01-21)


### Bug Fixes

* update SkipSystemSync via API and UI (Fixes [#33](https://github.com/pixelvide/kube-sentinel/issues/33)) ([#35](https://github.com/pixelvide/kube-sentinel/issues/35)) ([0cfd7dc](https://github.com/pixelvide/kube-sentinel/commit/0cfd7dc732277912a5bd4daa76336db8dd32efb5))

## [0.6.0](https://github.com/pixelvide/kube-sentinel/compare/v0.5.0...v0.6.0) (2026-01-21)


### Features

* Add INSECURE_SKIP_VERIFY option for HTTP clients and refine user and identity management. ([bd5213d](https://github.com/pixelvide/kube-sentinel/commit/bd5213db4e8bd55b27c4028e0698f768c88118c2))
* Introduce `.env.example` for Docker Compose environment variables and update `.gitignore`. ([3243fdd](https://github.com/pixelvide/kube-sentinel/commit/3243fdd099d9bb80f17ed356f9c1ad4a784f19ca))

## [0.5.0](https://github.com/pixelvide/kube-sentinel/compare/v0.4.0...v0.5.0) (2026-01-21)


### Features

* Add Helm section ([#29](https://github.com/pixelvide/kube-sentinel/issues/29)) ([9e6d9e9](https://github.com/pixelvide/kube-sentinel/commit/9e6d9e9a28f885a1272b9a18c133e59b65235ce2))
* Implement Helm Release Delete API ([#31](https://github.com/pixelvide/kube-sentinel/issues/31)) ([a79e471](https://github.com/pixelvide/kube-sentinel/commit/a79e4718d44f1b8299c0efd2ade1fd982343a57f))

## [0.4.0](https://github.com/pixelvide/kube-sentinel/compare/v0.3.1...v0.4.0) (2026-01-21)


### Features

* Add manual trigger for the release workflow via `workflow_dispatch`. ([2f553e9](https://github.com/pixelvide/kube-sentinel/commit/2f553e91b0d45e521eab954aabf6e4d639478b38))

## [0.3.1](https://github.com/pixelvide/kube-sentinel/compare/v0.3.0...v0.3.1) (2026-01-21)


### Bug Fixes

* Exclude "v0.0.0" tag from the list of versions used to build documentation. ([2cce097](https://github.com/pixelvide/kube-sentinel/commit/2cce097e2898c3c8202d308a20eaef3b4e74d3e4))

## [0.3.0](https://github.com/pixelvide/kube-sentinel/compare/v0.2.0...v0.3.0) (2026-01-21)


### Features

* Bump kube-sentinel chart and app versions to 0.2.0, update documentation, and enhance the release script with improved `sed` compatibility and workflow ([460ad2d](https://github.com/pixelvide/kube-sentinel/commit/460ad2d5319b6ba5dacc9a34ddd00bc86afc1bff))

## [0.2.0](https://github.com/pixelvide/kube-sentinel/compare/v0.1.0...v0.2.0) (2026-01-21)


### Features

* add Kubernetes icon SVG asset. ([01acd3f](https://github.com/pixelvide/kube-sentinel/commit/01acd3fbaf3c3859cf9cdeeebb02f025fb7315c6))
* Bump kube-sentinel chart and app version to 0.1.0 and refine the release script's version syncing logic. ([b828948](https://github.com/pixelvide/kube-sentinel/commit/b8289480d77a9d43cbd2592ec49abde8564918da))

## [0.1.0](https://github.com/pixelvide/kube-sentinel/compare/v0.0.0...v0.1.0) (2026-01-21)


### Features

* Add changelog-path to kube-sentinel release configuration. ([e5cc60b](https://github.com/pixelvide/kube-sentinel/commit/e5cc60b7c08628efdb796369f9eb9f94c778261c))
* configure release-please to update Helm chart versions and image tags in values.yaml and Chart.yaml. ([0f6385a](https://github.com/pixelvide/kube-sentinel/commit/0f6385a27291f54aad9ac35f034389cbbee04dae))
* Implement multi-version documentation deployment to GitHub Pages and update Docker image registry from `zzde` to `pixelvide`. ([107a881](https://github.com/pixelvide/kube-sentinel/commit/107a881ad9a63108777c522642f33e0f029ef706))
* Implement Release Please for automated versioning and releases, and add a semantic pull request validation workflow. ([a78a8a8](https://github.com/pixelvide/kube-sentinel/commit/a78a8a8bd8565d935aba2aea715f061d26cc74ab))
* Include `glab` and `aws-iam-authenticator` in the Docker image, refine Go module dependencies, and enhance test environment deployment with secret validation. ([8e5cf62](https://github.com/pixelvide/kube-sentinel/commit/8e5cf627f1f49e98a5b282b62a70fe93b2715a9f))
* Integrate kube-sentinel Helm chart into release-please and initialize versions to 0.0.0. ([3643f6e](https://github.com/pixelvide/kube-sentinel/commit/3643f6ea0ec25acfd42f5019b570ba8f93811bad))
* Move 'v' prefix from chart `version` to `appVersion` in Chart.yaml. ([48715c6](https://github.com/pixelvide/kube-sentinel/commit/48715c6c11e6d460cbf3f43d449f9c9267a97093))
* Update release-please-action to `googleapis` and rename the workflow to `Release Please`. ([f4d91e1](https://github.com/pixelvide/kube-sentinel/commit/f4d91e1eb6566cc4dde607b84d83272f11953ad8))
