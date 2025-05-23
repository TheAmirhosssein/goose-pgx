# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v3.24.3]

- Add `GOOSE_TABLE` environment variable -- lower priority than `-table` flag, but higher than the
  default table name. (#932)
- Dependency updates

## [v3.24.2]

- Add `TableExists` table existence check for the mysql dialect (#895)
- Upgrade **minimum Go version to 1.23**
- Various dependency updates

## [v3.24.1]

- Fix regression (`v3.23.1` and `v3.24.0`) in postgres migration table existence check for
  non-default schema. (#882, #883, #884).

## [v3.24.0]

- Add support for loading environment variables from `.env` files, enabled by default.
  - The default file name is `.env`, but can be changed with the `-env=<filename>` flag.
  - To disable this feature, set `-env=none`.

## [v3.23.1]

- Store implementations can **optionally** implement the `TableExists` method to provide optimized
  table existence checks (#860)
  - Default postgres Store implementation updated to use `pg_tables` system catalog, more to follow
  - Backward compatible change - existing implementations will continue to work without modification

```go
TableExists(ctx context.Context, db database.DBTxConn) (bool, error)
```

## [v3.23.0]

- Add `WithLogger` to `NewProvider` to allow custom loggers (#833)
- Update Provider `WithVerbose` behavior to log all SQL statements (#851)
- Upgrade dependencies and rebuild binaries with latest Go version (`go1.23.3`)

## [v3.22.1]

- Upgrade dependencies and rebuild binaries with latest Go version (`go1.23.1`)

## [v3.22.0]

- Minimum Go version is now 1.21
- Add Unwrap to PartialError (#815)
- Allow flags anywhere on the CLI (#814)

`goose` uses the default Go `flag` parsing library, which means flags **must** be defined before the
first positional argument. We've updated this behavior to allow flags to be defined anywhere. For
more details, see [blog post](https://mfridman.com/blog/2024/allowing-flags-anywhere-on-the-cli/).

- Update `WithDisableGlobalRegistry` behavior (#783). When set, this will ignore globally-registered
  migrationse entirely instead of the previous behavior of raising an error. Specifically, the
  following check is removed:

```go
if len(global) > 0 {
	return nil, errors.New("global registry disabled, but provider has registered go migrations")
}
```

This enables creating isolated goose provider(s) in legacy environments where global migrations may
be registered. Without updating this behavior, it would be impossible to use
`WithDisableGlobalRegistry` in combination with provider-scoped `WithGoMigrations`.

- Postgres, updated schema to use identity instead of serial and make `tstamp` not nullable (#556)

```diff
- id serial NOT NULL,
+ id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,

- tstamp timestamp NULL default now(),
+ tstamp timestamp NOT NULL DEFAULT now()
```

- MySQL, updated schema to not use SERIAL alias (#816)

```diff
- id serial NOT NULL,
+ id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
```

## [v3.21.1]

- Add `GetVersions` method to `goose.Provider`, returns the current (max db) version and the latest
  (max filesystem) version. (#756)
- Clarify `GetLatestVersion` method MUST return `ErrVersionNotFound` if no latest migration is
  found. Previously it was returning a -1 and nil error, which was inconsistent with the rest of the
  API surface.

- Add `GetLatestVersion` implementations to all existing dialects. This is an optimization to avoid
  loading all migrations when only the latest version is needed. This uses the `max` function in SQL
  to get the latest version_id irrespective of the order of applied migrations.
  - Refactor existing portions of the code to use the new `GetLatestVersion` method.

## [v3.21.0]

- Retracted. Broken release, please use v3.21.1 instead.

## [v3.20.0]

- Expand the `Store` interface by adding a `GetLatestVersion` method and make the interface public.
- Add a (non-blocking) method to check if there are pending migrations to the `goose.Provider`
  (#751):

```go
func (p *Provider) HasPending(context.Context) (bool, error) {}
```

The underlying implementation **does not respect the `SessionLocker`** (if one is enabled) and can
be used to check for pending migrations without blocking or being blocked by other operations.

- The methods `.Up`, `.UpByOne`, and `.UpTo` from `goose.Provider` will invoke `.HasPending` before
  acquiring a lock with `SessionLocker` (if enabled). This addresses an edge case in
  Kubernetes-style deployments where newer pods with long-running migrations prevent older pods -
  which have all known migrations applied - from starting up due to an advisory lock. For more
  details, refer to https://github.com/TheAmirhosssein/goose/pull/507#discussion_r1266498077 and #751.
- Move integration tests to `./internal/testing` and make it a separate Go module. This will allow
  us to have a cleaner top-level go.mod file and avoid imports unrelated to the goose project. See
  [integration/README.md](https://github.com/TheAmirhosssein/goose/blob/d0641b5bfb3bd5d38d95fe7a63d7ddf2d282234d/internal/testing/integration/README.md)
  for more details. This shouldn't affect users of the goose library.

## [v3.19.2] - 2024-03-13

- Remove duckdb support. The driver uses Cgo and we've decided to remove it until we can find a
  better solution. If you were using duckdb with goose, please let us know by opening an issue.

## [v3.19.1] - 2024-03-11

- Fix selecting dialect for `redshift`
- Add `GOOSE_MIGRATION_DIR` documentation
- Bump github.com/opencontainers/runc to `v1.1.12` (security fix)
- Update CI tests for go1.22
- Make goose annotations case-insensitive
  - All `-- +goose` annotations are now case-insensitive. This means that `-- +goose Up` and `--
+goose up` are now equivalent. This change was made to improve the user experience and to make the
    annotations more consistent.

## [v3.19.0] - 2024-03-11

- Use [v3.19.1] instead. This was tagged but not released and does not contain release binaries.

## [v3.18.0] - 2024-01-31

- Add environment variable substitution for SQL migrations. (#604)

  - This feature is **disabled by default**, and can be enabled by adding an annotation to the
    migration file:

    ```sql
    -- +goose ENVSUB ON
    ```

  - When enabled, goose will attempt to substitute environment variables in the SQL migration
    queries until the end of the file, or until the annotation `-- +goose ENVSUB OFF` is found. For
    example, if the environment variable `REGION` is set to `us_east_1`, the following SQL migration
    will be substituted to `SELECT * FROM regions WHERE name = 'us_east_1';`

    ```sql
    -- +goose ENVSUB ON
    -- +goose Up
    SELECT * FROM regions WHERE name = '${REGION}';
    ```

- Add native [Turso](https://turso.tech/) support with libsql driver. (#658)

- Fixed query for list migrations in YDB (#684)

## [v3.17.0] - 2023-12-15

- Standardised the MIT license (#647)
- Improve provider `Apply()` errors, add `ErrNotApplied` when attempting to rollback a migration
  that has not been previously applied. (#660)
- Add `WithDisableGlobalRegistry` option to `NewProvider` to disable the global registry. (#645)
- Add `-timeout` flag to CLI to set the maximum allowed duration for queries to run. Default remains
  no timeout. (#627)
- Add optional logging in `Provider` when `WithVerbose` option is supplied. (#668)

⚠️ Potential Breaking Change ⚠️

- Update `goose create` to use UTC time instead of local time. (#242)

## [v3.16.0] - 2023-11-12

- Added YDB support. (#592)
- Fix sqlserver query to ensure DB version. (#601)
- Allow setting / resetting the global Go migration registry. (#602)
  - `SetGlobalMigrations` and `ResetGlobalMigrations` functions have been added.
  - Introduce `NewGoMigration` for constructing Go migrations.
- Add initial implementation of `goose.NewProvider`.

🎉 Read more about this new feature here:

https://TheAmirhosssein.github.io/goose/blog/2023/goose-provider/

The motivation behind the Provider was simple - to reduce global state and make goose easier to
consume as an imported package.

Here's a quick summary:

- Avoid global state
- Make Provider safe to use concurrently
- Unlock (no pun intended) new features, such as database locking
- Make logging configurable
- Better error handling with proper return values
- Double down on Go migrations
- ... and more!

## [v3.15.1] - 2023-10-10

- Fix regression that prevented registering Go migrations that didn't have the corresponding files
  available in the filesystem. (#588)
  - If Go migrations have been registered globally, but there are no .go files in the filesystem,
    **always include** them.
  - If Go migrations have been registered, and there are .go files in the filesystem, **only
    include** those migrations. This was the original motivation behind #553.
  - If there are .go files in the filesystem but not registered, **raise an error**. This is to
    prevent accidentally adding valid looking Go migration files without explicitly registering
    them.

## [v3.15.0] - 2023-08-12

- Fix `sqlparser` to avoid skipping the last statement when it's not terminated with a semicolon
  within a StatementBegin/End block. (#580)
- Add `**go1.21**` to the CI matrix.
- Bump minimum version of module in go.mod to `go1.19`.
- Fix version output when installing pre-built binaries (#585).

## [v3.14.0] - 2023-07-26

- Filter registered Go migrations from the global map with corresponding .go files from the
  filesystem.
  - The code previously assumed all .go migrations would be in the same folder, so this should not
    be a breaking change.
  - See #553 for more details
- Improve output log message for applied up migrations. #562
- Fix an issue where `AddMigrationNoTxContext` was registering the wrong source because it skipped
  too many frames. #572
- Improve binary version output when using go install.

## [v3.13.4] - 2023-07-07

- Fix pre-built binary versioning and make small improvements to GoReleaser config.
- Fix an edge case in the `sqlparser` where the last up statement may be ignored if it's
  unterminated with a semicolon and followed by a `-- +goose Down` annotation.
- Trim `Logger` interface to `Printf` and `Fatalf` methods only. Projects that have previously
  implemented the `Logger` interface should not be affected, and can remove unused methods.

## [v3.13.1] - 2023-07-03

- Add pre-built binaries with GoReleaser and update the build process.

## [v3.13.0] - 2023-06-29

- Add a changelog to the project, based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
- Update go.mod and retract all `v3.12.X` tags. They were accidentally pushed and contain a
  reference to the wrong Go module.
- Fix `up` and `up -allowing-missing` behavior.
- Fix empty version in log output.
- Add new `context.Context`-aware functions and methods, for both sql and go migrations.
- Return error when no migration files found or dir is not a directory.

[Unreleased]: https://github.com/pressly/goose/compare/v3.24.3...HEAD
[v3.24.3]: https://github.com/pressly/goose/compare/v3.24.2...v3.24.3
[v3.24.2]: https://github.com/pressly/goose/compare/v3.24.1...v3.24.2
[v3.24.1]: https://github.com/pressly/goose/compare/v3.24.0...v3.24.1
[v3.24.0]: https://github.com/pressly/goose/compare/v3.23.1...v3.24.0
[v3.23.1]: https://github.com/pressly/goose/compare/v3.23.0...v3.23.1
[v3.23.0]: https://github.com/pressly/goose/compare/v3.22.1...v3.23.0
[v3.22.1]: https://github.com/pressly/goose/compare/v3.22.0...v3.22.1
[v3.22.0]: https://github.com/pressly/goose/compare/v3.21.1...v3.22.0
[v3.21.1]: https://github.com/pressly/goose/compare/v3.20.0...v3.21.1
[v3.21.0]: https://github.com/pressly/goose/compare/v3.20.0...v3.21.0
[v3.20.0]: https://github.com/pressly/goose/compare/v3.19.2...v3.20.0
[v3.19.2]: https://github.com/pressly/goose/compare/v3.19.1...v3.19.2
[v3.19.1]: https://github.com/pressly/goose/compare/v3.19.0...v3.19.1
[v3.19.0]: https://github.com/pressly/goose/compare/v3.18.0...v3.19.0
[v3.18.0]: https://github.com/pressly/goose/compare/v3.17.0...v3.18.0
[v3.17.0]: https://github.com/pressly/goose/compare/v3.16.0...v3.17.0
[v3.16.0]: https://github.com/pressly/goose/compare/v3.15.1...v3.16.0
[v3.15.1]: https://github.com/pressly/goose/compare/v3.15.0...v3.15.1
[v3.15.0]: https://github.com/pressly/goose/compare/v3.14.0...v3.15.0
[v3.14.0]: https://github.com/pressly/goose/compare/v3.13.4...v3.14.0
[v3.13.4]: https://github.com/pressly/goose/compare/v3.13.1...v3.13.4
[v3.13.1]: https://github.com/pressly/goose/compare/v3.13.0...v3.13.1
[v3.13.0]: https://github.com/pressly/goose/releases/tag/v3.13.0
