# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.7] - 2022-12-16

### Fixed

- Fix concurrency issue on user create/update. [PR #52](https://github.com/betr-io/terraform-provider-mssql/pull/52). Closes [#31](https://github.com/betr-io/terraform-provider-mssql/issues/31]. Thanks to [Isabel Andrade](https://github.com/beandrad) for the PR.
- Fix role reorder update issue. [PR #53](https://github.com/betr-io/terraform-provider-mssql/pull/53). Closes [#46](https://github.com/betr-io/terraform-provider-mssql/issues/46). Thanks to [Paul Brittain](https://github.com/paulbrittain) for the PR.

## [0.2.6] - 2022-11-25

### Added

- Support two of the auth forms available through the new [fedauth](https://github.com/denisenkom/go-mssqldb#azure-active-directory-authentication): `ActiveDirectoryDefault` and `ActiveDirectoryManagedIdentity` (because user-assigned identity) as these are the most useful variants. [PR #42](https://github.com/betr-io/terraform-provider-mssql/pull/42). Closes [#30](https://github.com/betr-io/terraform-provider-mssql/issues/30). Thanks to [Bittrance](https://github.com/bittrance) for the PR.
- Improve docs on managed identities. [PR #39](https://github.com/betr-io/terraform-provider-mssql/pull/36). Thanks to [Alexander Guth](https://github.com/alxy) for the PR.

## [0.2.5] - 2022-06-03

### Added

- Add SID as output attribute to the `mssql_user` resource. [PR #36](https://github.com/betr-io/terraform-provider-mssql/pull/36). Closes [#35](https://github.com/betr-io/terraform-provider-mssql/issues/35). Thanks to [rjbell](https://github.com/rjbell) for the PR.

### Changed

- Treat `password` attribute of `mssql_user` as sensitive. Closes [#37](https://github.com/betr-io/terraform-provider-mssql/issues/37).
- Fully qualify package name with Github repository. [PR #38](https://github.com/betr-io/terraform-provider-mssql/pull/38). Thanks to [Ewan Noble](https://github.com/EwanNoble) for the PR.
- Upgraded to go version 1.18
- Upgraded dependencies.
- Upgraded dependencies in test fixtures.

### Fixed

- Only get sql logins if user is not external. [PR #33](https://github.com/betr-io/terraform-provider-mssql/pull/33). Closes [#32](https://github.com/betr-io/terraform-provider-mssql/issues/32). Thanks to [Alexander Guth](https://github.com/alxy) for the PR.

## [0.2.4] - 2021-11-15

Thanks to [Richard Lavey](https://github.com/rlaveycal) ([PR #24](https://github.com/betr-io/terraform-provider-mssql/pull/24)).

### Fixed

- Race condition with String_Split causes failure ([#23](https://github.com/betr-io/terraform-provider-mssql/issues/23))

## [0.2.3] - 2021-09-16

Thanks to [Matthis Holleville](https://github.com/matthisholleville) ([PR #17](https://github.com/betr-io/terraform-provider-mssql/pull/17)), and [bruno-motacardoso](https://github.com/bruno-motacardoso) ([PR #14](https://github.com/betr-io/terraform-provider-mssql/pull/14)).

### Changed

- Add string split function, which should allow the provider to work on SQL Server 2014 (#17).
- Improved documentation (#14).

## [0.2.2] - 2021-08-24

### Changed

- Upgraded to go version 1.17.
- Upgraded dependencies.
- Upgraded dependencies in test fixtures.

## [0.2.1] - 2021-04-30

Thanks to [Anders Båtstrand](https://github.com/anderius) ([PR #8](https://github.com/betr-io/terraform-provider-mssql/pull/8), [PR #9](https://github.com/betr-io/terraform-provider-mssql/pull/9))

### Changed

- Upgrade go-mssqldb to support go version 1.16.

### Fixed

- Cannot create user because of conflicting collation. ([#6](https://github.com/betr-io/terraform-provider-mssql/issues/6))

## [0.2.0] - 2021-04-06

When it is not possible to give AD role: _Directory Readers_ to the Sql Server Identity or an AD Group, use *object_id* to add external user.

Thanks to [Brice Messeca](https://github.com/smag-bmesseca) ([PR #1](https://github.com/betr-io/terraform-provider-mssql/pull/1))

### Added

- Optional object_id attribute to mssql_user

## [0.1.1] - 2020-11-17

Update documentation and examples.

## [0.1.0] - 2020-11-17

Initial release.

### Added

- Resource `mssql_login` to manipulate logins to a SQL Server.
- Resource `mssql_user` to manipulate users in a SQL Server database.

[Unreleased]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.7...HEAD
[0.2.7]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.6...v0.2.7
[0.2.6]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/betr-io/terraform-provider-mssql/releases/tag/v0.1.0
