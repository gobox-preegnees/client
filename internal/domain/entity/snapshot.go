package entity

type Snapshot struct {
	Mode         int    `json:"mode"`
	RequesID     int    `json:"reques_id"`
	SnapshotHash string `json:"snapshot_hash"`
	Client       string `json:"client"`
	Files        []File `json:"files"`
}
