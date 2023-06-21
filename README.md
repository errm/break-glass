# multipass

Multipass is a simple tool to manage short lived AWS credentials.

It will assume a role(s) optionally with MFA authentication and save those
temporary credentials to an AWS profile.

## Usage

Multipass is configured with an additional config file in the `~/.aws` directory
`~/.aws/multipass`

e.g.

```ini
[admin]
aws_role_arn =  arn:aws:iam::012345678901:role/Admin
aws_mfa_device = arn:aws:iam::012345678901:mfa/iphone
duration = 3600
```

When `multipass` is run, if a MFA device is configured it will request
a token, then temporary credentials for the named role(s) will
be written to the `~/.aws/credentials` file.
