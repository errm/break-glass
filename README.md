# break-glass

break-glass is a simple tool to manage short lived AWS credentials.

![break glass for key](https://upload.wikimedia.org/wikipedia/commons/thumb/0/02/Sign_-_Key_-_Glass_%284891398099%29.jpg/319px-Sign_-_Key_-_Glass_%284891398099%29.jpg)

It will assume roles (optionally) with MFA authentication and save those
temporary credentials to an AWS profile.

## Usage

break-glass is configured with an additional config file in the `~/.aws` directory
`~/.aws/break-glass`

e.g.

```ini
[admin]
aws_role_arn =  arn:aws:iam::012345678901:role/Admin
aws_mfa_device = arn:aws:iam::012345678901:mfa/iphone
duration = 3600

[on-call]
aws_role_arn =  arn:aws:iam::012345678901:role/OnCall
aws_mfa_device = arn:aws:iam::012345678901:mfa/iphone
duration = 3600
```

When `break-glass` is run, if a MFA device is configured it will request
a token, then temporary credentials for the named profile(s) will
be written to the `~/.aws/credentials` file.

If you have more than one profile in your `~/.aws/break-glass` credentials
will be created for all profiles in the file, unless you set the `--profiles` flag to
target only the profile(s) that you want credentials for!
