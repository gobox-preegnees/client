package cache

type GetFileDTO struct {
	FileName string
	ModTime  int64
}

type PutFileDTO struct {
	FileName string
	ModTime  int64
	HashSum  string
	SizeFile int64
}
