package workflow

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func DeleteWorkflow(ctx context.Context, storage persistence.Storage, name string) error {
	log := logging.LoggerFromCtx(ctx)
	log.Debug().Str("name", name).Msg("Deleting workflow")
	return fault.Wrap(storage.DeleteWorkflow(ctx, name))
}
