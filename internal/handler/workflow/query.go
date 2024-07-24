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
	"strings"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func QueryWorkflows(ctx context.Context, storage persistence.Storage, paginationParams persistence.PaginationParams, sort *string) (*api.PaginatedWorkflowList, error) {
	log := logging.LoggerFromCtx(ctx)
	var sortParams persistence.SortParams
	if sort != nil {
		sortParams.Desc = strings.ToLower(*sort) == "desc"
	}
	log.Debug().Bool("desc", sortParams.Desc).Msg("Querying workflows")
	list, err := storage.QueryWorkflows(ctx, sortParams, paginationParams)
	return list, fault.Wrap(err)
}
