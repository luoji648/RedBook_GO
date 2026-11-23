package service

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/model"
)

type GroupCreateRequest struct {
	Name     string  `json:"name"`
	JoinMode int32   `json:"joinMode"`
	Avatar   *string `json:"avatar"`
}

type GroupUpdateRequest struct {
	Name   string  `json:"name"`
	Avatar *string `json:"avatar"`
}

func (s *Service) GroupCreate(ownerID int64, req GroupCreateRequest) (*model.ChatGroup, error) {
	if req.Name == "" {
		return nil, errors.New("群名称不能为空")
	}
	group := model.ChatGroup{OwnerID: ownerID, Name: req.Name, JoinMode: req.JoinMode, Avatar: req.Avatar}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		return tx.Create(&model.ChatGroupMember{GroupID: group.ID, UserID: ownerID, Role: 1}).Error
	})
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *Service) GroupByOwner(ownerID int64, viewerID *int64) ([]map[string]any, error) {
	var groups []model.ChatGroup
	if err := s.DB.Where("owner_id = ?", ownerID).Order("create_time DESC").Find(&groups).Error; err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		list = append(list, s.groupFlat(g, viewerID))
	}
	return list, nil
}

func (s *Service) GroupMy(userID int64) ([]map[string]any, error) {
	var members []model.ChatGroupMember
	if err := s.DB.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(members))
	for _, m := range members {
		var group model.ChatGroup
		s.DB.First(&group, m.GroupID)
		vo := s.groupFlat(group, &userID)
		vo["removedFromGroup"] = m.MemberStatus != 0
		vo["unreadCount"] = s.groupUnread(group.ID, userID)
		list = append(list, vo)
	}
	return list, nil
}

func (s *Service) GroupJoin(groupID, userID int64) (string, error) {
	var group model.ChatGroup
	if err := s.DB.First(&group, groupID).Error; err != nil {
		return "", errors.New("群不存在")
	}
	var member model.ChatGroupMember
	err := s.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err == nil && member.MemberStatus == 0 {
		return "already", nil
	}
	if group.JoinMode == 0 {
		if err == nil {
			s.DB.Model(&member).Update("member_status", 0)
		} else {
			s.DB.Create(&model.ChatGroupMember{GroupID: groupID, UserID: userID, Role: 0})
		}
		return "joined", nil
	}
	var req model.ChatGroupJoinRequest
	err = s.DB.Where("group_id = ? AND applicant_id = ?", groupID, userID).First(&req).Error
	if err == nil {
		if req.Status == 1 {
			return "already", nil
		}
		return "pending", nil
	}
	req = model.ChatGroupJoinRequest{GroupID: groupID, ApplicantID: userID, Status: 0}
	s.DB.Create(&req)
	s.notifyGroupJoinRequest(group, req, userID)
	return "pending", nil
}

func (s *Service) GroupHandleJoinRequest(ownerID, requestID int64, approve bool) error {
	var req model.ChatGroupJoinRequest
	if err := s.DB.First(&req, requestID).Error; err != nil {
		return errors.New("申请不存在")
	}
	var group model.ChatGroup
	if err := s.DB.First(&group, req.GroupID).Error; err != nil {
		return err
	}
	if group.OwnerID != ownerID {
		return errors.New("无权处理该申请")
	}
	status := int32(2)
	if approve {
		status = 1
		var member model.ChatGroupMember
		err := s.DB.Where("group_id = ? AND user_id = ?", group.ID, req.ApplicantID).First(&member).Error
		if err == nil {
			s.DB.Model(&member).Update("member_status", 0)
		} else {
			s.DB.Create(&model.ChatGroupMember{GroupID: group.ID, UserID: req.ApplicantID, Role: 0})
		}
	}
	return s.DB.Model(&req).Update("status", status).Error
}

func (s *Service) GroupMeta(groupID, userID int64) (map[string]any, error) {
	var group model.ChatGroup
	if err := s.DB.First(&group, groupID).Error; err != nil {
		return nil, errors.New("群不存在")
	}
	var count int64
	s.DB.Model(&model.ChatGroupMember{}).Where("group_id = ? AND member_status = 0", groupID).Count(&count)
	var member model.ChatGroupMember
	removed := s.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error == nil && member.MemberStatus != 0
	vo := s.groupFlat(group, &userID)
	vo["memberCount"] = count
	vo["removedFromGroup"] = removed
	vo["inGroup"] = s.isGroupMember(groupID, userID)
	return vo, nil
}

func (s *Service) GroupMessages(groupID, userID, current, size int64) ([]map[string]any, int64, error) {
	if !s.isGroupMember(groupID, userID) {
		return nil, 0, errors.New("请先加入群聊")
	}
	var msgs []model.GroupMessage
	var total int64
	s.DB.Model(&model.GroupMessage{}).Where("group_id = ?", groupID).Count(&total)
	if err := s.DB.Where("group_id = ?", groupID).Order("id DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		var sender model.User
		s.DB.First(&sender, m.SenderID)
		list = append(list, map[string]any{
			"id":             m.ID,
			"groupId":        m.GroupID,
			"senderId":       m.SenderID,
			"content":        m.Content,
			"createTime":     m.CreateTime,
			"senderNickName": sender.NickName,
			"senderIcon":     sender.Icon,
		})
	}
	return list, total, nil
}

