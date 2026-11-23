package model

import "time"

type ChatThread struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	UserLow           int64      `gorm:"column:user_low;uniqueIndex:uk_thread_users" json:"userLow"`
	UserHigh          int64      `gorm:"column:user_high;uniqueIndex:uk_thread_users" json:"userHigh"`
	LastMsgTime       *time.Time `gorm:"column:last_msg_time" json:"lastMsgTime"`
	UserLowReadMsgID  *int64     `gorm:"column:user_low_read_msg_id" json:"userLowReadMsgId"`
	UserHighReadMsgID *int64     `gorm:"column:user_high_read_msg_id" json:"userHighReadMsgId"`
}

func (ChatThread) TableName() string { return "tb_chat_thread" }

type ChatMessage struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	ThreadID   int64     `gorm:"column:thread_id" json:"threadId"`
	SenderID   int64     `gorm:"column:sender_id" json:"senderId"`
	Content    string    `gorm:"column:content;type:varchar(2048);not null" json:"content"`
	ReadFlag   int32     `gorm:"column:read_flag" json:"readFlag"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (ChatMessage) TableName() string { return "tb_chat_message" }

type ChatGroup struct {
	ID          int64      `gorm:"column:id;primaryKey" json:"id"`
	OwnerID     int64      `gorm:"column:owner_id" json:"ownerId"`
	Name        string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Avatar      *string    `gorm:"column:avatar;type:varchar(512)" json:"avatar"`
	JoinMode    int32      `gorm:"column:join_mode" json:"joinMode"`
	CreateTime  time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	LastMsgTime *time.Time `gorm:"column:last_msg_time" json:"lastMsgTime"`
}

func (ChatGroup) TableName() string { return "tb_chat_group" }

type ChatGroupMember struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	GroupID       int64     `gorm:"column:group_id;uniqueIndex:uk_cgm_group_user" json:"groupId"`
	UserID        int64     `gorm:"column:user_id;uniqueIndex:uk_cgm_group_user" json:"userId"`
	Role          int32     `gorm:"column:role" json:"role"`
	MemberStatus  int32     `gorm:"column:member_status" json:"memberStatus"`
	CreateTime    time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	LastReadMsgID *int64    `gorm:"column:last_read_msg_id" json:"lastReadMsgId"`
}

func (ChatGroupMember) TableName() string { return "tb_chat_group_member" }

type ChatGroupJoinRequest struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	GroupID     int64     `gorm:"column:group_id;uniqueIndex:uk_cgj_group_applicant" json:"groupId"`
	ApplicantID int64     `gorm:"column:applicant_id;uniqueIndex:uk_cgj_group_applicant" json:"applicantId"`
	Status      int32     `gorm:"column:status" json:"status"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (ChatGroupJoinRequest) TableName() string { return "tb_chat_group_join_request" }

type GroupMessage struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	GroupID    int64     `gorm:"column:group_id" json:"groupId"`
	SenderID   int64     `gorm:"column:sender_id" json:"senderId"`
	Content    string    `gorm:"column:content;type:varchar(2048);not null" json:"content"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (GroupMessage) TableName() string { return "tb_group_message" }
