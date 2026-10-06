package service

import (
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

// AnnouncementListResult 是公告分页列表的返回结构。
type AnnouncementListResult = PageResult[*model.Announcement]

// AnnouncementInput 是“发表公告”的业务入参 DTO。
type AnnouncementInput struct {
	Title   string
	Content string
}

type AnnouncementService struct {
	repository *repository.AnnouncementRepository
}

// NewAnnouncementService 由 router 在装配阶段调用，注入公告仓库。
func NewAnnouncementService(repository *repository.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{
		repository: repository,
	}
}

// Create 发布公告。核心业务规则：
//  1. 标题去空白后长度需在 1~200；
//  2. 内容去空白后长度需在 1~2000；
//  3. 发布者以登录身份(adminID)为准，不信任请求体。
func (s *AnnouncementService) Create(adminID uint64, input AnnouncementInput) (*model.Announcement, error) {
	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) == 0 || len(input.Title) > 200 {
		return nil, apperror.ParamError
	}
	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 2000 {
		return nil, apperror.ParamError
	}
	announcement := &model.Announcement{
		Title:   input.Title,
		Content: input.Content,
		AdminID: adminID,
	}
	if err := s.repository.Create(announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

// DeleteAnnouncement 删除公告：先确认公告存在且未删除，再执行软删除。
func (s *AnnouncementService) DeleteAnnouncement(announcementID uint64) error {
	if _, err := s.repository.GetAnnouncementByID(announcementID); err != nil {
		return err
	}
	return s.repository.Delete(announcementID)
}

// GetAnnouncements 分页查询公告列表(公开)。
func (s *AnnouncementService) GetAnnouncements(page, pageSize int) (*AnnouncementListResult, error) {
	offset := pagination.Offset(page, pageSize)
	announcements, total, err := s.repository.GetAnnouncements(pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &AnnouncementListResult{
		List:     announcements,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
