# PerfTiltBot

A performance-focused bot for analyzing and providing insights about gameplay tilt.

## Prerequisites

- Go 1.23.1 or higher
- Git

## Setup

1. Clone the repository:
```bash
git clone https://github.com/pbuckles22/PerfTiltBot.git
cd PerfTiltBot
```

2. Install dependencies:
```bash
go mod tidy
```

## Project Structure

```
.
├── cmd/           # Application entry points
├── internal/      # Private application and library code
├── pkg/          # Library code that could be used by external applications
└── configs/      # Configuration files
```

## Development

To run the bot locally:

```bash
go run cmd/bot/main.go
```

## Docker Deployment

To build and run the bot using Docker:

1. **Build the Docker image:**
   ```bash
   docker build -t perftiltbot .
   ```

2. **Run the container:**
   ```bash
   docker run -d --name perftiltbot perftiltbot
   ```

**Note:** Ensure that your `secrets.yaml` file is not committed to version control. It is copied into the Docker image during the build process.

## Amazon ECS Deployment

To deploy the bot on Amazon ECS:

1. **Install the AWS CLI and configure your credentials:**
   ```bash
   aws configure
   ```

2. **Create an ECR repository:**
   ```bash
   aws ecr create-repository --repository-name perftiltbot
   ```

3. **Authenticate Docker to ECR:**
   ```bash
   aws ecr get-login-password --region <your-region> | docker login --username AWS --password-stdin <your-account-id>.dkr.ecr.<your-region>.amazonaws.com
   ```

4. **Tag and push the Docker image to ECR:**
   ```bash
   docker tag perftiltbot:latest <your-account-id>.dkr.ecr.<your-region>.amazonaws.com/perftiltbot:latest
   docker push <your-account-id>.dkr.ecr.<your-region>.amazonaws.com/perftiltbot:latest
   ```

5. **Create an ECS task definition:**
   - Use the AWS Management Console or AWS CLI to create a task definition that uses the ECR image.
   - Example task definition (JSON):
     ```json
     {
       "family": "perftiltbot",
       "containerDefinitions": [
         {
           "name": "perftiltbot",
           "image": "<your-account-id>.dkr.ecr.<your-region>.amazonaws.com/perftiltbot:latest",
           "memory": 256,
           "cpu": 256,
           "essential": true
         }
       ]
     }
     ```

6. **Run the ECS task:**
   - Use the AWS Management Console or AWS CLI to run the task.
   - Example command:
     ```bash
     aws ecs run-task --cluster <your-cluster> --task-definition perftiltbot
     ```

**Note:** Ensure that your `secrets.yaml` file is securely managed (e.g., using AWS Secrets Manager or environment variables) and not hardcoded in the task definition.

### Required AWS IAM Role Permissions

To publish and run the bot on ECS, your AWS user or role needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:CreateRepository",
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
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecs:CreateCluster",
        "ecs:CreateService",
        "ecs:CreateTaskDefinition",
        "ecs:DescribeClusters",
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:ListClusters",
        "ecs:ListServices",
        "ecs:ListTaskDefinitions",
        "ecs:RunTask",
        "ecs:StartTask",
        "ecs:StopTask",
        "ecs:UpdateService"
      ],
      "Resource": "*"
    }
  ]
}
```

**Note:** This is a minimal set of permissions. Adjust as needed based on your specific requirements.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 