# Changelog

All notable changes to Macromog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 (2026-06-27)


### Features

* add lib/ modules (macros, validate, yaml) ([bf32f70](https://github.com/squatched/macromog/commit/bf32f70b46bc495f589a061e51750afa386bc584))
* add macromog.lua main addon entry point ([06b3205](https://github.com/squatched/macromog/commit/06b3205ce15a384e326a5c46a3227216b2c72982))
* align plugin implementation with updated spec ([#6](https://github.com/squatched/macromog/issues/6)) ([6ed0970](https://github.com/squatched/macromog/commit/6ed0970d62b32730a6f634b400719f70eee49b0e))
* character folder aliases ([#19](https://github.com/squatched/macromog/issues/19)) ([00698ed](https://github.com/squatched/macromog/commit/00698ed11855cfaaab34be519ed86daefa144b44))
* **cli:** add --output flag for JSON output ([#18](https://github.com/squatched/macromog/issues/18)) ([4e2b9b0](https://github.com/squatched/macromog/commit/4e2b9b035c8372fdf6bc980e45f6d9870f0514e1))
* **cli:** add export from FFXI macro .dat files ([#11](https://github.com/squatched/macromog/issues/11)) ([671bad0](https://github.com/squatched/macromog/commit/671bad0fe0c55f4a9f3b91122f088201cc9cb853))
* **cli:** colorized text output with TextWriter ([#21](https://github.com/squatched/macromog/issues/21)) ([40d0bd5](https://github.com/squatched/macromog/commit/40d0bd5bf2313f0fadf852090866f26df5a5a17a))
* **cli:** dense flg w/ comment placeholders ([#26](https://github.com/squatched/macromog/issues/26)) ([316894d](https://github.com/squatched/macromog/commit/316894d9de68a692f9dfe6323222052510b64294))
* **cli:** implement backup command ([#16](https://github.com/squatched/macromog/issues/16)) ([28c97c5](https://github.com/squatched/macromog/commit/28c97c5c309d02c0998737084f858fc3db4a2bf9))
* **cli:** implement configuration system ([#25](https://github.com/squatched/macromog/issues/25)) ([aabc6d1](https://github.com/squatched/macromog/commit/aabc6d16b383db08f0073798f550b068b939b0f5))
* **cli:** implement import command ([#12](https://github.com/squatched/macromog/issues/12)) ([5655b7d](https://github.com/squatched/macromog/commit/5655b7dcce59fcc21ada1ebdc1029afcb4262cc6))
* **cli:** implement list command ([#15](https://github.com/squatched/macromog/issues/15)) ([670a198](https://github.com/squatched/macromog/commit/670a19840eff9665c0dd5d28da0c13e72949776f))
* **cli:** migrate to cobra ([#29](https://github.com/squatched/macromog/issues/29)) ([7451598](https://github.com/squatched/macromog/commit/745159865382c31f52fd98883faa6d7d50663fad))
* **cli:** multi-char selection with --all bypass ([#17](https://github.com/squatched/macromog/issues/17)) ([6d5e8e3](https://github.com/squatched/macromog/commit/6d5e8e3e16bfa32b71251a7671cc4604ae20e597))
* **cli:** stub CLI and split validate targets ([#7](https://github.com/squatched/macromog/issues/7)) ([e89e77c](https://github.com/squatched/macromog/commit/e89e77ca9951ecfddd0c52db415a3686d8319c39))
* **plugin:** addon packaging & auto-config ([#27](https://github.com/squatched/macromog/issues/27)) ([4abd118](https://github.com/squatched/macromog/commit/4abd1189490f81f2f003a7730ac3e958d2129fac))
* scope-aware export, import, and template ([#22](https://github.com/squatched/macromog/issues/22)) ([c7b9d6b](https://github.com/squatched/macromog/commit/c7b9d6b2090c7856fea40a5366e05bec0d8ccd3e))
* **validate:** add 1-based book index support ([d6db8d9](https://github.com/squatched/macromog/commit/d6db8d92cdc3fddfbd703d085fa15ecc11f31fa1))
* **validate:** YAML validation and CLI command ([#8](https://github.com/squatched/macromog/issues/8)) ([0e4d64b](https://github.com/squatched/macromog/commit/0e4d64b35aa990e84f05f83bd181e03d62ed6986))
* **yaml:** preserve DAT header unknown field ([#14](https://github.com/squatched/macromog/issues/14)) ([9e6bf01](https://github.com/squatched/macromog/commit/9e6bf01c4a6eeec6967f8275f13a1291e59a07c6))


### Bug Fixes

* **ci:** add release-please with block for pr ([1ee2ef5](https://github.com/squatched/macromog/commit/1ee2ef5d5494ca732a4108b9ae43ea53f3b1099f))
* **ci:** scan untracked files in trailing-ws ([#31](https://github.com/squatched/macromog/issues/31)) ([15c8ea8](https://github.com/squatched/macromog/commit/15c8ea830820359561693d3cfa9deb7c2b7d922a))
* **ci:** update release-please params to v5 ([dbd1c2c](https://github.com/squatched/macromog/commit/dbd1c2c9105b54487eec4d64c41f0277cc138818))
* **cli:** discard build-cli-all output - /dev/null ([546a18d](https://github.com/squatched/macromog/commit/546a18d2892f18ef4694d5634d3bfaea8c1d36d4))
* more dat observational updates ([6b40d10](https://github.com/squatched/macromog/commit/6b40d1000a7cf1b6e1138995baee28a9ea460e17))
* **plugin,cli:** import and backup dest ([#38](https://github.com/squatched/macromog/issues/38)) ([777d4a1](https://github.com/squatched/macromog/commit/777d4a148d36e4a96a8fe654ca85d0870fbe9410))
* **plugin:** five UX improvements from TODO ([#37](https://github.com/squatched/macromog/issues/37)) ([3149571](https://github.com/squatched/macromog/commit/314957189f1b7f059a1b3ba6752f6b3ab7f3ee92))
* **plugin:** no-flash spawn DLL, XDG paths ([#34](https://github.com/squatched/macromog/issues/34)) ([f3182cb](https://github.com/squatched/macromog/commit/f3182cbf26624b8e04e2d60585d539444089f371))
* **scope:** C/A* and scope tests ([#40](https://github.com/squatched/macromog/issues/40)) ([84a2d6f](https://github.com/squatched/macromog/commit/84a2d6f25ef50332b53169e34ad6d4ce0a8a12d6))
* two character-registration bugs ([#35](https://github.com/squatched/macromog/issues/35)) ([918204f](https://github.com/squatched/macromog/commit/918204f24f3a6775d6e453f56394fbf790df46bb))
* updated observational behavior of dats ([eb9891a](https://github.com/squatched/macromog/commit/eb9891a1583ce064333f07bf085bbac26a2d33c5))

## [Unreleased]

### Added
- Initial addon skeleton with command dispatch (`export`, `import`, `validate`, `backup`, `help`)
- Validation module enforcing FFXI macro constraints (40 books, 10 sets, 20 macros, 8-char names, 6 lines)
- YAML serializer/parser stubs (`lib/yaml.lua`)
- Macro memory read/write stubs (`lib/macros.lua`)
- Automatic timestamped backup before any import
- GitHub Actions release workflow with CHANGELOG extraction
- Luacheck linting workflow

[Unreleased]: https://github.com/squatched/macromog/compare/HEAD...HEAD
