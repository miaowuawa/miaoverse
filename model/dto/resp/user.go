package resp

type CodeWithMsgUserID struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	UID  uint64 `json:"uid"`
}
