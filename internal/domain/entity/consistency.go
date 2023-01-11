package entity

type NeedToRename struct {
	OldFile File `json:"old_file"`
	NewFile File `json:"new_File"`
}

type NeedToRemove struct {
	File File `json:"file"`
}

type NeedToUpload struct {
	File  File   `json:"file"`
	Token string `json:"token"`
}

type NeedToDownload struct {
	File  File   `json:"file"`
	Token string `json:"token"`
}

type Consistency struct {
	RequestId int    `json:"request_id" validate:"required"`
	Client    string `json:"client" validate:"required"`

	NeedToRemove   []NeedToRemove
	NeedToRename   []NeedToRename
	NeedToUpload   []NeedToUpload
	NeedToDownload []NeedToDownload
}
