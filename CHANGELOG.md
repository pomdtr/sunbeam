# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1] - 2023-03-12

### Added

- `sunbeam read` command and `read` action now support http/https url !
- All pages now have a default `reload` action

### Changed

- `sunbeam read` do not require the `-` arg to read from stdin anymore
- `--check` flag for the `sunbeam read` and `sunbeam push` commands was replaced by the `sunbeam validate` command

## [0.4.0] - 2023-03-11

### Added

- Add `copy` and `open` command for usage in scripts
- the `read` action now update the dir of the sunbeam processes, so that relative paths are resolved correctly

### Changed

- revert back to the `read` command instead of `push` for both command and action
- onSuccess does not support `copy` and `open` anymore, use the `open` and `copy` commands instead

## [0.3.1] - 2023-03-10

### Changed

- Remove the default config file. Sunbeam now print the help if no `sunbeam.json` file is present in the current directory.

## [0.3.0] - 2023-03-10

### Changed

- `run` action and `push` action are now merged into a single `run` action, with a `onSuccess` property that can be set to `push`
- `push` action is now used to read a static list from a file

## [0.2.5] - 2023-03-10

### Added

- Add new `sunbeam read` cmd to read a sunbeam response from a file or stdin
- sunbeam now sets the `SUNBEAM_RUNNER` environment variable to `true` when running a script

## [0.2.4] - 2023-03-10

### Fixed

- Fix an issue that caused dynamic list to hang
- Dynamic list items are now correctly ordered

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
