#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

p "# use wfxctl to interact with wfx\n"

pe "wfxctl health"
wait

p "# query existing jobs:"
pe "wfxctl jobs query"
wait

p "# let's create a new job"
p "wfxctl job create \n\
\t--client-id=my_client \n\
\t--workflow=wfx.workflow.dau.direct \n\
\t--artifact=file:///firmware.swu \n\
\t--update-type=firmware"
wait
wfxctl job create \
    --client-id my_client \
    --workflow wfx.workflow.dau.direct

p "# query by id and apply a jq-filter:"
pe "wfxctl job get --id 1 --filter '.currentState'"
wait

p "# log request and response:"
pe "DEBUG=true wfxctl job get --id 1 > /dev/null"
wait

p "# same request using curl:"
pe 'curl -H "Accept: application/json" "http://localhost:8081/api/wfx/v1/jobs/1" | jq'
wait

p "# query all jobs by group:"
pe "wfxctl jobs query --group OPEN"
wait

p "# thanks for watching the demo. bye!"
