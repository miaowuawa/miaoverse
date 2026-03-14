package resp

type CodeWithMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
