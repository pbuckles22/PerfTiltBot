# CloudWatch Logging Setup for PerfTiltBot

This guide will help you set up CloudWatch logging for your PerfTiltBot instances on EC2.

## Prerequisites

- AWS CLI installed on EC2 instance
- AWS credentials configured
- IAM permissions for CloudWatch Logs

## Step 1: Create IAM Policy

1. Go to AWS IAM Console
2. Create a new policy using the `cloudwatch-logs-policy.json` file
3. Name it: `PerfTiltBot-CloudWatch-Logs-Policy`

## Step 2: Create IAM Role

1. Go to AWS IAM Console
2. Create a new role
3. Select "EC2" as the trusted entity
4. Attach the `PerfTiltBot-CloudWatch-Logs-Policy`
5. Name it: `PerfTiltBot-EC2-Role`

## Step 3: Attach Role to EC2 Instance

1. Go to EC2 Console
2. Select your instance
3. Actions → Security → Modify IAM role
4. Select the `PerfTiltBot-EC2-Role`
5. Save changes

## Step 4: Create CloudWatch Log Group

SSH into your EC2 instance and run:

```bash
aws logs create-log-group --log-group-name "/ec2/perftiltbot" --region us-west-2
```

## Step 5: Test CloudWatch Logging

1. Restart your bots with the enhanced script:
   ```bash
   ./deploy_ec2_enhanced.sh restart-all
   ```

2. View CloudWatch logs:
   ```bash
   ./deploy_ec2_enhanced.sh cloudwatch-logs PerfectTilt
   ```

## Alternative: Quick Setup with AWS Credentials

If you prefer not to use IAM roles, you can configure AWS credentials directly:

```bash
aws configure
```

Enter your AWS Access Key ID, Secret Access Key, and region (us-west-2).

## Cost Estimate

- **CloudWatch Logs**: ~$2-5/month for your bot usage
- **Total additional cost**: ~$2-5/month

## Benefits

- Centralized log viewing in AWS Console
- Log retention and search capabilities
- No need to SSH into EC2 for logs
- Log aggregation across all bots
