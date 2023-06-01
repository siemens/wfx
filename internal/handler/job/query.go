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
	"strings"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func QueryJobs(ctx context.Context, storage persistence.Storage, filterParams persistence.FilterParams, paginationParams persistence.PaginationParams, sort *string) (*model.PaginatedJobList, error) {
	log := logging.LoggerFromCtx(ctx)

	var sortParams persistence.SortParams
	if sort != nil {
		sortParams = parseSortParam(*sort)
	}

	jobs, err := storage.QueryJobs(ctx, filterParams, sortParams, paginationParams)
	if err != nil {
		log.Err(err).Msg("Failed to query jobs")
		return nil, fault.Wrap(err)
	}
	return jobs, nil
}

func parseSortParam(param string) persistence.SortParams {
	return persistence.SortParams{Desc: strings.ToLower(param) == "desc"}
}
