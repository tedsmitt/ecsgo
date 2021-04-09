# ecsgo

Heavily inspired by incredibly useful [gossm](https://github.com/gjbae1212/gossm), this tool makes use of the new [ECS ExecuteCommand API](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/) to connect to running ECS tasks. It provides an interactive prompt to select your cluster, task and container (if only one container in the task it will default to this), and opens a connection to it.

That's it! Nothing fancy.

> ⚠️ The ExecuteCommand API is quite new at time of creation and existing Services and Tasks may need to be updated/created with the `--enable-execute-command` flag via the CLI. Terraform support for this option is [now available](https://github.com/hashicorp/terraform-provider-aws/pull/18347))

## Prereqs
You'll need to follow the prerequisites for ECS Exec as outlined in the [blog post](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/). You can also view some additional documentation on using ECS Exec [here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-exec.html)

## Installation
#### MacOS/Homebrew
```
brew tap tedsmitt/ecsgo
brew install ecsgo
```

#### Linux
```
wget https://github.com/tedsmitt/ecsgo/releases/download/0.2.0/ecsgo_Linux_x86_64.tar.gz
tar xzf ecsgo_*.tar.gz
```
Move the `ecsgo` binary into your `$PATH`

## Usage
The tool uses your AWS Config/Environment Variables to run. If you aren't familiar with working on AWS via the CLI, you can read more about how to configure your environment [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)

| Flag        | Description | Default Value |
| ----------- | ----------- | ------------- |
| `-p`        | Specify the profile to load the credentials | `default` |
| `-c`        | Specify the command to be run on the container, defaults to |`/bin/sh`|
| `-r`        | Specify the AWS region to run in                            | N/A

In future releases there will be more flags that will allow you to narrow down the results so you can find your desired task/container quicker.

See it in action below:
![ecsgo0 2 0](https://user-images.githubusercontent.com/25430401/114218136-ef8f7b00-9960-11eb-9c3f-b353ae0ff7ca.gif)



