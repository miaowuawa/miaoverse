package resp

type CodeWithMsgBlock struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Target uint32 `json:"target"`
	Type   uint8  `json:"type"`
	Action string `json:"action"`
}
