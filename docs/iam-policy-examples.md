# IAM Policy Examples for RDS IAM Authentication

## Basic RDS Connect Policy

This policy allows connecting to a specific database user on a specific RDS instance.

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
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-ABCDEFGHIJKLMNOP/myiamuser"
      ]
    }
  ]
}
```

**Resource ARN Format:**
```
arn:aws:rds-db:<region>:<account-id>:dbuser:<dbi-resource-id>/<db-username>
```

Where:
- `region`: AWS region (e.g., `us-east-1`)
- `account-id`: Your AWS account ID
- `dbi-resource-id`: RDS instance resource ID (found in RDS console, starts with `db-`)
- `db-username`: Database username that will connect

## Finding Your DBI Resource ID

### AWS Console
1. Go to RDS Console
2. Select your database instance
3. Look for "Resource ID" in the Configuration tab
4. It will look like: `db-ABCDEFGHIJKLMNOP`

### AWS CLI
```bash
aws rds describe-db-instances \
  --db-instance-identifier mydb \
  --query 'DBInstances[0].DbiResourceId' \
  --output text
```

## Multiple Users Policy

Allow connecting as multiple database users:

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
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-ABCDEFGHIJKLMNOP/user1",
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-ABCDEFGHIJKLMNOP/user2",
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-ABCDEFGHIJKLMNOP/readonly"
      ]
    }
  ]
}
```

## Multiple RDS Instances Policy

Allow connecting to multiple RDS instances:

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
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-INSTANCE1RESID/myuser",
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-INSTANCE2RESID/myuser",
        "arn:aws:rds-db:us-west-2:123456789012:dbuser:db-INSTANCE3RESID/myuser"
      ]
    }
  ]
}
```

## Wildcard Policy (All Users on Instance)

⚠️ **Use with caution** - This allows connecting as any user on the instance:

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
        "arn:aws:rds-db:us-east-1:123456789012:dbuser:db-ABCDEFGHIJKLMNOP/*"
      ]
    }
  ]
}
```

## Trust Policy for IAM Role (IRSA)

When creating an IAM role for IRSA, you need a trust policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": "system:serviceaccount:default:wfx-service-account",
          "oidc.eks.us-east-1.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:aud": "sts.amazonaws.com"
        }
      }
    }
  ]
}
```

Replace:
- `123456789012`: Your AWS account ID
- `EXAMPLED539D4633E53DE1B71EXAMPLE`: Your EKS cluster's OIDC provider ID
- `default`: Kubernetes namespace
- `wfx-service-account`: Service account name

## Creating IAM Role with eksctl

```bash
# Create service account with IAM role
eksctl create iamserviceaccount \
  --name wfx-service-account \
  --namespace default \
  --cluster my-eks-cluster \
  --region us-east-1 \
  --attach-policy-arn arn:aws:iam::123456789012:policy/wfx-rds-connect \
  --approve \
  --override-existing-serviceaccounts
```

## Creating IAM Role Manually

1. **Create the policy:**
```bash
aws iam create-policy \
  --policy-name wfx-rds-connect \
  --policy-document file://rds-connect-policy.json
```

2. **Create the role:**
```bash
aws iam create-role \
  --role-name wfx-rds-iam-role \
  --assume-role-policy-document file://trust-policy.json
```

3. **Attach the policy to the role:**
```bash
aws iam attach-role-policy \
  --role-name wfx-rds-iam-role \
  --policy-arn arn:aws:iam::123456789012:policy/wfx-rds-connect
```

## EC2 Instance Profile

For EC2 instances, attach the RDS connect policy to an instance profile:

```bash
# Create role
aws iam create-role \
  --role-name wfx-ec2-rds-role \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"Service": "ec2.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }]
  }'

# Attach RDS connect policy
aws iam attach-role-policy \
  --role-name wfx-ec2-rds-role \
  --policy-arn arn:aws:iam::123456789012:policy/wfx-rds-connect

# Create instance profile
aws iam create-instance-profile \
  --instance-profile-name wfx-instance-profile

# Add role to instance profile
aws iam add-role-to-instance-profile \
  --instance-profile-name wfx-instance-profile \
  --role-name wfx-ec2-rds-role

# Attach to EC2 instance
aws ec2 associate-iam-instance-profile \
  --instance-id i-1234567890abcdef0 \
  --iam-instance-profile Name=wfx-instance-profile
```

## Testing IAM Policy

Test if your IAM credentials can generate an auth token:

```bash
# Using AWS CLI
aws rds generate-db-auth-token \
  --hostname mydb.abc123.us-east-1.rds.amazonaws.com \
  --port 5432 \
  --username myiamuser \
  --region us-east-1
```

If this succeeds, it will output a token. If it fails, check your IAM policy.

## Common Issues

### Issue: Access Denied when generating token
**Solution:** Check that:
- IAM policy includes `rds-db:connect` action
- Resource ARN matches your RDS instance and user
- Policy is attached to the role/user
- DBI Resource ID is correct (not instance identifier)

### Issue: Authentication failed after generating token
**Solution:** Check that:
- Database user exists and has `rds_iam` role
- RDS instance has IAM authentication enabled
- SSL is enabled (`sslmode=require`)
- Security group allows your connections

### Issue: Policy works in one region but not another
**Solution:** Ensure:
- Resource ARN includes correct region
- You're generating token for the correct region
- AWS_REGION environment variable is set correctly

## Security Best Practices

1. **Principle of Least Privilege**: Grant access only to specific users and instances
2. **Use Conditions**: Add conditions to restrict access by IP, time, or other factors
3. **Monitor Access**: Enable CloudTrail and RDS logging to audit database access
4. **Rotate Roles**: Periodically review and update IAM policies
5. **Separate Environments**: Use different IAM roles for dev, staging, and production

## Additional Resources

- [AWS RDS IAM Database Authentication](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html)
- [IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [RDS IAM Authentication Best Practices](https://aws.amazon.com/blogs/database/using-iam-authentication-to-connect-with-sql-workbenchj-to-amazon-aurora-postgresql-or-amazon-rds-for-postgresql/)
