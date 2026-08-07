package blockreq

type UpdateBlock struct {
	Target uint32 `json:"target"`
	Type   uint8  `json:"type"`
	Action string `json:"action"`
}
