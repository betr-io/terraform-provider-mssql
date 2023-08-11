# Terraform Provider `mssql`

> :warning: NOTE: Because the provider as it stands covers all of our current use cases, we will not be dedicating much time and effort to supporting it. We will, however, gladly accept pull requests. We will try to review and release those in a timely manner. Pull requests with included tests and documentation will be prioritized.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x
- [Go](https://golang.org/doc/install) 1.18 (to build the provider plugin)

## Usage

```hcl
terraform {
  required_version = "~> 0.13"
  required_providers {
    mssql = {
      versions = "~> 0.2.2"
      source = "betr.io/betr/mssql"
    }
  }
}

provider "mssql" {}
```

## Building the provider

Clone the repository

```shell
git clone git@github.com:betr-io/terraform-provider-mssql
```

Enter the provider directory and build the provider

```shell
cd terraform-provider-mssql
make build
```

To build and install the provider locally

```shell
make install
```

## Developing the provider

If you wish to work on the provider, you'll first need [Go](https://www.golang.org) installed on your machine (version 1.18+).

To compile the provider, run `make build`. This will build the provider.

To run the unit test, you can simply run `make test`.

To run acceptance tests against a local SQL Server running in Docker, you must have [Docker](https://docs.docker.com/get-docker/) installed. You can then run the following commands

```shell
make docker-start
TESTARGS=-count=1 make testacc-local
make docker-stop
```

This will spin up a SQL server running in a container on your local machine, run the tests that can run against a SQL Server, and destroy the container.

In order to run the full suite of acceptance tests, run `make testacc`. Again, to spin up a local SQL Server container in docker, and corresponding resources in Azure, modify `test-fixtures/all/terraform.tfvars` to match your environment and run

```shell
make azure-create
TESTARGS=-count=1 make testacc
make azure-destroy
```

> **NOTE**: This will create resources in Azure and _will_ incur costs.
>
> **Note to self**: Remember to set current IP address in `test-fixtures/all/terraform.tfvars`, and activate `Global Administrator` in PIM to run Azure tests.

## Release provider

To create a release, do:

- Update `CHANGELOG.md`.
- Update `VERSION` in `Makefile` (only used for installing the provider when developing).
- Push a new valid version tag (e.g. `v1.2.3`) to GitHub.
- See also [Publishing Providers](https://www.terraform.io/docs/registry/providers/publishing.html).
