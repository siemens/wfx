# Config Deployment Demo

This repository showcases a proof-of-concept (PoC) for deploying configurations to a device (e.g. for diagnostic purposes), using `wfx` and a custom workflow.

⚠️ **Warning: Not Suitable for Production!** This PoC was developed without a focus on security. For instance, it does not employ TLS.

## Features

- Deploy a configuration file
- After file has been deployed, an arbitrary shell action can be peformed, such as restarting a service

## Getting Started

Follow these steps to set up and run the demo:

1. **Build**: Simply execute `make`.
2. **Start wfx**: `wfx --simple-fileserver ./files`
3. **Create the workflow**: If it's your first run, create the workflow with `wfxctl workflow create wfx.workflow.config.deployment.yml`.
4. **Deploy the client**: Transfer the `config-deployer` binary to the target device and run it with the desired client id: `./config-deployer -c foo`.
5. **Initialize the Job**: Generate a new job for the mentioned client id with `./create-job.sh -c foo`.

## Licensing

This project is distributed under the [Apache-2.0](../../LICENSE) license.
