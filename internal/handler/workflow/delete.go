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
	if err := storage.DeleteWorkflow(ctx, name); err != nil {
		log.Err(err).Str("name", name).Msgf("Failed to delete workflow %q", name)
		return fault.Wrap(err)
	}
	log.Info().Str("name", name).Msgf("Deleted workflow %q", name)
	return nil
}
