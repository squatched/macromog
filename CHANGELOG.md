# Changelog

All notable changes to Macromog will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
