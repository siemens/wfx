# Remote Access Demo

This repository showcases a proof-of-concept (PoC) for establishing a remote terminal connection to a device (e.g. for diagnostic purposes), using `wfx` and a custom workflow.

⚠️ **Warning: Not Suitable for Production!** This PoC was developed without a focus on security. For instance, it does not employ TLS.

## Features

- Remote terminal connectivity with a configurable time-to-live (TTL)
- Basic Auth

## Prerequisites

For the client side, make sure you have the following tools installed:

- [ttyd](https://github.com/tsl0922/ttyd): a simple command-line tool for sharing terminals over the web

## Getting Started

Follow these steps to set up and run the demo:

1. **Build**: Simply execute `make`.
2. **Start wfx**: `wfx`
3. **Create the workflow**: If it's your first run, create the workflow with `wfxctl workflow create wfx.workflow.remote.access.yml`.
4. **Deploy the client**: Transfer the `remote-access-client` binary to the target device and run it with the desired client id: `./remote-access-client -c foo`.
5. **Initialize the Job**: Generate a new job for the mentioned client id with `./create-job.sh -c foo`. This job prompts the client to activate a remote terminal, accessible through a web browser.
6. **Access the terminal:** Navigate to <http://localhost:1337> in your browser. Log-in using the credentials `admin:secret`.

## Additional Notes

If the client isn't directly accessible externally (e.g. due to network restrictions), [socat](http://www.dest-unreach.org/socat/) can be leveraged to use the wfx host as a relay:

```bash
# execute this on the wfx host, assuming 10.0.5.42 is the client's IP address
socat TCP4-LISTEN:1337,fork,reuseaddr TCP4:10.0.5.42:1337
```

Following this, you can use `http://<wfx address>:1337` to access the client.
It is left as an **exercise to the reader** to implement a client on the wfx host which auto-establishes a `socat` tunnel once the client reaches the `OPENED` status.

## Licensing

This project is distributed under the [Apache-2.0](../../LICENSE) license.
