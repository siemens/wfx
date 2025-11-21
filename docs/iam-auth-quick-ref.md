# Quick Reference: PostgreSQL IAM Authentication

## Enable IAM Auth

```bash
export WFX_POSTGRES_IAM_AUTH=true
```

## Run wfx

```bash
wfx --storage postgres \
    --storage-opt "host=mydb.us-east-1.rds.amazonaws.com port=5432 user=myuser database=wfx sslmode=require"
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `WFX_POSTGRES_IAM_AUTH` | Enable IAM auth | `true` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `PGHOST` | Database host | `mydb.us-east-1.rds.amazonaws.com` |
| `PGPORT` | Database port | `5432` |
| `PGUSER` | Database user | `myiamuser` |
| `PGDATABASE` | Database name | `wfx` |
| `PGSSLMODE` | SSL mode | `require` |

## Kubernetes IRSA

```yaml
serviceAccountName: wfx-service-account
env:
  - name: WFX_POSTGRES_IAM_AUTH
    value: "true"
```

## Database Setup

```sql
CREATE USER myiamuser WITH LOGIN;
GRANT rds_iam TO myiamuser;
GRANT ALL PRIVILEGES ON DATABASE wfx TO myiamuser;
```

## IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["rds-db:connect"],
    "Resource": ["arn:aws:rds-db:REGION:ACCOUNT:dbuser:DBI_ID/USER"]
  }]
}
```

## Test Connection

```bash
# Generate token (test IAM permissions)
aws rds generate-db-auth-token \
  --hostname mydb.us-east-1.rds.amazonaws.com \
  --port 5432 \
  --username myiamuser \
  --region us-east-1

# Test connection
export WFX_POSTGRES_IAM_AUTH=true
wfx --storage postgres --log-level debug
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Access Denied | Check IAM policy has `rds-db:connect` |
| Auth Failed | Verify user has `rds_iam` role |
| Connection Timeout | Check security groups and network |
| Token Error | Verify AWS credentials are configured |
| SSL Error | Use `sslmode=require` |

## Documentation

- Full guide: `docs/postgres-iam-auth.md`
- IAM policies: `docs/iam-policy-examples.md`
- Configuration: `docs/configuration.md`
