package resp

type CodeWithMsgInteract struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Target uint64 `json:"target"`
	Action string `json:"action"`
}
