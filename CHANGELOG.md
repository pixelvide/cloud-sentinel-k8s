# Changelog

## 1.0.0 (2026-01-16)


### Features

* add delete action for K8s resources with audit logging ([7cd1c27](https://github.com/pixelvide/cloud-sentinel-k8s/commit/7cd1c27a3e94b6decac8b0ae661043d4ab893e7f))
* add helm releases management ([f5892ea](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f5892ea68704ce97a66ab1d17555b786d37e7dba))
* add Kubernetes Config resources (ConfigMaps, Secrets, Quotas, Limits, HPA) and reorganize sidebar navigation ([e3abd1e](https://github.com/pixelvide/cloud-sentinel-k8s/commit/e3abd1e3f35830e2e68644f2a2a18b5c4a31b59f))
* add multi-pod logging support with pod selector UI ([a0489f0](https://github.com/pixelvide/cloud-sentinel-k8s/commit/a0489f0eb46f8782e44530f57d66ed842516eb12))
* add project documentation, roadmap and unified Dockerfile ([74027bf](https://github.com/pixelvide/cloud-sentinel-k8s/commit/74027bf02e817c8861173556bb1caec6b61c8bd9))
* add related pods table and upgrade k8s deps ([e89a697](https://github.com/pixelvide/cloud-sentinel-k8s/commit/e89a69721b05a326d40cc9c7089fc13e7f96a058))
* add ReplicaSet support to workload menu and dashboard ([05a2e92](https://github.com/pixelvide/cloud-sentinel-k8s/commit/05a2e92b710cd9a84a178a8cc14194a4d39412da))
* **api:** update backend routes to hyphenated paths and group in main.go ([8550765](https://github.com/pixelvide/cloud-sentinel-k8s/commit/8550765fd2f460b638a5111481b39583793ef187))
* auto-derive OIDC redirect, add docs, and refine config ([5dde4c3](https://github.com/pixelvide/cloud-sentinel-k8s/commit/5dde4c3c16c7fc4be3a42237981363dcb25d8c95))
* CRD support, dashboard updates, and UI fixes ([08c70e2](https://github.com/pixelvide/cloud-sentinel-k8s/commit/08c70e227585a66da78465e1ae93dd8256f45e28))
* enhance kube resource properties with taints, conditions, and scheduling details ([100c55b](https://github.com/pixelvide/cloud-sentinel-k8s/commit/100c55bd07030f2e1d17906d2b6ac65d4817cbfd))
* enhance kubernetes resource details with properties section and light theme sidebar ([ed4f617](https://github.com/pixelvide/cloud-sentinel-k8s/commit/ed4f617ef70986d1a3e9e1237be4164143bd2b9d))
* enhance log viewer with init containers, unified streaming, and stability fixes ([5ca9141](https://github.com/pixelvide/cloud-sentinel-k8s/commit/5ca9141bd876a04a4f4ad7f0bae3ad36f3269ce1))
* Enhance MultiSelect and Implement All Namespaces/Containers Support ([c9aa6be](https://github.com/pixelvide/cloud-sentinel-k8s/commit/c9aa6be34fb0577827c0686094fc6e0ddd941362))
* expand k8s resources (storage, access control) and refactor navigation ([f07d1a9](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f07d1a94dd8780c8a7fca35b5e3dcfe59ddde2e8))
* expand kube resources and centralize navigation config ([e421624](https://github.com/pixelvide/cloud-sentinel-k8s/commit/e4216246eb059502eec86a542735eb18868429d9))
* **frontend:** centralize log viewing and refine side panel theming ([53e927e](https://github.com/pixelvide/cloud-sentinel-k8s/commit/53e927e6d813dd346afa2ab09c70d9fae33b13e9))
* implement comprehensive audit logging and multi-auth support ([ecd39eb](https://github.com/pixelvide/cloud-sentinel-k8s/commit/ecd39eb4df08039838581ce061571b11549274f7))
* implement CronJob suspend/resume and fix cluster-scoped resource details ([1bc0620](https://github.com/pixelvide/cloud-sentinel-k8s/commit/1bc0620c5862172c058b5a260617ab3b0b19e711))
* implement mobile keyboard, exec header fix, and PWA support ([32266da](https://github.com/pixelvide/cloud-sentinel-k8s/commit/32266da7eddfb78cc10a8dd3d8e74c8c68a57039))
* implement node cordon/drain with reusable confirm dialog and audit logging ([b2fb135](https://github.com/pixelvide/cloud-sentinel-k8s/commit/b2fb1350a334f11f0fef83702b7c05e1149bdc0d))
* implement resource edit with dynamic client refactor ([528da6b](https://github.com/pixelvide/cloud-sentinel-k8s/commit/528da6b553c0bc66b94ea3f6a9001aa7e48d3efc))
* **kube:** implement conditional namespace selector and improve resource details ([9323468](https://github.com/pixelvide/cloud-sentinel-k8s/commit/93234688faf742e323533982c82ec512256bcfd0))
* **logs:** implement multi-pod selection and container awareness ([f9bbbb1](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f9bbbb10ab5c17be5082ee2e3cf118f7611dbb14))
* make frontend port configurable via FRONTEND_PORT ([08215da](https://github.com/pixelvide/cloud-sentinel-k8s/commit/08215da6939a6273772db1619d4a72ec43720762))
* Pod Log Streaming, Resource Details Slider & UI Improvements ([97bc463](https://github.com/pixelvide/cloud-sentinel-k8s/commit/97bc46355a655d0969301e9c6396d4ed3ce79843))
* refactor kubernetes routes and dynamic sidebar ([d8878d5](https://github.com/pixelvide/cloud-sentinel-k8s/commit/d8878d545fcebbd7906fad4249fc3d7aaf570632))
* refine CRD UI and fix namespaced resource fetching ([2c865dd](https://github.com/pixelvide/cloud-sentinel-k8s/commit/2c865dd94054dbfd9f399b87de731d23b3492dc8))
* **ui:** improve checkbox and multi-select components, refine log viewer UI ([393579a](https://github.com/pixelvide/cloud-sentinel-k8s/commit/393579a7e444de421fbc92b7b9ec949513e1de21))
* **ui:** refine checkbox styling and multi-select icons ([f83e51c](https://github.com/pixelvide/cloud-sentinel-k8s/commit/f83e51cb4e221c36ee79558f1e24f474f53681a9))


### Bug Fixes

* standardise timestamps and hide mobile nav on exec ([032c5ae](https://github.com/pixelvide/cloud-sentinel-k8s/commit/032c5ae004c0185027b350e1684dd6db265455d9))
