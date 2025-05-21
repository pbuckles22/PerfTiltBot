# PerfTiltBot

A performance-focused bot for analyzing and providing insights about gameplay tilt.

## Prerequisites

- Go 1.23.1 or higher
- Git

### Windows Users

This project requires [Scoop](https://scoop.sh/) to be installed for managing dependencies like `yq` (used for YAML merging in the management scripts).

**To install Scoop:**
```powershell
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex
```

**To install yq using Scoop:**
```powershell
scoop install yq
```

After installing these, you can use the management scripts as described below.

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

## Bot Authentication and Configuration

### Configuration Structure

The bot uses a cascading configuration system that separates bot-specific authentication from channel-specific settings:

1. **Bot Authentication File**
   - Create a file named `configs/<bot_name>_auth_secrets.yaml` for your bot's authentication
   - Example (`configs/perftiltbot_auth_secrets.yaml`):
   ```yaml
   bot_name: "perftiltbot"
   oauth: "oauth:your_bot_oauth_token"
   client_id: "your_bot_client_id"
   client_secret: "your_bot_client_secret"
   # Optionally, API keys for services
   apis:
     openai:
       api_key: "your_openai_api_key"
     twitch:
       client_id: "your_twitch_client_id"
       client_secret: "your_twitch_client_secret"
   ```

2. **Channel Configuration File**
   - Create a file named `configs/<channel_name>_config_secrets.yaml` for each channel
   - Example (`configs/pbuckles_config_secrets.yaml`):
   ```yaml
   bot_name: "perftiltbot"  # Links to the bot's authentication file
   channel: "pbuckles"      # The Twitch channel name
   data_path: "/app/data/pbuckles"
   commands:
     queue:
       max_size: 100
       default_position: 1
       default_pop_count: 1
     cooldowns:
       default: 5
       moderator: 2
       vip: 3
   ```

### Configuration Flow

1. The system first reads the bot authentication file
2. Then merges it with the channel configuration file
3. Channel-specific settings can override bot settings if needed

### Managing Configurations

Use the management scripts to handle configurations:

```bash
# List all channels using a specific bot
.\run_bot.ps1 list-channels perftiltbot

# Update bot configuration
.\run_bot.ps1 update-bot perftiltbot

# Start a bot for a channel
.\run_bot.ps1 start pbuckles
```

The script will automatically:
- Validate the configuration
- Merge bot and channel settings
- Mount the correct configuration files when starting the bot

### Security Notes

- Keep your bot's OAuth token and client credentials secure
- Never commit `*_auth_secrets.yaml` or `*_config_secrets.yaml` files to version control
- Use different bot configurations for different environments (development, production)
- Regularly rotate OAuth tokens and client credentials

## License

This project is licensed under the MIT License - see the LICENSE file for details. 