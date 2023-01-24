terraform {
  required_version = "~> 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4"
    }
  }
}


####################
# DATA SOURCES
####################

# add id parameter to specify a VPC, in my case I'm just testing using the default vpc
data "aws_vpc" "test" {}

data "aws_subnets" "test" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.test.id]
  }
}

data "aws_ami" "ecs_optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-arm64-ebs"]
  }

  owners = ["amazon"]
}

data "aws_ami" "windows_ecs_optimized" {
  most_recent = true

  filter {
    name   = "name"
    values = ["Windows_Server-2022-English-Core-ECS_Optimized-*"]
  }

  owners = ["amazon"]
}

####################
# IAM
####################

# EC2 Instance IAM resources
data "aws_iam_policy_document" "ecs_instance" {
  statement {
    actions = [
      "cloudwatch:PutMetricData",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    actions = [
      "ec2:*",
      "ecs:*",
      "kms:*",
    ]

    resources = [
      "*",
    ]
  }
}

resource "aws_iam_role" "ecs_instance" {
  name = "test-ecsgo-instance-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name   = "instance-policy"
    policy = data.aws_iam_policy_document.ecs_instance.json
  }
}

resource "aws_iam_role_policy_attachment" "instance_ssm" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = "test-ecsgo-instance"
  role = aws_iam_role.ecs_instance.name
}

# ECS Task IAM resources
resource "aws_iam_role" "ecs_task" {
  name = "test-ecsgo-task-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "task_ssm" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy_attachment" "task_ecs" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

####################
# CLOUDWATCH
####################
resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/ecs/test-ecsgo"
}

####################
# EC2
####################
resource "aws_security_group" "ecs_instance" {
  name        = "test-ecsgo-instance-sg"
  description = "Security group for ecsgo test environment"
  vpc_id      = data.aws_vpc.test.id

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_launch_template" "ecs_instance" {
  name                                 = "test-ecsgo-instance-lt"
  image_id                             = data.aws_ami.ecs_optimized.image_id
  instance_type                        = "t4g.medium"
  ebs_optimized                        = true
  instance_initiated_shutdown_behavior = "terminate"

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "test-ecsgo-instance"
    }
  }

  iam_instance_profile {
    name = aws_iam_instance_profile.ecs_instance.id
  }

  monitoring {
    enabled = true
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.ecs_instance.id]
  }

  user_data = base64encode(<<-EOF
    #!/usr/bin/env bash

    cat <<EOF >> /etc/ecs/ecs.config
    ECS_CLUSTER=${aws_ecs_cluster.test.name}
    ECS_ENABLE_CONTAINER_METADATA=true
    ECS_CONTAINER_INSTANCE_PROPAGATE_TAGS_FROM=ec2_instance
    EOF
  )

  instance_market_options {
    market_type = "spot"
  }
}

resource "aws_autoscaling_group" "ecs" {
  name                = "test-ecsgo-instance-asg"
  vpc_zone_identifier = data.aws_subnets.test.ids
  max_size            = 1
  min_size            = 0
  desired_capacity    = 1

  launch_template {
    id      = aws_launch_template.ecs_instance.id
    version = "$Latest"
  }

  health_check_grace_period = 300
  health_check_type         = "EC2"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_launch_template" "windows_ecs_instance" {
  name                                 = "test-ecsgo-windows-instance-lt"
  image_id                             = data.aws_ami.windows_ecs_optimized.image_id
  instance_type                        = "m5.large"
  ebs_optimized                        = true
  instance_initiated_shutdown_behavior = "terminate"

  block_device_mappings {
    device_name = "/dev/sda1"
    ebs {
      volume_size = 100
    }
  }

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "test-ecsgo-windows-instance"
    }
  }

  iam_instance_profile {
    name = aws_iam_instance_profile.ecs_instance.id
  }

  monitoring {
    enabled = true
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.ecs_instance.id]
  }

  user_data = base64encode(<<-EOF
  <powershell>
  Initialize-ECSAgent -Cluster ${aws_ecs_cluster.windows_test.name} -EnableTaskIAMRole -AwsvpcBlockIMDS -EnableTaskENI -LoggingDrivers '["json-file","awslogs"]'
  [Environment]::SetEnvironmentVariable("ECS_ENABLE_AWSLOGS_EXECUTIONROLE_OVERRIDE",$TRUE, "Machine")
  </powershell>
  EOF
  )

  instance_market_options {
    market_type = "spot"
  }
}

