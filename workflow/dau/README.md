# Device Artifact Update (DAU)

The Device Artifact Update Workflow Family has been specifically designed for updating software on devices.

## Workflows

Currently, the DAU Workflow Family comprises two workflows that model the software update process for devices:

The [`wfx.workflow.dau.direct`](wfx.workflow.dau.direct.yml) workflow caters for the fully automated software update use case while the
[`wfx.workflow.dau.phased`](wfx.workflow.dau.phased.yml) workflow operates in distinct phases requiring external input to advance the workflow.

Both workflows operate in **stages**, each of which follows the Command, Feedback, and Completion (CFC) scheme.
These stages consist of:

1. a _command_ state which instructs the device to start the update process,
2. an _action + feedback_ state during which the device performs the necessary work and provides progress updates to wfx, and
3. a _completion_ state in which the device signals the completion of the stage.

A sequence of such CFC loops constitutes the particular workflow.
In case of `wfx.workflow.dau.direct`, the wfx transitions to the next stage automatically while in case of `wfx.workflow.dau.phased`,
wfx waits for external input, e.g., from an operator.

### wfx.workflow.dau.direct

The [`wfx.workflow.dau.direct`](wfx.workflow.dau.direct.yml) workflow consists of the two stages _installation_ and _activation_:
During the installation stage, the device downloads and installs the update artifacts.
In the subsequent activation stage, the device takes action to activate the update.
Depending on the type of artifact(s), this activation action varies as, e.g.,
container images need a different activation action than firmware disk images
(see below Section [Job Definition](#job-definition)).

The graph representation of the `wfx.workflow.dau.direct` workflow is depicted in the following figure, omitting state descriptions and transition eligibles for legibility:

```txt
 INSTALL ────────┐
    │            │
    ├─◀─┐        │
    ▼   │        │
INSTALLING ──────┤
    │            │
    ▼            │
INSTALLED        │
    │            │
    ▼            │
 ACTIVATE ───────┤
    │            │
    ├─◀─┐        │
    ▼   │        │
ACTIVATING ──────┤
    │            │
    ▼            ▼
ACTIVATED    TERMINATED
```

### wfx.workflow.dau.phased

The [`wfx.workflow.dau.phased`](wfx.workflow.dau.phased.yml) workflow is similar to the `wfx.workflow.dau.direct` workflow but starts in the `CREATED` state instead and
introduces another _download_ stage in between installation and activation to decouple artifact download from its installation:
The initial state `CREATED` serves as an anchor to actually kickstart the external input-driven CFC scheme.
With the additional download stage, a maintenance window could be realized prior to which the artifact is downloaded but only installed when
wfx commands the begin of the installation phase (e.g. a certain time window has been reached).

Consequently ― and in contrast to the `wfx.workflow.dau.direct` ― wfx doesn't transition to the next stage automatically
but waits for external input to do so, e.g., by an operator.

The graph representation of the `wfx.workflow.dau.phased` workflow is depicted in the following figure, again omitting
state descriptions and transition eligibles for legibility:

```txt
  CREATED
     │
     ▼
  DOWNLOAD ───────┐
     │            │
     ├─◀─┐        │
     ▼   │        │
DOWNLOADING ──────┤
     │            │
     ▼            │
 DOWNLOADED       │
     │            │
     ▼            │
  INSTALL ────────┤
     │            │
     ├─◀─┐        │
     ▼   │        │
 INSTALLING ──────┤
     │            │
     ▼            │
 INSTALLED        │
     │            │
     ▼            │
  ACTIVATE ───────┤
     │            │
     ├─◀─┐        │
     ▼   │        │
 ACTIVATING ──────┤
     │            │
     ▼            ▼
 ACTIVATED    TERMINATED
```

## Job Definition

As being general purpose, wfx doesn't impose a particular schema on the information conveyed to the device describing its action(s) to perform, except that it's in JSON format.
Instead, the job definition is a contract between the operator creating jobs, each possibly following a different workflow, and the client(s) executing those jobs in lock-step with the wfx.
The same is true for the type of update artifacts that are specified in the job definition and that can be of any form such as, e.g., firmware disk images, container images, or configurations:
The operator has to has to take care to only assign jobs to devices that are known to be able to digest this type of update artifact.
wfx doesn't exercise any checks on compatibility.

An exemplary job definition using the [built-in simple file server](../../docs/configuration.md#file-server) may look like the following JSON document:

```json
{
  "version": "1.0",
  "type": ["firmware", "dummy"],
  "artifacts": [
    {
      "name": "Example Device Firmware Artifact",
      "version": "1.1",
      "uri": "http://wfx.host:8080/download/example_artifact.swu"
    }
  ]
}
```

The `type` list field allows to label update jobs. Labels may be used by wfx's on-device counterpart to determine the activation action(s) to execute since,
e.g., container images need a different activation action than firmware disk images or a configuration change.

In the preceding example, the presence of the `firmware` label may be used to instruct the on-device client to test-reboot into the new firmware.

Since wfx isn't concerned with the job definition except for conveying it to the device, it can be adapted to specific
needs by feeding in a different job definition into the wfx on job creation and having a client on the device that can
digest it.
