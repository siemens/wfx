# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
FROM gcr.io/distroless/static-debian12:nonroot@sha256:386df8cadd9fecafe6a7a0bf11b4d47f0edf8d1de39551f9ebd25c4b87f5b01f

# this file is generated by goreleaser as part of the CI pipeline
COPY wfx /usr/bin/wfx

EXPOSE 8080 8081

ENTRYPOINT ["wfx"]
