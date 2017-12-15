# WolfBeacon Infrastructure Operations
WolfBeacon's AWS Infrastructure Operations and Deployment Automation Service

## How to Build

First you must add two JSON files to `config/` folder: `settings.json` and `users.json`, you can find examples named `xx.example.json`.

Then you have to add a file called `aws_credentials` to `config/` folder, simply duplicate `aws_credentials.example` and modify the content.

To build, simply run `docker build -t <docker image name> .`

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
