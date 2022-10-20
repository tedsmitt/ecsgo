# add id parameter to specify a VPC
data "aws_vpc" "test" {}

data "aws_subnets" "test" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.test.id]
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "/aws/ecs/test-ecsgo"
}

resource "aws_iam_role" "test" {
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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role" "test_exec" {
  name = "test-ecsgo-execution-role"
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

resource "aws_iam_role_policy_attachment" "test_exec" {
  role       = aws_iam_role.test_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_ecs_cluster" "test" {
  name = "test-ecsgo"
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

resource "aws_ecs_task_definition" "test" {
  family                   = "test"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  network_mode             = "awsvpc"
  task_role_arn            = aws_iam_role.test.arn
  execution_role_arn       = aws_iam_role.test_exec.arn
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
    }
  ])

}

resource "aws_ecs_service" "test" {
  name                   = "test"
  cluster                = aws_ecs_cluster.test.id
  task_definition        = aws_ecs_task_definition.test.arn
  desired_count          = 1
  enable_execute_command = true

  network_configuration {
    subnets          = data.aws_subnets.test.ids
    assign_public_ip = true
  }
}
