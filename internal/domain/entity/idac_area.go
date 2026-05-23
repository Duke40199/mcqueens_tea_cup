package entity

import (
	"time"

	"github.com/lib/pq"
)

// AreaSyncInfo represents an area that is part of a synchronization group
type AreaSyncInfo struct {
	AreaCode string
	AreaName string
	Timezone string
}

type IDACAreaMetadata struct {
	ID           string
	Name         string
	AreaType     string
	Aliases      pq.StringArray
	SegaAreaCode string
	ALlNetCode   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
