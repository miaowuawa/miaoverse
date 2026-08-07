package resp

import "miaoverse/model/dao/user"

type RelationUser struct {
	user.User
	BlockStatus uint8 `json:"block_status"`
}

type CodeWithMsgRelationList struct {
	Code  int            `json:"code"`
	Msg   string         `json:"msg"`
	Count int64          `json:"count"`
	Users []RelationUser `json:"users"`
}
