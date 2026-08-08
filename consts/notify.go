package consts

// 通知域常量：类型、状态。

const (
	NotifyTypeAccountSecurity uint8 = 0 // 账号安全
	NotifyTypeTransaction     uint8 = 1 // 事务事项
	NotifyTypeLike            uint8 = 2 // 互动-赞
	NotifyTypeFollow          uint8 = 3 // 互动-关注
	NotifyTypeMention         uint8 = 4 // 互动-@我
	NotifyTypeReply           uint8 = 5 // 互动-回复与评论
)

const (
	NotifyStatusUnread  uint8 = 0 // 未读
	NotifyStatusRead    uint8 = 1 // 已读
	NotifyStatusDeleted uint8 = 2 // 已删除
)
