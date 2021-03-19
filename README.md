# ecsgo

Heavily inspired by incredibly useful [gossm](https://github.com/gjbae1212/gossm), this tool makes use of the new [ECS ExecuteCommand API](https://aws.amazon.com/blogs/containers/new-using-amazon-ecs-exec-access-your-containers-fargate-ec2/) to connect to running ECS tasks. It provides an interactive prompt to select your cluster, task and container (if only one container in the task it will default to this), and opens a connection to it.

That's it! Nothing fancy.

> ⚠️ The ExecuteCommand API is very new at time of creation and existing Services and Tasks need to be updated/created with the `--enable-execute-command` flag via the CLI. Terraform support for this option is currently in progress and can be viewed [here](https://github.com/hashicorp/terraform-provider-aws/issues/18112)

### Flags

| Flag        | Description |
| ----------- | ----------- |
| `-c`        | Specify the command to be run on the container, defaults to `/bin/sh` |