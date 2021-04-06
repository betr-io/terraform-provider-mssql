# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.1...HEAD
[0.2.0]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/betr-io/terraform-provider-mssql/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/betr-io/terraform-provider-mssql/releases/tag/v0.1.0
