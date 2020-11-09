# mssql_login

Description of what this resource does, with links to official
app/service documentation.

## Example Usage

```hcl
resource "mssql_login" "example" {
  server {
    azure_login {}
  }
  login_name = "testlogin"
}
```

## Argument Reference

* `attribute_name` - (Optional/Required) List arguments this resource takes.

## Attribute Reference

* `attribute_name` - List attributes that this resource exports.
