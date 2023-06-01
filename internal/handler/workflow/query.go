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
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func QueryWorkflows(ctx context.Context, storage persistence.Storage, paginationParams persistence.PaginationParams) (*model.PaginatedWorkflowList, error) {
	log := logging.LoggerFromCtx(ctx)
	log.Debug().Msg("Querying workflows")
	list, err := storage.QueryWorkflows(ctx, paginationParams)
	return list, fault.Wrap(err)
}
