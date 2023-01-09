package entity

type NeedToRename struct {
	OldFile File `json:"old_file"`
	NewFile File `json:"new_File"`
}

type NeedToRemove struct {
	File `json:"file"`
}

type NeedToLock struct {
	File `json:"file"`
}

type NeedToUpload struct {
	File `json:"file"`
	// This field stores a jwt token,
	// with which the client will later contact the server to send or download a file.
	// Fields: folder, username, client, file_name, size_file, mod_time, hash_sum, action:upload|download|download
	Token string `json:"token"`
}

type NeedToDownload struct {
	File `json:"file"`
	// Fields: Virtual_file_name, folder, username, client, file_name, size_file, mod_time, hash_sum, action:upload|download|download
	Token string `json:"token"`
}

type Consistency struct {
	// эти данные придут в ответе от сервера их нужно провалидировать, в контроллере
	// // Installed on the server from field of client
	// ExternalId int `json:"external_id"`
	// // Installed on the server
	// InternalId int `json:"internal_id"`
	// // Installed on the server
	// Timestamp int `json:"timestamp" validate:"required"`

	// Update bool `json:"update" validate:"required"`

	// Эти данные возможно нужно удалить
	// Username string `json:"username" validate:"required"`
	// Folder   string `json:"folder" validate:"required"`
	// Client   string `json:"client" validate:"required"`

	NeedToRemove   []NeedToRemove
	NeedToLock     []NeedToLock
	NeedToRename   []NeedToRename
	NeedToUpload   []NeedToUpload
	NeedToDownload []NeedToDownload
}
