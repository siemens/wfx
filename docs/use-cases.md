# Use Cases

This document presents a collection of use cases for employing wfx.
Owing to wfx's versatility, the list provided here is not exhaustive.
Should you identify an important use case missing, please feel free to [contribute](../CONTRIBUTING.md).

**Note**: Each use case requires an appropriate client implementation to execute the specific "business logic". If no such client exists yet, this implies writing custom code.

## Software Update

Perform over-the-air (OTA) firmware updates using [SWUpdate](https://swupdate.org/) and the [Device Artifact Update (DAU) Workflow Family](../workflow/dau/README.md).
See also SWUpdate's [Suricatta documentation](https://github.com/sbabic/swupdate/blob/master/doc/source/suricatta.rst#support-for-wfx).

## Remote Access

Establish remote terminal (debug) sessions to devices, e.g. for diagnostic purposes. The process involves:

1. Creating a custom workflow [wfx.workflow.remote.access](../contrib/remote-access/wfx.workflow.remote.access.yml) that encapsulates the steps to initiate a remote terminal session.
2. Generating a new job for the device using this workflow. The job metadata could contain authentication credentials.
3. The client checks wfx periodically for new jobs. Upon finding the job from step 2, it opens its firewall and starts a (secure) service to accept remote terminal connections.
4. Following a pre-set timeout (e.g. configurable in the job metadata), the client will close its firewall and terminate the terminal service.

A proof-of-concept demonstrating remote terminal connections via WebSockets (browser-based) is available [here](../contrib/remote-access/README.md).

## Config Deployment

Another common use case involves configuration deployment, akin to the [Software Update](#software-update) use case.
The goal is to roll out a new configuration to one or multiple client(s), leading to the restart of certain services.
Again this requires defining an [appropriate workflow](../contrib/config-deployment/wfx.workflow.config.deployment.yml) and a corresponding [client](../contrib/config-deployment/client/worker.go).
A proof-of-concept is available [here](../contrib/config-deployment/README.md).
