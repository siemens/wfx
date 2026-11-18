// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0

package gron

import (
	"fmt"
	"io"
	"sort"

	"github.com/Southclaws/fault"
)

// Encode writes gron assignments for JSON read from r.
func Encode(w io.Writer, r io.Reader) error {
	ss, err := statementsFromJSON(r, statement{{"json", typBare}})
	if err != nil {
		return err
	}
	sort.Sort(ss)
	for _, s := range ss {
		if _, err := fmt.Fprintln(w, s.String()); err != nil {
			return fault.Wrap(err)
		}
	}
	return nil
}
