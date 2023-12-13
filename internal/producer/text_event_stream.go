package producer

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"io"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/runtime"
)

func TextEventStreamProducer() runtime.Producer {
	return runtime.ProducerFunc(func(rw io.Writer, data any) error {
		if _, err := rw.Write([]byte("data: ")); err != nil {
			return fault.Wrap(err)
		}
		b, err := json.Marshal(data)
		if err != nil {
			return fault.Wrap(err)
		}
		if _, err := rw.Write(b); err != nil {
			return fault.Wrap(err)
		}
		// text/event-stream responses are "chunked" with double newline breaks
		if _, err := rw.Write([]byte("\n\n")); err != nil {
			return fault.Wrap(err)
		}
		return nil
	})
}
