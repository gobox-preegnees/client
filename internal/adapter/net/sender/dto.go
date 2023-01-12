package sender

import "context"

type SendSnapshotDTO struct {
	Ctx      context.Context
	Snapshot []byte
}
