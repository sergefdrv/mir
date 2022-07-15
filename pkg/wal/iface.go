// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package wal

import (
	"context"

	"github.com/filecoin-project/mir/pkg/events"
)

// WAL provides access to the write-ahead log.
//
// WALLoadAll method loads all events stored in the WAL and returns
// them as a new EventList.
type WAL interface {
	WALLoadAll(ctx context.Context) (*events.EventList, error)
}
