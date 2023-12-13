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

### Job Events

Job events provide a notification mechanism that informs clients about certain operations happening on jobs.
This approach eliminates the need for clients to continuously poll wfx, thus optimizing network usage and client resources.
Job events can be useful for user interfaces (UIs) and other applications that demand near-instantaneous updates.

#### Architecture

Below is a high-level overview of how the communication flow operates:

```txt
 ┌────────┐                            ┌─────┐
 │ Client │                            │ wfx │
 └────────┘                            └─────┘
     |                                    |
     |        HTTP GET /jobs/events       |
     |-----------------------------------►|
     |                                    |
     |                                    |
     |           Event Loop               |
 ┌───|────────────────────────────────────|───┐
 │   |  [Content-Type: text/event-stream] |   │
 │   |                                    |   │
 │   |                                    |   │
 │   |            Push Event              |   │
 │   |◄-----------------------------------|   │
 │   |                                    |   │
 └───|────────────────────────────────────|───┘
     |                                    |
     ▼                                    ▼
 ┌────────┐                            ┌─────┐
 │ Client │                            │ wfx │
 └────────┘                            └─────┘
```

1. The client initiates communication by sending an HTTP `GET` request to the `/jobs/events` endpoint. Clients may also
   include optional [filter parameters](#filter-parameters) within the request.
2. Upon receipt of the request, `wfx` sets the `Content-Type` header to `text/event-stream`.
3. The server then initiates a stream of job events in the response body, allowing clients to receive instant updates.

#### Event Format Specification

The job events stream is composed of [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) (SSE).
Accordingly, the stream is structured as follows:

```
data: [...]
id: 1

data: [...]
id: 2

[...]
```

An individual event within the stream conforms to this format:

```
data: { "action": "<ACTION>", "ctime": <CTIME>, "tags": <TAGS>, "job": <JOB> }
id: <EVENT_ID>\n\n
```

**Note**: Each event is terminated by a pair of newline characters `\n\n` (as required by the SSE spec).

The semantics of the individual fields is:

- `<ACTION>` specifies the type of event that occurred. The valid actions are:
  - `CREATE`: a new job has been created
  - `DELETE`: an existing job has been deleted
  - `ADD_TAGS`: tags were added to a job
  - `DELETE_TAGS`: tags were removed from a job
  - `UPDATE_STATUS`: job status has been updated
  - `UPDATE_DEFINITION`: job definition has been updated
- `<CTIME>`: event creation time (ISO8601)
- `<TAGS>`: JSON array of tags as provided by the client
- `<JOB>` is a JSON object containing the portion of the job object which was changed, e.g., for an `UPDATE_STATUS` event, the job status is sent but not its definition. To enable [filtering](#filter-parameters), the fields `id`, `clientId` and `workflow.name` are _always_ part of the response.
- `<EVENT_ID>`: an integer which uniquely identifies each event, starting at 1 and incrementing by 1 for every subsequent event. Clients can use this to identify any missed messages. If an overflow occurs, the integer resets to zero, a scenario the client can recognize and address.

**Example:**

```
data: {"action":"UPDATE_STATUS","job":{"clientId":"Dana","id":"c6698105-6386-4940-a311-de1b57e3faeb","status":{"definitionHash":"adc1cfc1577119ba2a0852133340088390c1103bdf82d8102970d3e6c53ec10b","state":"PROGRESS"},"workflow":{"name":"wfx.workflow.kanban"}}}
id: 1\n\n
```

#### Filter Parameters

Job events can be filtered using any combination of the following parameters:

- Job IDs
- Client IDs
- Workflow Names

This enables more precise control over the dispatched events. Note that it is entirely possible to subscribe multiple
times to job events using various filters in order to create a more advanced event recognition model.

#### Examples

`wfxctl` offers a reference client implementation. The following command subscribes to **all** job events:

```bash
wfxctl job events
```

This may result in a large number of events, though. For a more targeted approach, filter parameters may be used.
Assuming the job IDs are known (either because the jobs have been created already or the IDs are received via another
subscription channel), the following will subscribe to events matching either of the two specified job IDs:

```bash
wfxctl job events --job-id=d305e539-1d41-4c95-b19a-2a7055c469d0 --job-id=e692ad92-45e6-4164-b3fd-8c6aa884011c
```

See `wfxctl job events --help` for other filter parameters, e.g. workflow names.

#### Considerations and Limitations

1. **Asynchronous Job Status Updates**: Job status updates are dispatched asynchronously to avoid the risk of a
   subscriber interfering with the actual job operation. In particular, there is no guarantee that the messages sent to
   the client arrive in a linear order (another reason for that may be networking-related). While this is typically not
   a concern, it could become an issue in high-concurrency situations. For example, when multiple clients try to modify
   the same job or when a single client issues a rapid sequence of status updates. As a result, messages could arrive in
   a not necessarily linear order, possibly deviating from the client's expectation. However, the client can use the
   (event) `id` and `ctime` fields to establish a natural ordering of events as emitted by wfx.
2. **Unacknowledged Server-Sent Events (SSE)**: SSE operates on a one-way communication model and does not include an
   acknowledgment or handshake protocol to confirm message delivery. This design choice aligns with the fundamental
   principles of SSE but does mean that there's a possibility some events may not reach the intended subscriber (which
   the client can possibly detect by keeping track of SSE event IDs).
3. **Event Stream Orchestration**: Each wfx instance only yields the events happening on that instance. Consequently, if
   there are multiple wfx instances, a consolidated "global" event stream can only be assembled by subscribing to all
   wfx instances (and aggregating the events).
4. **Browser Connection Limits for SSE**: Web browsers typically restrict the number of SSE connections to six per
   domain. To overcome this limitation, HTTP/2 can be used, allowing up to 100 connections by default, or [filter
   parameters](#filter-parameters) can be utilized to efficiently manage the connections.

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
