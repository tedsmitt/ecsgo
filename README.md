# ecsgo

Heavily inspired by the incredibly useful [gossm](https://github.com/gjbae1212/gossm), this tool makes use of the new [ECS ExecuteCommand API](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/) to connect to running ECS tasks. It provides an interactive prompt to select your cluster, task and container (if only one container in the task it will default to this), and opens a connection to it. You can also use it to port-forward to containers within your tasks.

That's it! Nothing fancy.

### Installation

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

### Pre-requisites

#### session-manager-plugin

This tool makes use of the [session-manager-plugin](https://github.com/aws/session-manager-plugin). For instructions on how to install, please check out https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html.

MacOS users can alternatively install this via Homebrew:
`brew install --cask session-manager-plugin`

#### Infrastructure

Use [ecs-exec-checker](https://github.com/aws-containers/amazon-ecs-exec-checker) to check for the pre-requisites to use ECS exec.

### Usage

By default, the tool will prompt you to interactively select which cluster, service, task and container to connect to. You can change the behaviour using the flags detailed below:
| Long | Short | Description | Default Value |
| -------------- | ----- | --------------------------------------------------------------------------------------------------------- | -------------------------- |
| `--cluster` | `-n` | Specify the ECS cluster name | N/A |
| `--service` | `-s` | Specify the ECS service name | N/A |
| `--task` | `-t` | Specify the ECS Task ID | N/A |
| `--container` | `-u` | Specify the container name in the ECS Task (if task only has one container this will selected by default) | N/A |
| `--cmd` | `-c` | Specify the command to be run on the container (default will change depending on OS family). | `/bin/sh`,`powershell.exe` |
| `--forward` | `-f` | Port-forward to the container (Remote port will be taken from task/container definitions) | `false` |
| `--local-port` | `-l` | Specify local port to forward (will prompt if not specified) | N/A |
| `--profile` | `-p` | Specify the profile to load the credentials | `default` |
| `--region` | `-r` | Specify the AWS region to run in | N/A |
| `--quiet` | `-q` | Disable output detailing the Cluster/Service/Task information | `false` |
| `--aws-endpoint-url` | `-e` | Specify the AWS endpoint used for all service requests | N/A |
| `--enable-env` | `-v` | Enable ENV population of cli args | `false` |



The tool also supports AWS Config/Environment Variables for configuration. If you aren't familiar with working on AWS via the CLI, you can read more about how to configure your environment [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

##### See it in action below

![ecsgo0 2 0](https://user-images.githubusercontent.com/25430401/114218136-ef8f7b00-9960-11eb-9c3f-b353ae0ff7ca.gif)