func (s *Service) GroupSend(groupID, userID int64, content string) (int64, error) {
	if content == "" {
		return 0, errors.New("消息不能为空")
	}
	if !s.isGroupMember(groupID, userID) {
		return 0, errors.New("请先加入群聊")
	}
	msg := model.GroupMessage{GroupID: groupID, SenderID: userID, Content: content}
	if err := s.DB.Create(&msg).Error; err != nil {
		return 0, err
	}
	s.DB.Model(&model.ChatGroup{}).Where("id = ?", groupID).Update("last_msg_time", time.Now())
	return msg.ID, nil
}

func (s *Service) GroupMarkRead(groupID, userID int64) error {
	var member model.ChatGroupMember
	if err := s.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return err
	}
	var last model.GroupMessage
	s.DB.Where("group_id = ?", groupID).Order("id DESC").First(&last)
	if last.ID > 0 {
		return s.DB.Model(&member).Update("last_read_msg_id", last.ID).Error
	}
	return nil
}

func (s *Service) GroupDetail(groupID, userID int64) (map[string]any, error) {
	var group model.ChatGroup
	if err := s.DB.First(&group, groupID).Error; err != nil {
		return nil, errors.New("群不存在")
	}
	var members []model.ChatGroupMember
	s.DB.Where("group_id = ? AND member_status = 0", groupID).Find(&members)
	memberVOs := make([]map[string]any, 0, len(members))
	for _, m := range members {
		var user model.User
		s.DB.First(&user, m.UserID)
		memberVOs = append(memberVOs, map[string]any{
			"userId":   user.ID,
			"nickName": user.NickName,
			"icon":     user.Icon,
			"isOwner":  m.Role == 1,
		})
	}
	var count int64
	s.DB.Model(&model.ChatGroupMember{}).Where("group_id = ? AND member_status = 0", groupID).Count(&count)
	vo := s.groupFlat(group, &userID)
	vo["members"] = memberVOs
	vo["memberCount"] = count
	vo["inGroup"] = s.isGroupMember(groupID, userID)
	return vo, nil
}

func (s *Service) groupFlat(g model.ChatGroup, viewerID *int64) map[string]any {
	inGroup := false
	if viewerID != nil {
		inGroup = s.isGroupMember(g.ID, *viewerID)
	}
	return map[string]any{
		"id":          g.ID,
		"ownerId":     g.OwnerID,
		"name":        g.Name,
		"avatar":      g.Avatar,
		"joinMode":    g.JoinMode,
		"createTime":  g.CreateTime,
		"lastMsgTime": g.LastMsgTime,
		"inGroup":     inGroup,
	}
}

func (s *Service) GroupKick(ownerID, groupID, targetID int64) error {
	var group model.ChatGroup
	if err := s.DB.First(&group, groupID).Error; err != nil {
		return err
	}
	if group.OwnerID != ownerID {
		return errors.New("仅群主可移出成员")
	}
	if targetID == ownerID {
		return errors.New("不能移出群主")
	}
	res := s.DB.Model(&model.ChatGroupMember{}).Where("group_id = ? AND user_id = ?", groupID, targetID).
		Update("member_status", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("成员不存在")
	}
	return nil
}

func (s *Service) GroupUpdateProfile(ownerID, groupID int64, req GroupUpdateRequest) error {
	var group model.ChatGroup
	if err := s.DB.First(&group, groupID).Error; err != nil {
		return err
	}
	if group.OwnerID != ownerID {
		return errors.New("仅群主可修改群资料")
	}
	updates := map[string]any{"name": req.Name, "avatar": req.Avatar}
	return s.DB.Model(&group).Updates(updates).Error
}

func (s *Service) isGroupMember(groupID, userID int64) bool {
	var member model.ChatGroupMember
	return s.DB.Where("group_id = ? AND user_id = ? AND member_status = 0", groupID, userID).First(&member).Error == nil
}

func (s *Service) notifyGroupJoinRequest(group model.ChatGroup, req model.ChatGroupJoinRequest, applicantID int64) {
	payload := map[string]any{
		"requestId":   req.ID,
		"groupId":     group.ID,
		"groupName":   group.Name,
		"applicantId": applicantID,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	threadID, err := s.ChatEnsureThread(applicantID, group.OwnerID)
	if err != nil {
		return
	}
	now := time.Now()
	msg := model.ChatMessage{
		ThreadID:   threadID,
		SenderID:   applicantID,
		Content:    "__RB_GROUP_JOIN__" + string(raw),
		CreateTime: now,
	}
	_ = s.DB.Create(&msg).Error
	_ = s.DB.Model(&model.ChatThread{}).Where("id = ?", threadID).Update("last_msg_time", now).Error
}
