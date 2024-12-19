# Workflows

A _workflow_ is a finite-state machine that is defined in [YAML](https://github.com/go-yaml/yaml#compatibility) format and follows specific rules and constraints.
More precisely, a _workflow_ consists of

- a non-empty unique `name`,
- a non-empty finite set of named `states`,
- a non-empty finite set of `transitions` correlating states by their names, and
- optionally, a set of `groups` collating states by their names,

see the [wfx OpenAPI Specification](../spec/wfx.openapiv3.yml)'s `Workflow` object for reference.

wfx supports dynamic loading and unloading of workflows at run-time; when a workflow is loaded into wfx, it is validated
to ensure that it adheres to these rules and constraints.
Upon loading a workflow into wfx it is persistently stored (see [Configuration](configuration.md)) and is henceforth **immutable**.
However, as long as there is no job referencing it ― including finished ones ― workflows can be unloaded, i.e., deleted
from wfx's persistent storage.
After having corrected the unloaded workflow, it can be loaded into wfx again.

Graph-wise, the set of states defines the nodes of the finite-state machine graph and the set of transitions defines the directed edges among the nodes.
In the optional set of groups, disjoint sets of states can be combined for semantic grouping and easier state query selection.

Each _state_ in `states` consists of

- a unique `name`, and
- a `description`.

Each _group_ in `groups` consists of

- a unique `name` and
- a non-empty set of `states`' names.

Each _transition_ in `transitions` consists of

- a starting state name `from` matching one of the unique state names in `states`,
- an ending state name `to` matching one of the unique state names in `states`,
- an `eligible` attribute denoting the entity that may execute the transition, either `CLIENT` or `WFX`, and
- an optional `action` attribute that ― depending on the `eligible` entity ― specifies the transition execution action.

**Note**: Trivial transitions, where the source and destination states are the same (`from == to`), are implicit in the workflow.
These transitions allow the client to report progress within the same state without requiring the transition to be
explicitly defined in the workflow (thereby clobbering the workflow).

The currently valid actions for `WFX`-eligible transitions are

- `IMMEDIATE`: wfx instantly transitions to the transition's ending state `to`, and
- `WAIT`: external north-bound input, e.g., by an operator or a higher-up hierarchical wfx is required to advance the workflow

with `WAIT` being the default `WFX` action.

For `CLIENT`-eligible transitions, there are currently no pre-defined actions to assign to in a workflow specification.
Instead, the transition execution actions are encoded in the client implementations which are specific to a workflow (family).

Extending the available actions for `WFX`-eligible transitions and providing actions for `CLIENT`-eligible transitions is recorded in the [Roadmap](https://github.com/siemens/wfx/issues).

Beyond syntactic requirements such as, e.g., a particular set being non-empty or the uniqueness property of names, the
following semantic rules and constraints are checked and enforced by wfx upon loading a workflow:

- There's exactly one initial state, i.e., there are no incoming transitions to this state with this state's name in `to`.
- There are no unreachable states, i.e., states without an incoming transition.
- For each state, there can't be more than one outgoing transition whose `action` is `IMMEDIATE`.
- Transition tuples (`from`, `to`, `eligible`, `action`) must be unique.
- There are no cycles in the workflow graph _except_ for trivial cycles, i.e. transitions where `from` equals `to` (used for e.g. progress reporting).
- Each state belongs to _at most one_ group.

**Tip**: `wfxctl` can be used to validate workflows offline, e.g.

```bash
wfxctl workflow validate workflow/dau/wfx.workflow.dau.direct.yml
```

These definitions are illustrated in more detail in the following exemplary Kanban workflow.

## Jobs

A _job_ is an instance of a workflow in which wfx and a client progress in lock-step. As a result, a job can only be
created if the corresponding workflow already exists. Jobs can be driven by different workflows. This allows wfx to
accommodate a variety of use cases and workflows that are tailored to specific needs.
For example, firmware and container image updates may have different requirements and constraints that require separate
workflows to address them.

### Creating Jobs

A new job can be created by sending a `JobRequest` object (see [wfx OpenAPI Specification](../spec/wfx.openapiv3.yml)) to wfx's [northbound REST API](operations.md#API).
A JobRequest consists of the following:

- a non-empty `clientId` to assign the job to a specific client
- a non-empty `workflow` name to select a workflow for the job
- an optional array of `tags` that can be used to query the job
- an optional job `definition`, which is a freeform JSON object that provides job-specific data to the client.

For example, the job `definition` could contain a download URL for an encrypted firmware artifact that only the specific
client can decrypt.

When wfx accepts the `JobRequest`, a new `Job` entity is created, and the following fields are populated by wfx:

- `id`: the globally unique job id
- `state`: the unique initial state of the workflow that drives the job
- `stime`, `mtime`: the date and time (ISO8601) when the job was created
- `status.definitionHash`: a hash value computed over the `definition` field, used to detect job `definition` modifications.

### Updating Jobs

After a job has been created, its `definition`, `status`, and `tags` can be updated using wfx's REST APIs. If a job
`definition` is updated _after_ a client has already started working on the job, this assumes that the client can handle
the changes. To simplify things for client authors, wfx automatically updates the `status.definitionHash` whenever the
`definition` changes. This provides a mechanism for detecting changes in the `definition`.

### Job History

Whenever a job's `status` or `definition` changes, wfx prepends the current value to the job's `history` array.
This allows for reviewing job updates later on and diagnosing problems.

**Note**: By default, the job history is omitted from all `Job` responses, primarily due to its diagnostic nature but
also for bandwidth reasons. Clients requiring a job's historical data must explicitly request the specific job and
include the `history=true` parameter in their request.

### Deleting Jobs

Jobs can be deleted using the northbound REST API. For example, this can be used to perform maintenance on old jobs.
Note that wfx does not perform any housekeeping on its own.

## Kanban Example Workflow

An exemplary [Kanban](https://en.wikipedia.org/wiki/Kanban)-inspired workflow YAML specification may be defined as follows:

```yaml
name: wfx.workflow.kanban

groups:
  - name: OPEN
    description: The task is ready for the client(s).
    states:
      - NEW
      - PROGRESS
      - VALIDATE

  - name: CLOSED
    description: The task is in a final state, i.e. it cannot progress any further.
    states:
      - DONE
      - DISCARDED

states:
  - name: BACKLOG
    description: Task is created

  - name: NEW
    description: task is ready to be pulled

  - name: PROGRESS
    description: task is being worked on

  - name: VALIDATE
    description: task is validated

  - name: DONE
    description: task is done according to the definition of done

  - name: DISCARDED
    description: task is discarded

transitions:
  - from: BACKLOG
    to: NEW
    eligible: WFX
    action: IMMEDIATE
    description: |
      Immediately transition to "NEW" upon a task hitting the backlog,
      conveniently done by wfx "on behalf of" the Product Owner.

  - from: NEW
    to: PROGRESS
    eligible: CLIENT
    description: |
      A Developer pulls the task or
      the Product Owner discards it (see below transition),
      whoever comes first.

  - from: NEW
    to: DISCARDED
    eligible: WFX
    description: |
      The Product Owner discards the task or
      a Developer pulls it (see preceding transition),
      whoever comes first.

  - from: PROGRESS
    to: VALIDATE
    eligible: CLIENT
    description: |
      The Developer has completed the task, it's ready for validation.

  - from: PROGRESS
    to: PROGRESS
    eligible: CLIENT
    description: |
      The Developer reports task completion progress percentage.

  - from: VALIDATE
    to: DISCARDED
    eligible: WFX
    description: |
      The task result has no customer value.

  - from: VALIDATE
    to: DISCARDED
    eligible: CLIENT
    description: |
      The task result cannot be integrated into Production software.

  - from: VALIDATE
    to: DONE
    eligible: CLIENT
    description: |
      A Developer has validated the task result as useful.

  - from: VALIDATE
    to: DONE
    eligible: WFX
    action: WAIT
    description: |
      The Product Owner has validated the task result as useful
```

resulting in the following graph representation:

```
 BACKLOG
    │
    ▼
   NEW ───────┐
    │         │
    ├◀─┐      │
    ▼  │      │
PROGRESS ─────┤
    │         │
    ▼         │
VALIDATE ─────┤
    │         │
    ▼         ▼
  DONE    DISCARDED
```

### Hands-on: Playing Kanban

Assuming the preceding Kanban workflow is saved in the file `wfx.workflow.kanban.yml`,

```sh
wfxctl workflow create --filter=.transitions wfx.workflow.kanban.yml
```

loads the Kanban workflow into wfx making it available to create jobs driven by it.
Note that the the command's output was made less verbose by using a `--filter` to only show the transitions.

Henceforth, a job is identified with a "Task" and a state is identified with a "Lane" in Kanban board parlance.
The Product Owner Parker as also owning and managing the Kanban "Board" instruments the northbound wfx management
interface while Developers instrument the southbound client interface.

Now that the Kanban workflow is available, Product Owner Parker creates a new Backlog item ― overriding the Pull
principle ― to be taken by the highly specialized Developer Dana:

```sh
echo '{ "title": "expose job api" }' | \
    wfxctl job create --workflow wfx.workflow.kanban \
                      --client-id dana \
                      --filter='del(.workflow)' -
```

The Task is immediately transitioned to the "NEW" state (Kanban Lane) by wfx on behalf of the Product Owner Parker as
the transition's `action` is `IMMEDIATE`.
Note that the piped-in JSON document is the _Job Definition_ (see [wfx OpenAPI Specification](../spec/wfx.openapiv3.yml)) which is the contract
between the operator (Product Owner) creating jobs (Tasks) and the clients (Developers) executing those jobs so that
they're actually able to process the job.

Then, Developer Dana, knowing the job (Task) Identifier, pulls the task into "PROGRESS"

```sh
wfxctl job update-status \
    --actor=client \
    --id=1 \
    --state=PROGRESS
```

and starts working on it.

Traditional tools like `curl` can also be used instead of `wfxctl` to achieve the same result:

```sh
curl -X PUT \
  http://localhost:8080/api/wfx/v1/jobs/1/status \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{"state":"PROGRESS"}'
```

Meanwhile, Developer Dana sporadically reports progress

```sh
wfxctl job update-status \
    --actor=client \
    --id=1 \
    --state=PROGRESS \
    --progress $((RANDOM % 100))
```

until having finished the task and progressing it into the "VALIDATE" state:

```sh
wfxctl job update-status \
    --actor=client \
    --id=1 \
    --state=VALIDATE
```

Then, Developer Dana may realize that the task result cannot be integrated into the Production software and may put it
into "DISCARDED" or Product Owner Parker may come to the conclusion that no customer value is provided by the task
result, also putting it into "DISCARDED" ― whoever realizes this first.

Or, if the task result greatly increases customer value, Developer Dana as being highly experienced may do the
validation herself, thereafter putting the task to "DONE". The Product Owner Parker may come to the same conclusion ―
again whoever is faster in realizing the customer benefit.

Alternatively, if the task in question holds significant potential for increased customer value, Developer Dana, being highly
experienced, may undertake the validation process herself and subsequently mark the task as "DONE."
Likewise, Product Owner Parker may arrive at the same conclusion and act accordingly - again, whoever realizes this
first may advance the task accordingly.
