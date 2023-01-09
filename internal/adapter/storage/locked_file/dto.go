package dto

import (
	"context"
	"os"

	entity "github.com/gobox-preegnees/gobox-client/internal/domain/entity"
)

type CreateOneLockedFileDTO struct {
	entity.File
	Ctx context.Context
	F *os.File
}

type CreateOneUnlockedFileDTO struct {
	entity.File
	Ctx context.Context
}