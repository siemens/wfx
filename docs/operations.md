# Operations

## UI

A Web UI is available, currently in an experimental state, allowing you to explore wfx's state within a Web browser.
This UI must be built separately and can optionally be embedded into the wfx binary by using a build tag (see [Installation](installation.md#build-tags)).
When enabled, the UI is accessible through the northbound interface at the `/ui` path, defaulting to [http://localhost:8081/ui/](http://localhost:8081/ui/):

[![wfxui](images/ui-jobs-table.png)](images/ui-jobs-table.png)

**Note**: The UI is intentionally read-only; it is not possible to perform any write operations within wfx.

For releases, wfx is published in two "flavors": one without the UI and one with the UI included.

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

The complete [wfx API specification](../spec/wfx.openapi.yml) is accessible at runtime via the `/api/wfx/v1/openapi.json` endpoint.
Clients may inspect this specification at run-time so to obey the various limits imposed, e.g, for parameter value ranges and array lengths.

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

**Note**: To prevent the connection from being closed due to inactivity (e.g., when no job events occur), periodic keep-alive events are sent during such idle periods.
This ensures that the connection remains open, preventing closure by proxies, the kernel, or other entities since it may not be possible to control all involved parties.
The keep-alive events are technically comments (as defined in the SSE specification) and must be ignored by clients.

#### Event Format Specification

The job events stream is composed of [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) (SSE).
Accordingly, the stream is structured as follows:

```
data: [...]
id: 1

data: [...]
id: 2

[...]

: keepalive

data: [...]
id: 42

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
- Action type of the event

This enables more precise control over the dispatched events.
Note: The filter parameters are independent of each other; an event matches if it satisfies any of the specified filter parameters.

#### Examples

`wfxctl` offers a reference client implementation. The following command subscribes to **all** job events:

```bash
wfxctl job events

# auto reconnect if connection is lost; also waits for wfx to be up and running.
wfxctl job events --auto-reconnect
```

**Note**: The `--auto-reconnect` flag should be used with caution, as it may result in missed events after a connection loss.
When this flag is used, `wfxctl` does not terminate upon losing the connection, so its logs should be monitored to detect such occurrences.
After a connection loss, fetching the job's current status and comparing it with the received events can help identify any missed events.

The above commands monitor events for _all_ jobs globally, which may result in a large number of events.
For a more targeted approach, filter parameters may be used.
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
   domain. This limitation can be addressed by subscribing to all job events through a single SSE connection, or by
   supplying appropriate filter parameters.
5. **HTTP/2**: Not supported currently. The HTTP protocol is limited to HTTP/1.1.

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

## Plugins

wfx offers a flexible (out-of-tree) plugin mechanism for extending its request processing capabilities.
A plugin functions as a subprocess, both initiated and supervised by wfx.
Communication between the plugin and wfx is facilitated through the exchange of [flatbuffer](https://flatbuffers.dev/)
messages over stdin/stdout, thus permitting plugins to be developed in _any_ programming language.

### Design Choices

Due to the potential use of plugins for authentication, it is critical that **all requests are passed through the
plugins** before further processing and that that **no request can slip through** without being processed by the plugins.
This has led to the following deliberate design choices:

1. Should any plugin **exit** (e.g., due to a crash), **wfx is designed to terminate gracefully**. While it might be
   feasible for wfx to attempt restarting the affected plugins, this responsibility is more suitably handled by a
   dedicated process supervisor like systemd. The shutdown of wfx enables the process supervisor to restart wfx, which,
   in turn, starts all its plugins again.
2. All plugins are **initialized before wfx starts processing any requests**. In particular, after the completion of
   wfx's startup phase, it's not possible to add or remove any plugins.
3. Plugins are expected to function properly. Specifically, if a plugin returns an invalid response type or an unexpected
   response (for example, in response to a request that was never sent to the plugin), wfx will terminate gracefully.
   This is because such behavior usually indicates a misconfiguration. The overall strategy is to fail fast and early.

### Using Plugins

To use plugins at runtime, wfx must be compiled with the `plugin` tag (enabled by default) and started with the
`--mgmt-plugins-dir` resp. `--client-plugins-dir` flag, specifying a directory containing the plugins to be used. This
enables the use of different plugin sets for the north- resp. southbound API.

**Note**: In a plugin directory, all _executable_ files (including symlinks to executables) are assumed to be plugins.
Non-executable files, like configuration files, are excluded. For deterministic behavior, plugins are sorted and
executed in lexicographic order based on their filenames during the startup of wfx.

### Developing Plugins

Communication between wfx and a plugin is achieved by exchanging [flatbuffer](https://flatbuffers.dev/) messages via
stdin/stdout. The flatbuffer specification is available in the [fbs](../fbs) directory.
A plugin can use `stderr` for logging purposes (`stderr` is forwarded and prefixed by wfx).

For every incoming request, wfx generates a unique number called `cookie`. The `cookie`, along with the complete request
(e.g., headers and body in the case of HTTP), is written to the plugin's stdin. The plugin then sends its response,
paired with the same `cookie`, back to wfx by writing to its stdout. This `cookie` mechanism ensures that wfx can
accurately associate responses with their corresponding requests.

**Technical Note**: Cookies are represented as unsigned 64-bit integers, which may lead to wraparound. This means there is a
slight possibility that a cookie could be reused for more than one request over the lifespan of a plugin. However, this
event occurs only once every 2^64 requests. By the time such a reuse might happen, the original request associated with
the cookie would have already timed out.

**Note**:

1. It is crucial for the plugin to read data from its stdin descriptor promptly to prevent blocking writes by wfx. The
   `cookie` mechanism facilitates (and encourages!) asynchronous processing.
2. The working directory for the plugin process is the same as the working directory of wfx, which is the directory from which wfx was launched.

Based on the plugin's response, wfx can:

- Modify the incoming request before it undergoes further processing by wfx in the usual manner.
- Send a preemptive response back to the client, such as a "permission denied" or "service unavailable" message.
- Leave the request unchanged.

### Use Cases

Plugins are typically used for:

- Enforcing authentication and authorization for API endpoints.
- Handling URL rewriting and redirection tasks.

### Example

An [example plugin](../example/plugin) written in Go demonstrates denying access to the `/api/wfx/v1/workflows` endpoint.

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
The status codes indicate that a total of 5,625 status updates were sent with HTTP status code 200 and 375 jobs were created with HTTP status code 201.
No errors were reported.

The latency over time distribution is illustrated in the following figure:
[![benchmark plot](images/benchmark.png)](images/benchmark.png)
