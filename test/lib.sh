# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'
load 'test_helper/bats-file/load'

# shellcheck disable=SC2034
BASEURL="http://localhost:8080/api/wfx/v1"

wait_wfx_running() {
    local expected=${1:-2}
    local count=0
    for _i in {1..20}; do
        count=$(wfxctl health | grep -c up || echo 0)
            if [[ "$count" -ge "$expected" ]]; then
                break
            fi
        sleep 0.1
    done
    echo "$count"
}

launch_wfx() {
    wfx --storage sqlite --storage-opt "file:wfx?mode=memory&cache=shared&_fk=1" &
}

create_kanban_workflow() {
    local fname="$BATS_FILE_TMPDIR/kanban.yml"
    cat <<EOF >>"$fname"
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
EOF
    wfxctl workflow create "$fname" >/dev/null
}

create_kanban_job() {
    local id
    id=$(wfxctl job create --workflow wfx.workflow.kanban \
        --client-id Dana --filter '.id' --raw)
    echo "$id"
}

# vim: ft=bash
