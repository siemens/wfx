# Implementation Summary: PostgreSQL IAM Authentication for AWS RDS

## What Has Been Done

I've modified the wfx codebase to enable PostgreSQL connection with AWS RDS IAM authentication using IRSA (IAM Roles for Service Accounts). The implementation is based on the reference: https://github.com/califlower/golang-aws-rds-iam-postgres

### Files Modified

1. **`internal/persistence/entgo/postgres.go`** - Main implementation
   - Added AWS SDK v2 imports for IAM authentication
   - Implemented `iamDB` struct as a custom `driver.Connector` 
   - Added `Connect()` method that generates fresh IAM tokens for each connection
   - Added `resolveRDSEndpoint()` to handle CNAME resolution and region extraction
   - Added `checkIAMAuthEnabled()` to check if IAM auth should be enabled
   - Modified `Initialize()` to support both password and IAM authentication modes

2. **`docs/configuration.md`** - User documentation
   - Added section about AWS RDS IAM Authentication
   - Includes example usage and link to detailed guide

### Files Created

1. **`docs/postgres-iam-auth.md`** - Comprehensive IAM auth guide
   - Prerequisites and IAM policy setup
   - Database configuration steps
   - IRSA setup for Kubernetes/EKS
   - EC2 instance profile usage
   - Troubleshooting guide
   - Security considerations

2. **`docs/iam-policy-examples.md`** - IAM policy templates and examples
   - Multiple policy formats and use cases
   - Trust policies for IRSA
   - Testing and validation commands

3. **`docs/iam-auth-quick-ref.md`** - Quick reference card
   - Environment variables
   - Common commands
   - Troubleshooting table

4. **`scripts/add-aws-deps.sh`** - Shell script for Linux/macOS
   - Automates adding AWS SDK dependencies

5. **`scripts/add-aws-deps.ps1`** - PowerShell script for Windows
   - Automates adding AWS SDK dependencies

6. **`examples/run-with-iam-auth.sh`** - Example standalone script
   - Shows how to run wfx with IAM auth

7. **`examples/kubernetes-iam-auth.yaml`** - Kubernetes deployment example
   - Complete K8s deployment with IRSA configuration

## What You Need to Do Next

### Step 1: Install AWS SDK Dependencies (REQUIRED)

Run one of these commands to add AWS SDK v2 dependencies:

**Option A: Using the provided script (Windows PowerShell)**
```powershell
cd d:\wfx\wfx
.\scripts\add-aws-deps.ps1
```

**Option B: Using the provided script (Linux/macOS)**
```bash
cd /path/to/wfx
./scripts/add-aws-deps.sh
```

**Option C: Manually**
```bash
go get github.com/aws/aws-sdk-go-v2/aws@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/feature/rds/auth@latest
go mod tidy
```

### Step 2: Build the Application

```bash
go build -o wfx.exe ./cmd/wfx
```

### Step 3: Test the Implementation

#### Test backward compatibility (password auth):
```bash
wfx --storage postgres \
    --storage-opt "host=localhost port=5432 user=wfx password=secret database=wfx"
```

#### Test with IAM auth enabled:
```bash
export WFX_POSTGRES_IAM_AUTH=true
# or Windows: $env:WFX_POSTGRES_IAM_AUTH = "true"

wfx --storage postgres \
    --storage-opt "host=mydb.us-east-1.rds.amazonaws.com port=5432 user=myuser database=wfx sslmode=require"
```

## Documentation

All documentation is located in the `docs/` folder:

- **[docs/postgres-iam-auth.md](docs/postgres-iam-auth.md)** - Complete setup guide with step-by-step instructions
- **[docs/iam-policy-examples.md](docs/iam-policy-examples.md)** - IAM policy templates and examples
- **[docs/iam-auth-quick-ref.md](docs/iam-auth-quick-ref.md)** - Quick reference card
- **[docs/configuration.md](docs/configuration.md)** - Updated main configuration docs

## Examples

Example configurations are in the `examples/` folder:

- **[examples/run-with-iam-auth.sh](examples/run-with-iam-auth.sh)** - Standalone usage script
- **[examples/kubernetes-iam-auth.yaml](examples/kubernetes-iam-auth.yaml)** - Kubernetes deployment with IRSA

## Key Features Implemented

✅ **IAM Authentication Support**: Use IAM roles instead of passwords  
✅ **IRSA Compatible**: Works with Kubernetes IAM Roles for Service Accounts  
✅ **Automatic Token Rotation**: Fresh tokens generated for each connection (15-min validity)  
✅ **CNAME Resolution**: Automatically resolves CNAMEs to actual RDS endpoints  
✅ **Region Auto-Detection**: Extracts AWS region from RDS endpoint or environment  
✅ **Backward Compatible**: Existing password authentication still works  
✅ **Opt-in**: Enabled via `WFX_POSTGRES_IAM_AUTH` environment variable  

## Quick Start

```bash
# 1. Add dependencies
./scripts/add-aws-deps.sh  # or .ps1 for Windows

# 2. Build
go build ./cmd/wfx

# 3. Enable IAM auth
export WFX_POSTGRES_IAM_AUTH=true

# 4. Run
./wfx --storage postgres \
  --storage-opt "host=mydb.us-east-1.rds.amazonaws.com port=5432 user=myuser database=wfx sslmode=require"
```

## Important Notes

- **SSL Required**: Always use `sslmode=require` or `sslmode=verify-full` with IAM auth
- **No Password Needed**: When IAM auth is enabled, password in DSN is ignored
- **Token Expiry**: IAM tokens expire after 15 minutes, but existing connections remain active
- **AWS Credentials**: Must be configured via IRSA, instance profile, or environment variables

## Next Steps

1. Install dependencies using the helper script
2. Build the application
3. Review the documentation in `docs/postgres-iam-auth.md`
4. Test with your RDS instance
5. Deploy using examples in `examples/kubernetes-iam-auth.yaml` if using Kubernetes

For detailed setup instructions, troubleshooting, and AWS configuration, see **[docs/postgres-iam-auth.md](docs/postgres-iam-auth.md)**.
