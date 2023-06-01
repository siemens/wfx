package job

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

func DeleteJob(ctx context.Context, storage persistence.Storage, jobID string) error {
	log := logging.LoggerFromCtx(ctx)
	log.Info().Str("id", jobID).Msg("Deleting job")
	return fault.Wrap(storage.DeleteJob(ctx, jobID))
}
