# Operations

## API

wfx provides two RESTful APIs to interact with it: the northbound operator/management interface and the southbound interface used by clients as illustrated in the following figure:

```txt
     Management

         │
         │
         ▼ Northbound API
┌──────────────────┐
│       wfx        │
└──────────────────┘
         ▲ Southbound API
         │
         │

      Device
```

The northbound API is used to create jobs and execute server-side state transitions, whereas the southbound
API is used for client-side transitions.

The complete [wfx API specification](../spec/wfx.swagger.yml) is accessible at runtime via the `/swagger.json` endpoint.
Clients may inspect this specification at run-time so to obey the various limits imposed, e.g, for parameter value ranges and array lengths.

For convenience, wfx includes a built-in Swagger UI accessible at runtime via <http://localhost:8080/api/wfx/v1/docs>, assuming default listening host and port [configuration](configuration.md).

### Job Status Notifications

Job status notifications inform clients of changes in job status instantly, avoiding the need for repeated server polling.
This can be useful for UIs and applications requiring prompt status updates.

The status notifications are built using [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) (SSE):

```txt
 ┌────────┐                            ┌─────┐
 │ Client │                            │ wfx │
 └────────┘                            └─────┘
     |                                    |
     |   GET /jobs/{id}/status/subscribe  |
     |-----------------------------------►|
     |                                    |
     |                                    |
     |            loop                    |
 ┌───|────────────────────────────────────|───┐
 │   |       [text/event-stream]          |   │
 │   |                                    |   │
 │   |            send event              |   │
 │   |◄-----------------------------------|   │
 │   |  e.g. data: {"state":"INSTALLING"} |   │
 │   |                                    |   │
 └───|────────────────────────────────────|───┘
     |                                    |
     ▼                                    ▼
 ┌────────┐                            ┌─────┐
 │ Client │                            │ wfx │
 └────────┘                            └─────┘
```

1. The client sends a `GET` request to the endpoint `/jobs/{id}/status/subscribe`.
2. The server sets the `Content-Type` set to `text/event-stream`.
3. The response body contains a stream of job status updates, with **the first event being the current job status**.
   Each event is terminated by a pair of newline characters `\n\n` (as required by the SSE spec).

**Example:** `wfxctl` provides a reference client implementation, e.g.

```bash
# subscribe to a job with ID c28ef750-bdda-4d1c-a3fa-bb0ea1dca9d3
wfxctl job subscribe-status --id c28ef750-bdda-4d1c-a3fa-bb0ea1dca9d3
```

**Client Notes**

1. Subscribing to a non-existent job yields a 404 Not Found error.
2. Subscribing to a job in a terminal state (a state without outgoing transitions) yields a 400 Bad Request error.
3. The connection closes when the job reaches a terminal state (after the last status update has been sent).
4. On issues (e.g., slow client), wfx closes connection. The client must reconnect if needed.

**Caveats**

1. Job status updates are dispatched asynchronously to prevent a subscriber from causing a delay in status updates.
   Consequently, messages may be **delivered out-of-order**. This is typically not an issue; however, under
   circumstances with high concurrency, such as multiple clients attempting to modify the same job or a single client
   sending a burst of status updates, this behavior might occur.
2. Server-Sent Events use unidirectional communication, which means that **message delivery is not guaranteed**. In
   other words, there are scenarios where certain events may never reach the subscriber.
3. When scaling out horizontally, the load balancer should consistently route job updates for a specific job ID to the same wfx instance.
   Alternatively, the client can subscribe to updates from every wfx instance and subsequently aggregate the events it receives.
4. Browsers typically limit SSE connections (6 per domain). HTTP/2 can be used instead (100 connections by default) or some kind of
   aggregation service could be used.

### Response Filters