resource "aws_autoscaling_group" "windows_ecs" {
  name                = "test-ecsgo-windows-instance-asg"
  vpc_zone_identifier = data.aws_subnets.test.ids
  max_size            = 1
  min_size            = 0
  desired_capacity    = 1

  launch_template {
    id      = aws_launch_template.windows_ecs_instance.id
    version = "$Latest"
  }

  health_check_grace_period = 300
  health_check_type         = "EC2"

  lifecycle {
    create_before_destroy = true
  }
}

####################
# ECS
####################
resource "aws_ecs_cluster" "test" {
  name = "test-ecsgo"
}

resource "aws_ecs_cluster" "windows_test" {
  name = "test-windows-ecsgo"
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

resource "aws_ecs_task_definition" "fargate" {
  family                   = "test-ecsgo-fargate"
  task_role_arn            = aws_iam_role.ecs_task.arn
  execution_role_arn       = aws_iam_role.ecs_task.arn
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 1024
  memory                   = 2048

  container_definitions = jsonencode([
    {
      name      = "nginx"
      image     = "nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
    {
      name      = "redis"
      image     = "redis:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 6379
          hostPort      = 6379
      }]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
    {
      name      = "rabbitmq"
      image     = "rabbitmq:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 5672
          hostPort      = 5672
      }]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

resource "aws_ecs_task_definition" "ec2_launch" {
  family                   = "test-ecsgo-ec2-launch"
  task_role_arn            = aws_iam_role.ecs_task.arn
  execution_role_arn       = aws_iam_role.ecs_task.arn
  requires_compatibilities = ["EC2"]

  runtime_platform {
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name      = "nginx"
      image     = "nginx:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 80
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
    {
      name      = "redis"
      image     = "redis:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 6379
      }]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    },
    {
      name      = "rabbitmq"
      image     = "rabbitmq:latest"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 5672
      }]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

resource "aws_ecs_task_definition" "windows_ec2_launch" {
  family                   = "test-ecsgo-windows-ec2-launch"
  task_role_arn            = aws_iam_role.ecs_task.arn
  requires_compatibilities = ["EC2"]

  container_definitions = jsonencode([
    {
      name      = "iis"
      image     = "mcr.microsoft.com/windows/servercore/iis:windowsservercore-ltsc2022"
      cpu       = 1024
      memory    = 2048
      essential = true
      portMappings = [
        {
          containerPort = 80
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group" : aws_cloudwatch_log_group.test.id
          "awslogs-region" : "eu-west-1"
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

resource "aws_ecs_service" "fargate" {
  name                   = "fargate-test"
  cluster                = aws_ecs_cluster.test.id
  task_definition        = aws_ecs_task_definition.fargate.arn
  desired_count          = 1
  enable_execute_command = true
  launch_type            = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.test.ids
    assign_public_ip = true
    security_groups  = [aws_security_group.ecs_instance.id]
  }

  depends_on = [
    aws_ecs_cluster_capacity_providers.test
  ]
}

resource "aws_ecs_service" "ec2_launch" {
  name                   = "ec2-launch-test"
  cluster                = aws_ecs_cluster.test.id
  task_definition        = aws_ecs_task_definition.ec2_launch.arn
  desired_count          = 1
  enable_execute_command = true
  launch_type            = "EC2"
}

resource "aws_ecs_service" "windows-ec2_launch" {
  name                   = "ec2-windows-launch-test"
  cluster                = aws_ecs_cluster.windows_test.id
  task_definition        = aws_ecs_task_definition.windows_ec2_launch.arn
  desired_count          = 1
  enable_execute_command = true
  launch_type            = "EC2"
}

