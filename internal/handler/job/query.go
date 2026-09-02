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
	"github.com/Southclaws/fault/fmsg"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
)

func QueryJobs(
	ctx context.Context,
	storage persistence.Storage,
	filterParams persistence.FilterParams,
	paginationParams persistence.PaginationParams,
	sort *string,
) (*api.PaginatedJobList, error) {
	var sortParams persistence.SortParams
	if sort != nil {
		sortParams = parseSortParam(*sort)
	}

	jobs, err := storage.QueryJobs(ctx, filterParams, sortParams, paginationParams)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("failed to query jobs"))
	}
	return jobs, nil
}

func parseSortParam(param string) persistence.SortParams {
	return persistence.SortParams{Desc: strings.ToLower(param) == "desc"}
}