wfx allows server-side response content filtering prior to sending the response to the client so to tailor it to client information needs.
For example, if a client isn't interested in the job's history, it may request to omit it in wfx responses ― which also saves bandwidth.
To this end, clients send a custom HTTP header, `X-Response-Filter`, with a [`jq`](https://stedolan.github.io/jq/)-like expression value.
For example, assuming a job with ID 1 exists,

```bash
curl -s -f http://localhost:8080/api/wfx/v1/jobs/1/status \
  -H "X-Response-Filter: .state"
```

returns the current `state` of the job as a string.

Note that the (filtered) response might no longer be a valid JSON expression as is the case in this example.
It's the client's responsibility to handle the filtered response properly ― which it asked for being filtered in the first place.

### Health Check

wfx includes an internal health check service that's accessible at `/health`, e.g., via

```
curl http://localhost:8080/health
```

in standard [configuration](configuration.md) or, alternatively,

```
wfxctl health
```

### wfx Version

The version of wfx running is accessible at `/version`, e.g., via

```
curl http://localhost:8080/version
```

in standard [configuration](configuration.md) or, alternatively,

```
wfxctl version
```

## Deployment

To assist a secure deployment, both, the northbound operator/management interface and the southbound client interface, are isolated from each other by being bound to distinct ports so that, e.g., a firewall can be used to steer and restrict access.

While this separation provides a basic level of security, it doesn't prevent clients interfering with each other:
For example, there is no mechanism in place to prevent a client `A` from updating the jobs of another client `B`.
This is a deliberate design choice following the Unix philosophy of "Make each program do one thing well".
It allows for flexible integration of wfx into existing or new infrastructure.
Existing infrastructure most probably has request authentication and authorization measures in place.
For new infrastructure, such a feature is likely better provided by specialized components/services supporting the overall deployment security architecture.

Thus, for productive deployments, a deployment along the lines of the following figure is recommended with an _API
Gateway_ subsuming the discussed security requirements and performing access steering with regard to, e.g., client
access.

```txt
┌──────────────────┐
│    Operator /    │
│    Management    │
└──────────────────┘
         │
         │ Request Authentication & Authorization
         ▼
┌──────────────────┐
│   API Gateway    │
└──────────────────┘
         │
         │ Northbound: Management API
         ▼
┌──────────────────┐
│       wfx        │
└──────────────────┘
         ▲
         │ Southbound: Client API
         │
┌──────────────────┐
│   API Gateway    │
└──────────────────┘
         ▲
         │ Request Authentication & Authorization
         │
 ┌──────────────────┐
┌┴─────────────────┐│
│      Client      ├┘
└──────────────────┘
```

## Telemetry

No telemetry or user data is collected or processed by wfx.

Note that there is an indirect dependency on `go.opentelemetry.io/otel` via [Go OpenAPI](https://github.com/go-openapi) by the **client runtime** (as used by `wfxctl`).
Telemetry is deliberately turned off in wfx.
See [Add support for tracing via OpenTelemetry](https://github.com/go-openapi/runtime/pull/254) for details.

## Performance / Benchmarking

wfx has been designed with performance and horizontal scalability in mind.

To regress-test and gauge the performance of wfx in particular scenarios, the `wfx-loadtest` tool stressing the REST API
can be helpful.
It's build by default alongside the other wfx binaries, see Section [Building wfx](installation.md#building-wfx).

As an example, the following commands execute a benchmark of the default SQLite persistent storage:

```bash
wfx --log-format json --log-level warn &
wfx-loadtest --log-level warn --duration 60s
```

Note: In the above example, the log format is JSON since pretty-printing is an expensive operation.

The benchmark result including statistics is printed to the terminal after 60 seconds and also available in the
`results` directory for further inspection.

```text
*******************************************************************************
 Summary
*******************************************************************************
Requests      [total, rate, throughput]         6000, 100.02, 100.02
Duration      [total, attack, wait]             59.987s, 59.986s, 792.516µs
Latencies     [min, mean, 50, 90, 95, 99, max]  187.363µs, 612.69µs, 602.423µs, 744.281µs, 809.622µs, 1.062ms, 7.242ms
Bytes In      [total, mean]                     4041243, 673.54
Bytes Out     [total, mean]                     81693, 13.62
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:5625  201:375
Error Set:
```

and is to be interpreted as follows:
There were 6,000 total requests at a rate of 100.02 requests per second.
In terms of latency, the minimum response time was 187.363 microseconds, the mean was 612.69 microseconds, and the 99th percentile was 1.062 milliseconds.
All requests were successful with a success ratio of 100%.
The status codes indicate that a total of 5,625 status updates were sent with HTTP error code 200 and 375 jobs were created with HTTP error code 201.
No errors were reported.

The latency over time distribution is illustrated in the following figure:
[![benchmark plot](images/benchmark.png)](images/benchmark.png)
