package entity

type File struct {
	FileName string `json:"file_name" validate:"required"`
	HashSum  string `json:"hash_sum"`
	SizeFile int64  `json:"size_file"`
	ModTime  int64  `json:"mod_time" validate:"required"`
}
