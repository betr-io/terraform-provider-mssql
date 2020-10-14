# Terraform Provider mssql

Run the following command to build the provider

```shell
go build -o terraform-provider-mssql
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```
