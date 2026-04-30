# AWS IAM Roles Setup

## Overview
IAM roles define permissions for different components of the DropsAndGrinds infrastructure.

## Required IAM Roles

### 1. EC2 Instance Role
Attached to EC2 instances running the application.

**Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds-db:connect"
      ],
      "Resource": [
        "arn:aws:rds-db:region:account:dbuser:db-id/dbadmin"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "elasticache:Connect"
      ],
      "Resource": [
        "arn:aws:elasticache:region:account:cluster:droppandgrinds-redis"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::dropsandgrinds-static/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:region:account:secret:droppandgrinds/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:aws:logs:region:account:log-group:/droppandgrinds/*"
      ]
    }
  ]
}
```

### 2. ECS Task Execution Role
For ECS/Fargate deployments.

**Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:aws:logs:region:account:log-group:/ecs/droppandgrinds:*"
      ]
    }
  ]
}
```

### 3. ECS Task Role
For the application running in ECS.

**Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds-db:connect"
      ],
      "Resource": [
        "arn:aws:rds-db:region:account:dbuser:db-id/dbadmin"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "elasticache:Connect"
      ],
      "Resource": [
        "arn:aws:elasticache:region:account:cluster:droppandgrinds-redis"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:region:account:secret:droppandgrinds/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::droppandgrinds-static/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage"
      ],
      "Resource": [
        "arn:aws:sqs:region:account:droppandgrinds-*"
      ]
    }
  ]
}
```

### 4. Lambda Execution Role
For Lambda functions (e.g., price scraping, email alerts).

**Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds-db:connect"
      ],
      "Resource": [
        "arn:aws:rds-db:region:account:dbuser:db-id/dbadmin"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "elasticache:Connect"
      ],
      "Resource": [
        "arn:aws:elasticache:region:account:cluster:droppandgrinds-redis"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": [
        "arn:aws:secretsmanager:region:account:secret:droppandgrinds/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:aws:logs:region:account:log-group:/aws/lambda/droppandgrinds-*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ses:SendEmail",
        "ses:SendRawEmail"
      ],
      "Resource": "*"
    }
  ]
}
```

### 5. CodeBuild Service Role
For CI/CD pipeline.

**Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
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
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "arn:aws:logs:region:account:log-group:/aws/codebuild/droppandgrinds*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:GetObjectVersion"
      ],
      "Resource": [
        "arn:aws:s3:::droppandgrinds-codebuild-*"
      ]
    }
  ]
}
```

## Security Best Practices

1. **Principle of Least Privilege**: Only grant permissions that are explicitly needed
2. **Use IAM Roles Instead of Keys**: Never embed AWS credentials in code
3. **Rotate Credentials Regularly**: If access keys are used, rotate them every 90 days
4. **Enable MFA**: For human users accessing AWS console
5. **Use Condition Keys**: Restrict access based on time, IP, or other conditions
6. **Monitor with CloudTrail**: Log all IAM API calls
7. **Regular Audits**: Review permissions and remove unused roles

## Secrets Management

Store sensitive values in AWS Secrets Manager:
- Database credentials
- Redis AUTH token
- JWT secret
- API keys (Steam, CheapShark, etc.)
- SMTP credentials for email

**Example Secret:**
```bash
aws secretsmanager create-secret \
  --name dropsandgrinds/database \
  --secret-string '{"username":"dbadmin","password":"secure-password"}'
```

Access from application:
```go
import "github.com/aws/aws-sdk-go-v2/service/secretsmanager"

// Retrieve secret
result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
  SecretId: aws.String("dropsandgrinds/database"),
})
```

## Troubleshooting

### Access Denied Errors
- Verify role is attached to the resource
- Check policy allows the specific action
- Review resource ARN format
- Check for explicit deny statements

### Role Not Found
- Verify role exists in correct region
- Check role name spelling
- Ensure role is created before attaching

### Temporary Credentials Expired
- EC2 instance role credentials refresh automatically
- For external applications, use STS AssumeRole
- Implement credential refresh logic
