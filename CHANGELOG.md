# Changelog

## [1.1.0](https://github.com/AobaIwaki123/lumitree/compare/lumitree-v1.0.0...lumitree-v1.1.0) (2026-08-22)


### Features

* **api:** fix OpenAPI YAML syntax, setup oapi-codegen, and generate typed client ([bd3b669](https://github.com/AobaIwaki123/lumitree/commit/bd3b66909f6a2f63eb88d2e034548ec1909e4d89))
* **api:** OpenAPI スキーマの構文修正と oapi-codegen によるクライアント自動生成基盤の導入 ([6597e7b](https://github.com/AobaIwaki123/lumitree/commit/6597e7b1edbd66f8607eace50e16f01236d7806a))
* **ci:** add docker-publish workflow for automated GHCR push on release branch, main, and tag ([f8f50dc](https://github.com/AobaIwaki123/lumitree/commit/f8f50dca7765036cd5b1320fdb005f9f11d2aa05))
* **ci:** add release branch automation script and tag-on-merge workflow ([9182318](https://github.com/AobaIwaki123/lumitree/commit/9182318d5419dc46a726cac84d34313cde1050f7))
* **ci:** add release-please workflow for automated Release PR on main merge ([fd184e3](https://github.com/AobaIwaki123/lumitree/commit/fd184e3c5becea570f32b7c4d8857c4500a541b9))
* **ci:** Phase 1 - setup Go module, CI workflow, config and structured logger ([b1c401d](https://github.com/AobaIwaki123/lumitree/commit/b1c401d943af1792713b18783fd5d8d435da104f))
* **ci:** release branch / main / tag 連動の OCI コンテナイメージ自動 Push ワークフロー導入 ([30cd31c](https://github.com/AobaIwaki123/lumitree/commit/30cd31c728cffa509336a19b4fd0b970ec486582))
* **ci:** release branch & release PR 自動化ワークフローおよび作成スクリプトの整備 ([b102609](https://github.com/AobaIwaki123/lumitree/commit/b102609d4eafbba3ed953a9e5cd34abff2801bd2))
* **ci:** release-please による main マージ連動の Release PR 自動起票 & リリース自動化の導入 ([3d6ce17](https://github.com/AobaIwaki123/lumitree/commit/3d6ce17c53c121772768707577096d4da175bedb))
* **core:** Phase 1 - TimeTree client adapter, model normalizer, ical exporter and CLI runner ([5a95d57](https://github.com/AobaIwaki123/lumitree/commit/5a95d57da679079f067ff749c1ed42c2480f823e))
* **k8s:** align manifests and ingress with journee style (cloudflare-tunnel, argocd) and fix errcheck ([ca9827b](https://github.com/AobaIwaki123/lumitree/commit/ca9827b8d8e08483a8722534798f557356d4d7f6))
* **release:** Phase 3 - GoReleaser, release workflow, live monitoring cron and complete docs ([9525b0a](https://github.com/AobaIwaki123/lumitree/commit/9525b0a3195ff7469cbe769b99664886562daf93))
* **server:** Phase 2 - HTTP server proxy, memory TTL cache, Dockerfile and Kubernetes manifests ([9ede8b0](https://github.com/AobaIwaki123/lumitree/commit/9ede8b09ed13ebe5cb74e9480df5f07315e4b2a3))
* **spec:** add OpenAPI 3.0.3 specification and mapping documentation ([f92970c](https://github.com/AobaIwaki123/lumitree/commit/f92970c7e78b253e3a479948e94a421c7819ae81))


### Bug Fixes

* **ci:** fix jsonpointer module version and simplify golangci exclusions ([ecf2b3f](https://github.com/AobaIwaki123/lumitree/commit/ecf2b3fedef7f56432f4375d0b51ab76c4c3ad06))
* **ci:** update release-please action to googleapis and configure ArgoCD syncOptions ([#12](https://github.com/AobaIwaki123/lumitree/issues/12)) ([7b7ade3](https://github.com/AobaIwaki123/lumitree/commit/7b7ade3dadbb077545750694b0cfeaf2f8f60506))
* **deps:** lock Go version to 1.23 for golangci-lint compatibility ([74f0eda](https://github.com/AobaIwaki123/lumitree/commit/74f0edaef30c35ffacbff9208a8bcef4bd8c53c1))
* **deps:** lock golang.org/x/text and tools for Go 1.23 compatibility ([2d330d2](https://github.com/AobaIwaki123/lumitree/commit/2d330d2c141c8b9ace1d179159bd6e94ebe97031))
* **deps:** lock kin-openapi to v0.128.0 and oapi-codegen to v2.4.1 for Go 1.23 compatibility ([edcdfdb](https://github.com/AobaIwaki123/lumitree/commit/edcdfdb0f84b0e05cef89276e8f957cb806aece8))
* **deps:** lock x/sync and x/mod for Go 1.23 and add verify-all script ([e10f9f5](https://github.com/AobaIwaki123/lumitree/commit/e10f9f5494522a28d6a12ef0b254edcb92179143))
* **docs:** fix Mermaid diagram syntax in architecture.md ([d85e1c5](https://github.com/AobaIwaki123/lumitree/commit/d85e1c55e99123a0031d789895c81a660a961e04))
* **lint:** address all remaining revive linter warnings ([04d630b](https://github.com/AobaIwaki123/lumitree/commit/04d630b57b99e52ac689e98b4294258cbb229197))
* **lint:** configure golangci-lint exclusions and fix main.go comments and error handling ([610558b](https://github.com/AobaIwaki123/lumitree/commit/610558b75c52c523c2970f2cf467115135bba5da))
* **test:** use t.Setenv in config_test.go to fix linter and disable go.sum cache ([c9bdcaf](https://github.com/AobaIwaki123/lumitree/commit/c9bdcaf567ffd539d55e96d2a1660741cd0ca089))
