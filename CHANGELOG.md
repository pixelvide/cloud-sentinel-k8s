# Changelog

## [0.13.1](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.13.0...v0.13.1) (2026-01-28)


### Bug Fixes

* ensure namespace refresh on cluster switch ([43c0d52](https://github.com/pixelvide/cloud-sentinel-k8s/commit/43c0d5215c62d5f3f40e306a9cc93639a4fb16c7))

## [0.13.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.12.0...v0.13.0) (2026-01-27)


### Features

* add AI knowledge base and debug tool ([ef34310](https://github.com/pixelvide/cloud-sentinel-k8s/commit/ef34310b398dec8bb869183c1c0dcd620cf48413))
* add debug_app_connection tool and improve AI chat handling ([d5b6970](https://github.com/pixelvide/cloud-sentinel-k8s/commit/d5b6970ea5204dbba1cb250e75d5496599499d4c))

## [0.12.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.11.0...v0.12.0) (2026-01-27)


### Features

* enhance security dashboard with comprehensive report  ([#64](https://github.com/pixelvide/cloud-sentinel-k8s/issues/64)) ([0091eea](https://github.com/pixelvide/cloud-sentinel-k8s/commit/0091eea588f39c1b2b131add2d258ad6a3e51e71))

## [0.11.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.10.0...v0.11.0) (2026-01-26)


### Features

* Expand DescribeResourceTool support ([#62](https://github.com/pixelvide/cloud-sentinel-k8s/issues/62)) ([8d721c3](https://github.com/pixelvide/cloud-sentinel-k8s/commit/8d721c39d969e5a00e4c449338d070eb23579d4f))


### Bug Fixes

* **cr_handler:** support multi-namespace listing for CRDs ([0146168](https://github.com/pixelvide/cloud-sentinel-k8s/commit/01461682859952bd66c2f202c8f572a9edb8f188))

## [0.10.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.9.0...v0.10.0) (2026-01-23)


### Features

* ⚡ Optimize AWS config restoration performance ([#58](https://github.com/pixelvide/cloud-sentinel-k8s/issues/58)) ([5b494e3](https://github.com/pixelvide/cloud-sentinel-k8s/commit/5b494e36a294b1b9977f9553fa163cb6682b7ee9))
* **ai:** integrate AI assistant with chat history and k8s tools ([#59](https://github.com/pixelvide/cloud-sentinel-k8s/issues/59)) ([5a8f4de](https://github.com/pixelvide/cloud-sentinel-k8s/commit/5a8f4decffee77067010a4cffc621c441d137336))
* Replace Bubble Sort with efficient sort.Slice ([#55](https://github.com/pixelvide/cloud-sentinel-k8s/issues/55)) ([aea4d4f](https://github.com/pixelvide/cloud-sentinel-k8s/commit/aea4d4f41fca9ed5d063d41a514027aff1b7d828))

## [0.9.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.8.0...v0.9.0) (2026-01-22)


### Features

* upgrade helm release UI and backend ([#54](https://github.com/pixelvide/cloud-sentinel-k8s/issues/54)) ([f1f9402](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f1f9402cd005bbc01b5c0a519980b2b264677086))


### Bug Fixes

* **ui:** correct pluralization for k8s resources using pluralize library ([#50](https://github.com/pixelvide/cloud-sentinel-k8s/issues/50)) ([f8dcc1a](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f8dcc1af1849fcd3a4ad51d294e9be731d90147b))

## [0.8.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.7.0...v0.8.0) (2026-01-21)


### Features

* add Endpoints, IngressClasses, and NetworkPolicies to Traffic menu ([#46](https://github.com/pixelvide/cloud-sentinel-k8s/issues/46)) ([528ddc4](https://github.com/pixelvide/cloud-sentinel-k8s/commit/528ddc42bb5cf8a18e88b6f5fe8b8513dd549cff))
* add ReplicaSets and ReplicationControllers to Workloads menu ([#48](https://github.com/pixelvide/cloud-sentinel-k8s/issues/48)) ([d182af4](https://github.com/pixelvide/cloud-sentinel-k8s/commit/d182af40508acba01d942d7b3e0826ffafd9d700))
* add Resource Quotas and Limit Ranges menus ([#44](https://github.com/pixelvide/cloud-sentinel-k8s/issues/44)) ([58a6f10](https://github.com/pixelvide/cloud-sentinel-k8s/commit/58a6f10b1cdbfdd09042a78c5f83cc82d8d37750))


### Bug Fixes

* **ui:** correct empty state message for all namespaces ([ac9e15d](https://github.com/pixelvide/cloud-sentinel-k8s/commit/ac9e15dc03d6c0427f586eaa137b7c15cf6d44f9))

## [0.7.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.6.2...v0.7.0) (2026-01-21)


### Features

* Add Mutating and Validating Webhook menus ([#42](https://github.com/pixelvide/cloud-sentinel-k8s/issues/42)) ([c0439b9](https://github.com/pixelvide/cloud-sentinel-k8s/commit/c0439b98207e7a766a812dc4bdba428774081e84))
* Add Pod Disruption Budgets menu item ([#39](https://github.com/pixelvide/cloud-sentinel-k8s/issues/39)) ([196f520](https://github.com/pixelvide/cloud-sentinel-k8s/commit/196f520fdf7f0a5d18673199f00829f9769d9f35))
* Add PriorityClasses, RuntimeClasses, and Leases to sidebar ([#41](https://github.com/pixelvide/cloud-sentinel-k8s/issues/41)) ([a2e0331](https://github.com/pixelvide/cloud-sentinel-k8s/commit/a2e0331092ad43f510c13e807cde3908afe167bd))


### Performance Improvements

* Debounce user activity updates to prevent writes more frequently than every 5 minutes. ([d21afdf](https://github.com/pixelvide/cloud-sentinel-k8s/commit/d21afdfd4c43e31680f3ffd2dbf011a2a0fa4e6c))

## [0.6.2](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.6.1...v0.6.2) (2026-01-21)


### Bug Fixes

* **ui:** enable password reset for admin users ([#34](https://github.com/pixelvide/cloud-sentinel-k8s/issues/34)) ([#37](https://github.com/pixelvide/cloud-sentinel-k8s/issues/37)) ([376052b](https://github.com/pixelvide/cloud-sentinel-k8s/commit/376052bdcb12485544a5ab7f35248085bfa870e9))

## [0.6.1](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.6.0...v0.6.1) (2026-01-21)


### Bug Fixes

* update SkipSystemSync via API and UI (Fixes [#33](https://github.com/pixelvide/cloud-sentinel-k8s/issues/33)) ([#35](https://github.com/pixelvide/cloud-sentinel-k8s/issues/35)) ([0cfd7dc](https://github.com/pixelvide/cloud-sentinel-k8s/commit/0cfd7dc732277912a5bd4daa76336db8dd32efb5))

## [0.6.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.5.0...v0.6.0) (2026-01-21)


### Features

* Add INSECURE_SKIP_VERIFY option for HTTP clients and refine user and identity management. ([bd5213d](https://github.com/pixelvide/cloud-sentinel-k8s/commit/bd5213db4e8bd55b27c4028e0698f768c88118c2))
* Introduce `.env.example` for Docker Compose environment variables and update `.gitignore`. ([3243fdd](https://github.com/pixelvide/cloud-sentinel-k8s/commit/3243fdd099d9bb80f17ed356f9c1ad4a784f19ca))

## [0.5.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.4.0...v0.5.0) (2026-01-21)


### Features

* Add Helm section ([#29](https://github.com/pixelvide/cloud-sentinel-k8s/issues/29)) ([9e6d9e9](https://github.com/pixelvide/cloud-sentinel-k8s/commit/9e6d9e9a28f885a1272b9a18c133e59b65235ce2))
* Implement Helm Release Delete API ([#31](https://github.com/pixelvide/cloud-sentinel-k8s/issues/31)) ([a79e471](https://github.com/pixelvide/cloud-sentinel-k8s/commit/a79e4718d44f1b8299c0efd2ade1fd982343a57f))

## [0.4.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.3.1...v0.4.0) (2026-01-21)


### Features

* Add manual trigger for the release workflow via `workflow_dispatch`. ([2f553e9](https://github.com/pixelvide/cloud-sentinel-k8s/commit/2f553e91b0d45e521eab954aabf6e4d639478b38))

## [0.3.1](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.3.0...v0.3.1) (2026-01-21)


### Bug Fixes

* Exclude "v0.0.0" tag from the list of versions used to build documentation. ([2cce097](https://github.com/pixelvide/cloud-sentinel-k8s/commit/2cce097e2898c3c8202d308a20eaef3b4e74d3e4))

## [0.3.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.2.0...v0.3.0) (2026-01-21)


### Features

* Bump cloud-sentinel-k8s chart and app versions to 0.2.0, update documentation, and enhance the release script with improved `sed` compatibility and workflow ([460ad2d](https://github.com/pixelvide/cloud-sentinel-k8s/commit/460ad2d5319b6ba5dacc9a34ddd00bc86afc1bff))

## [0.2.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.1.0...v0.2.0) (2026-01-21)


### Features

* add Kubernetes icon SVG asset. ([01acd3f](https://github.com/pixelvide/cloud-sentinel-k8s/commit/01acd3fbaf3c3859cf9cdeeebb02f025fb7315c6))
* Bump cloud-sentinel-k8s chart and app version to 0.1.0 and refine the release script's version syncing logic. ([b828948](https://github.com/pixelvide/cloud-sentinel-k8s/commit/b8289480d77a9d43cbd2592ec49abde8564918da))

## [0.1.0](https://github.com/pixelvide/cloud-sentinel-k8s/compare/v0.0.0...v0.1.0) (2026-01-21)


### Features

* Add changelog-path to cloud-sentinel-k8s release configuration. ([e5cc60b](https://github.com/pixelvide/cloud-sentinel-k8s/commit/e5cc60b7c08628efdb796369f9eb9f94c778261c))
* configure release-please to update Helm chart versions and image tags in values.yaml and Chart.yaml. ([0f6385a](https://github.com/pixelvide/cloud-sentinel-k8s/commit/0f6385a27291f54aad9ac35f034389cbbee04dae))
* Implement multi-version documentation deployment to GitHub Pages and update Docker image registry from `zzde` to `pixelvide`. ([107a881](https://github.com/pixelvide/cloud-sentinel-k8s/commit/107a881ad9a63108777c522642f33e0f029ef706))
* Implement Release Please for automated versioning and releases, and add a semantic pull request validation workflow. ([a78a8a8](https://github.com/pixelvide/cloud-sentinel-k8s/commit/a78a8a8bd8565d935aba2aea715f061d26cc74ab))
* Include `glab` and `aws-iam-authenticator` in the Docker image, refine Go module dependencies, and enhance test environment deployment with secret validation. ([8e5cf62](https://github.com/pixelvide/cloud-sentinel-k8s/commit/8e5cf627f1f49e98a5b282b62a70fe93b2715a9f))
* Integrate cloud-sentinel-k8s Helm chart into release-please and initialize versions to 0.0.0. ([3643f6e](https://github.com/pixelvide/cloud-sentinel-k8s/commit/3643f6ea0ec25acfd42f5019b570ba8f93811bad))
* Move 'v' prefix from chart `version` to `appVersion` in Chart.yaml. ([48715c6](https://github.com/pixelvide/cloud-sentinel-k8s/commit/48715c6c11e6d460cbf3f43d449f9c9267a97093))
* Update release-please-action to `googleapis` and rename the workflow to `Release Please`. ([f4d91e1](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f4d91e1eb6566cc4dde607b84d83272f11953ad8))
