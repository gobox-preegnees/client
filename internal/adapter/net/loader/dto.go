package loader

import (
	"context"
	"os"
)

type DownloadReqDTO struct {
	Ctx      context.Context
	FileName string
	SizeFile int64
	ModFile  int64
	HashSum  string
	F        *os.File
	DToken   string
}

type UploadDTO struct {
}
