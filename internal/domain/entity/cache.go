package entity

type FileId struct {
	fileName string
	modTime  int64
}

type Cahce struct {
	files map[FileId]File
}