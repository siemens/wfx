#!/bin/bash
# Example: Running wfx with AWS RDS IAM Authentication

# Prerequisites:
# 1. RDS PostgreSQL instance with IAM auth enabled
# 2. Database user created with rds_iam role
# 3. AWS credentials configured (IRSA, instance profile, or env vars)
# 4. IAM policy allowing rds-db:connect

# Example RDS endpoint
RDS_ENDPOINT="mydb.abc123xyz.us-east-1.rds.amazonaws.com"
DB_USER="myiamuser"
DB_NAME="wfx"
DB_PORT="5432"

echo "Starting wfx with AWS RDS IAM Authentication..."
echo "RDS Endpoint: $RDS_ENDPOINT"
echo "Database User: $DB_USER"
echo "Database Name: $DB_NAME"
echo ""

# Enable IAM authentication
export WFX_POSTGRES_IAM_AUTH=true

# Optional: Set AWS region if not auto-detected
export AWS_REGION=us-east-1

# Start wfx
./wfx --storage postgres \
    --storage-opt "host=$RDS_ENDPOINT port=$DB_PORT user=$DB_USER database=$DB_NAME sslmode=require" \
    --log-level debug

# Alternative: Using environment variables
# export PGHOST=$RDS_ENDPOINT
# export PGPORT=$DB_PORT
# export PGUSER=$DB_USER
# export PGDATABASE=$DB_NAME
# export PGSSLMODE=require
# export WFX_POSTGRES_IAM_AUTH=true
# ./wfx --storage postgres --log-level debug
