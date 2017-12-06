# WolfBeacon Ops
WolfBeacon Operation Automation Software

## AWS Credentials

You must add a file called `config/aws_credentials` before building Docker image.

```
[default]
aws_access_key_id=AWS_ACCESS_KEY
aws_secret_access_key=AWS_SECRET_ACCESS_KEY
```

## How to use Slack bot

### Build

`run build <project name>`

For example: run build wolfbeacon-core-api

After build compelete, a message will be sent to operations channel.

`list builds`

List all current builds.

`delete build <build id>`

Delete a build.

### Projects / Enviroments

`list projects`

List all projects

`list envs`

List all enviroments

`rebuild env <project name> <env name>`

### Examples

Rebuild a image

![image](https://preview.ibb.co/cnfNxw/Screenshot_from_2017_12_06_01_38_51.png)

List enviroments, and update one enviroment to latest app version

<a href="https://ibb.co/jXoHxw"><img src="https://preview.ibb.co/cwYT4b/Screenshot_from_2017_12_06_01_41_28.png" alt="Screenshot_from_2017_12_06_01_41_28" border="0"></a>

<a href="https://imgbb.com/"><img src="https://image.ibb.co/dx36qG/Screenshot_from_2017_12_06_01_42_36.png" alt="Screenshot_from_2017_12_06_01_42_36" border="0"></a>