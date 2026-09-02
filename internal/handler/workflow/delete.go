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
	"github.com/Southclaws/fault/fmsg"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func DeleteWorkflow(ctx context.Context, storage persistence.Storage, name string) error {
	log := logging.LoggerFromCtx(ctx)
	if err := storage.DeleteWorkflow(ctx, name); err != nil {
		return fault.Wrap(err, fmsg.Withf("failed to delete workflow %q", name))
	}
	log.Info().Str("name", name).Msgf("Deleted workflow %q", name)
	return nil
}
