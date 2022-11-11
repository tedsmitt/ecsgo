# ecsgo

Heavily inspired by incredibly useful [gossm](https://github.com/gjbae1212/gossm), this tool makes use of the new [ECS ExecuteCommand API](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/) to connect to running ECS tasks. It provides an interactive prompt to select your cluster, task and container (if only one container in the task it will default to this), and opens a connection to it.

That's it! Nothing fancy.

## Installation

#### MacOS/Homebrew

```
brew tap tedsmitt/ecsgo
brew install ecsgo
```

#### Linux

```
wget https://github.com/tedsmitt/ecsgo/releases/latest/download/ecsgo_Linux_x86_64.tar.gz
tar xzf ecsgo_*.tar.gz
```

Move the `ecsgo` binary into your `$PATH`

## Pre-reqs

#### session-manager-plugin

This tool makes use of the [session-manager-plugin](https://github.com/aws/session-manager-plugin). For instructions on how to install, please check out https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html.

MacOS users can alternatively install this via Homebrew:
`brew install --cask session-manager-plugin`

#### Infrastructure

You'll need to follow the prerequisites for ECS Exec as outlined in the [blog post](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/).

You can also view some additional documentation on using ECS Exec [here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-exec.html).

## Usage

| Flag | Description                                                                                  | Default Value              |
| ---- | -------------------------------------------------------------------------------------------- | -------------------------- |
| `-p` | Specify the profile to load the credentials                                                  | `default`                  |
| `-c` | Specify the command to be run on the container (default will change depending on OS family). | `/bin/sh`,`powershell.exe` |
| `-r` | Specify the AWS region to run in                                                             | N/A                        |

The tool also supports AWS Config/Environment Variables for configuration. If you aren't familiar with working on AWS via the CLI, you can read more about how to configure your environment [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

#### See it in action below

![ecsgo0 2 0](https://user-images.githubusercontent.com/25430401/114218136-ef8f7b00-9960-11eb-9c3f-b353ae0ff7ca.gif)

#### Why would I use this over something like AWS Copilot?

At this moment in time copilot only supports connecting to resources that are created and/or managed by the copilot CLI. This tool allows you to leverage ECS Exec easily with your existing resources, and plugs the gap until you are able to do the same with Copilot.
