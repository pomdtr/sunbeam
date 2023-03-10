# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3] - 2023-03-10

### Changed

- Move inputs to an `inputs` field instead of using the `command` field.

## [0.2.2] - 2023-03-10

### Added

- Automatically generate the global config file if it doesn't exist, with a link to the documentation.

## [0.2.1] - 2023-03-10

### Added

- Add windows binary to release

## [0.2.0] - 2023-03-10

### Added

- `sunbeam` command now read from `~/.config/sunbeam/sunbeam.json` if no `sunbeam.json` is found in the current directory.

### Changed

- add `sunbeam run <script>` command (replace `sunbeam <script>`)

## [0.1.0] - 2023-03-10

### Added

- Initial release
