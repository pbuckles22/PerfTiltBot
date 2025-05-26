# AWS Deployment Guide

This guide covers deploying PBChatBot to Amazon ECS (Elastic Container Service).

## Prerequisites

1. **AWS CLI**
   ```bash
   # Install AWS CLI
   pip install awscli

   # Configure AWS CLI
   aws configure
   ```

2. **AWS ECS CLI**
   ```bash
   # Install ECS CLI
   pip install ecs-cli

   # Configure ECS CLI
   ecs-cli configure profile --profile-name pbchatbot
   ```

3. **Docker**
   - Docker Desktop (Windows/macOS)
   - Docker Engine (Linux)

## AWS Resources

### Required IAM Permissions

The following IAM permissions are required for the bot to function:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:CreateCluster",
                "ecs:DeleteCluster",
                "ecs:DescribeClusters",
                "ecs:ListClusters",
                "ecs:CreateService",
                "ecs:DeleteService",
                "ecs:DescribeServices",
                "ecs:ListServices",
                "ecs:UpdateService",
                "ecs:CreateTaskDefinition",
                "ecs:DeleteTaskDefinition",
                "ecs:DescribeTaskDefinition",
                "ecs:ListTaskDefinitions",
                "ecs:RunTask",
                "ecs:StopTask",
                "ecs:DescribeTasks",
                "ecs:ListTasks"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:GetRepositoryPolicy",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:DescribeImages",
                "ecr:BatchGetImage",
                "ecr:InitiateLayerUpload",
                "ecr:UploadLayerPart",
                "ecr:CompleteLayerUpload",
                "ecr:PutImage"
            ],
            "Resource": "*"
        }
    ]
}
```

### Required AWS Services

1. **Amazon ECR (Elastic Container Registry)**
   - Repository for Docker images
   - Access via IAM roles

2. **Amazon ECS (Elastic Container Service)**
   - Container orchestration
   - Task definitions
   - Service definitions

3. **Amazon CloudWatch**
   - Log groups
   - Metrics
   - Alarms

## Deployment Steps

1. **Create ECR Repository**
   ```bash
   aws ecr create-repository --repository-name pbchatbot
   ```

2. **Build and Push Docker Image**
   ```bash
   # Build image
   docker build -t pbchatbot .

   # Tag image
   docker tag pbchatbot:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/pbchatbot:latest

   # Push to ECR
   docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/pbchatbot:latest
   ```

3. **Create ECS Cluster**
   ```bash
   ecs-cli up --cluster pbchatbot --instance-type t2.micro --size 1
   ```

4. **Create Task Definition**
   ```bash
   # Create task definition file
   cat > task-definition.json << EOF
   {
       "family": "pbchatbot",
       "containerDefinitions": [
           {
               "name": "pbchatbot",
               "image": "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/pbchatbot:latest",
               "memory": 256,
               "cpu": 256,
               "essential": true,
               "environment": [
                   {
                       "name": "CHANNEL_NAME",
                       "value": "your_channel"
                   }
               ],
               "mountPoints": [
                   {
                       "sourceVolume": "config",
                       "containerPath": "/app/configs",
                       "readOnly": true
                   },
                   {
                       "sourceVolume": "data",
                       "containerPath": "/app/data",
                       "readOnly": false
                   }
               ]
           }
       ],
       "volumes": [
           {
               "name": "config",
               "host": {
                   "sourcePath": "/ecs/config"
               }
           },
           {
               "name": "data",
               "host": {
                   "sourcePath": "/ecs/data"
               }
           }
       ]
   }
   EOF

   # Register task definition
   aws ecs register-task-definition --cli-input-json file://task-definition.json
   ```

5. **Create ECS Service**
   ```bash
   aws ecs create-service \
       --cluster pbchatbot \
       --service-name pbchatbot \
       --task-definition pbchatbot:1 \
       --desired-count 1
   ```

## Monitoring

1. **CloudWatch Logs**
   ```bash
   # View logs
   aws logs get-log-events --log-group-name /ecs/pbchatbot --log-stream-name <stream-name>
   ```

2. **ECS Metrics**
   - CPU utilization
   - Memory utilization
   - Container health

3. **Alarms**
   - Set up CloudWatch alarms for:
     - High CPU usage
     - High memory usage
     - Container failures

## Scaling

1. **Manual Scaling**
   ```bash
   aws ecs update-service --cluster pbchatbot --service pbchatbot --desired-count 2
   ```

2. **Auto Scaling**
   - Configure auto scaling based on:
     - CPU utilization
     - Memory utilization
     - Request count

## Maintenance

1. **Update Bot**
   ```bash
   # Build and push new image
   docker build -t pbchatbot .
   docker tag pbchatbot:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/pbchatbot:latest
   docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/pbchatbot:latest

   # Update service
   aws ecs update-service --cluster pbchatbot --service pbchatbot --force-new-deployment
   ```

2. **Backup Data**
   ```bash
   # Backup data volume
   aws ec2 create-snapshot --volume-id <volume-id> --description "PBChatBot data backup"
   ```

3. **Cleanup**
   ```bash
   # Delete service
   aws ecs delete-service --cluster pbchatbot --service pbchatbot --force

   # Delete cluster
   aws ecs delete-cluster --cluster pbchatbot
   ``` 