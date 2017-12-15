# WolfBeacon Infrastructure Operations
WolfBeacon's AWS Infrastructure Operations and Deployment Automation Service

[![Build Status](https://travis-ci.org/wolfbeacon/wolfbeacon-infra-ops.svg?branch=list-projects)](https://travis-ci.org/wolfbeacon/wolfbeacon-infra-ops)

## Build Natively Using Go

First, make sure you have the repository in your `$GOPATH`.

CD into the repository, and run the following commands:

```
go get ./...
go build
```

After successful build, you should see a executable file called `wolfbeacon-infra-ops`.

Then, configure `config/` folder so you have all the right configurations.

Run `./wolfbeacon-infra-ops` to start the bot. If using Windows `./wolfbeacon-infra-ops.exe`.

## Build Using Docker

First you must add two JSON files to `config/` folder: `settings.json` and `users.json`, you can find examples named `xx.example.json`.

Then you have to add a file called `aws_credentials` to `config/` folder, simply duplicate `aws_credentials.example` and modify the content.

To build, simply run `docker build -t <docker image name> .`

## Command List

To execute a command, you must direct metion the bot. (@ the bot at the beginning).

### list projects

Get a list of CodeBuild projects.

### list builds

List last 5 builds.

### list envs

List all enviroments.

### run build {project name}

Start a new build.

### rebuild env {project name} {env name}

Update an enviroment to the latest version.