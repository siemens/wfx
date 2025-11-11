# PostgreSQL IAM Authentication for AWS RDS

This document describes how to enable and use AWS RDS IAM authentication with wfx.

## Overview

wfx supports AWS RDS IAM authentication for PostgreSQL connections, allowing you to use IAM roles (such as IRSA in Kubernetes) instead of traditional database passwords. This provides:

- **Enhanced Security**: No need to store database passwords in configuration files or environment variables
- **Automatic Token Rotation**: IAM authentication tokens are regenerated automatically every 15 minutes
- **IRSA Support**: Works seamlessly with IAM Roles for Service Accounts in Kubernetes/EKS
- **Instance Profile Support**: Works with EC2 instance profiles
- **CNAME Resolution**: Automatically resolves CNAMEs to actual RDS endpoints

## Prerequisites

1. **AWS RDS PostgreSQL Instance** with IAM authentication enabled
2. **IAM Policy** allowing `rds-db:connect` action
3. **Database User** created with IAM authentication enabled
4. **AWS Credentials** configured (IRSA, instance profile, or environment variables)

## IAM Policy

Create an IAM policy that allows connecting to your RDS instance:

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
        "arn:aws:rds-db:REGION:ACCOUNT_ID:dbuser:DBI_RESOURCE_ID/DATABASE_USER"
      ]
    }
  ]
}
```

Replace:
- `REGION`: Your AWS region (e.g., `us-east-1`)
- `ACCOUNT_ID`: Your AWS account ID
- `DBI_RESOURCE_ID`: Your RDS instance resource ID (found in RDS console)
- `DATABASE_USER`: Your database username

## Database Setup

1. Enable IAM authentication on your RDS instance
2. Create a database user with IAM authentication:

```sql
CREATE USER myuser WITH LOGIN;
GRANT rds_iam TO myuser;
GRANT ALL PRIVILEGES ON DATABASE wfx TO myuser;
```

## Configuration

### Required Dependencies

Add the following dependencies to your `go.mod`:

```bash
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/feature/rds/auth
```

### Enable IAM Authentication

Set the environment variable to enable IAM authentication:

```bash
export WFX_POSTGRES_IAM_AUTH=true
```

Or in Kubernetes deployment:

```yaml
env:
  - name: WFX_POSTGRES_IAM_AUTH
    value: "true"
```

### Connection String

When IAM authentication is enabled, the password in the connection string is ignored. Configure wfx with:

```bash
wfx --storage postgres \
    --storage-opt "host=mydb.abc123.us-east-1.rds.amazonaws.com port=5432 user=myuser database=wfx sslmode=require"
```

Or using environment variables:

```bash
export PGHOST=mydb.abc123.us-east-1.rds.amazonaws.com
export PGPORT=5432
export PGUSER=myuser
export PGDATABASE=wfx
export PGSSLMODE=require
export WFX_POSTGRES_IAM_AUTH=true

wfx --storage postgres
```

## IRSA (IAM Roles for Service Accounts) in Kubernetes

### Step 1: Create IAM Role for Service Account

```bash
eksctl create iamserviceaccount \
  --name wfx-service-account \
  --namespace default \
  --cluster my-cluster \
  --attach-policy-arn arn:aws:iam::ACCOUNT_ID:policy/wfx-rds-connect \
  --approve
```

### Step 2: Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wfx
spec:
  template:
    spec:
      serviceAccountName: wfx-service-account
      containers:
      - name: wfx
        image: wfx:latest
        env:
        - name: WFX_POSTGRES_IAM_AUTH
          value: "true"
        - name: PGHOST
          value: "mydb.abc123.us-east-1.rds.amazonaws.com"
        - name: PGPORT
          value: "5432"
        - name: PGUSER
          value: "myuser"
        - name: PGDATABASE
          value: "wfx"
        - name: PGSSLMODE
          value: "require"
        - name: AWS_REGION
          value: "us-east-1"
        command:
        - /usr/bin/wfx
        - --storage
        - postgres
```

## EC2 Instance Profile

For EC2 instances, attach the IAM policy to the instance profile:

```bash
# The application will automatically use the instance profile
export WFX_POSTGRES_IAM_AUTH=true
wfx --storage postgres --storage-opt "host=mydb.abc123.us-east-1.rds.amazonaws.com port=5432 user=myuser database=wfx sslmode=require"
```

## Region Detection

The AWS region is automatically detected from:
1. RDS endpoint format (e.g., `*.us-east-1.rds.amazonaws.com`)
2. `AWS_REGION` environment variable
3. `AWS_DEFAULT_REGION` environment variable
4. Falls back to `us-east-1` if not detected

## CNAME Support

If you use a CNAME record pointing to your RDS endpoint, wfx will automatically resolve it to the actual RDS endpoint. This is useful for DNS-based failover or aliasing.

Example:
```bash
# CNAME: db.example.com -> mydb.abc123.us-east-1.rds.amazonaws.com
export PGHOST=db.example.com
export WFX_POSTGRES_IAM_AUTH=true
wfx --storage postgres
```

## How It Works

1. When IAM authentication is enabled, wfx creates a custom database connector
2. For each new connection, the connector:
   - Resolves any CNAME to the actual RDS endpoint
   - Extracts the AWS region from the endpoint
   - Generates a fresh IAM authentication token (valid for 15 minutes)
   - Establishes the connection using the token
3. The database connection pool automatically creates new connections with fresh tokens as needed

## Troubleshooting

### Connection Timeout

If you see connection timeouts, check:
- Network connectivity to RDS endpoint
- Security group rules allow connections from your workload
- SSL/TLS is configured correctly (`sslmode=require`)

### Authentication Failed

If authentication fails:
- Verify IAM policy allows `rds-db:connect` for the specific database user
- Check that the database user exists and has `rds_iam` role
- Ensure AWS credentials are properly configured
- Verify the region is correctly detected

### CNAME Issues

IAM authentication tokens are bound to the actual RDS endpoint, not CNAMEs. wfx automatically resolves CNAMEs, but if resolution fails:
- Check DNS configuration
- Manually specify the actual RDS endpoint instead of CNAME
- Ensure network allows DNS queries

### Proxy Configuration

In enterprise environments with proxies, you may need:

```bash
export NO_PROXY=169.254.169.254  # For instance metadata
export HTTPS_PROXY=http://proxy.example.com:8080
export HTTP_PROXY=http://proxy.example.com:8080
```

## Disabling IAM Authentication

To revert to password authentication, simply unset the environment variable:

```bash
unset WFX_POSTGRES_IAM_AUTH
```

Or set it to `false`:

```bash
export WFX_POSTGRES_IAM_AUTH=false
```

Then provide the password in the connection string:

```bash
wfx --storage postgres \
    --storage-opt "host=localhost port=5432 user=wfx password=secret database=wfx"
```

## Security Considerations

1. **Always use SSL/TLS** (`sslmode=require` or `sslmode=verify-full`)
2. **Principle of Least Privilege**: Grant only necessary database permissions to IAM-authenticated users
3. **Audit Logging**: Enable RDS CloudWatch logs to track database access
4. **Network Isolation**: Use VPC security groups to restrict database access
5. **Token Expiry**: IAM tokens expire after 15 minutes, but active connections remain valid

## References

- [AWS RDS IAM Database Authentication](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html)
- [IAM Roles for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [Reference Implementation](https://github.com/califlower/golang-aws-rds-iam-postgres)
