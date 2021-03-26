# ecsgo

Heavily inspired by incredibly useful [gossm](https://github.com/gjbae1212/gossm), this tool makes use of the new [ECS ExecuteCommand API](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/) to connect to running ECS tasks. It provides an interactive prompt to select your cluster, task and container (if only one container in the task it will default to this), and opens a connection to it.

That's it! Nothing fancy.

> ⚠️ The ExecuteCommand API is quite new at time of creation and existing Services and Tasks may need to be updated/created with the `--enable-execute-command` flag via the CLI. Terraform support for this option is [now available](https://github.com/hashicorp/terraform-provider-aws/pull/18347))

## Usage
The tool uses your AWS Config/Environment Variables to run. If you aren't familiar with working on AWS via the CLI, you can read more about how to configure your environment [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)

| Flag        | Description | Default Value |
| ----------- | ----------- | ------------- |
| `-c`        | Specify the command to be run on the container, defaults to |`/bin/sh`|
| `-r`        | Specify the AWS region to run in                            | N/A

In future releases there will be more flags that will allow you to narrow down the results so you can find your desired task/container quicker.

