// Package service 是“业务逻辑层”。
// 职责：做业务规则校验、权限判断，并把“仓库层的领域错误”翻译成“对外业务错误(apperror)”。
// 它不直接接触 SQL(那是 repository 的事)，也不处理 HTTP(那是 handler 的事)，
// 这样业务规则集中在一处，便于复用与测试。
package service

import (
	//"errors"
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

type AnnouncementListResult = PageResult[*model.Announcement]

type AnnouncementInput struct {
	Title   string
	Content string
}

type AnnouncementService struct {
	repository *repository.AnnouncementRepository
}

// NewAnnouncementService 由 router 在装配阶段调用，注入两个仓库。
func NewAnnouncementService(repository *repository.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{
		repository: repository,
	}
}

// Create 发布公告。核心业务规则：
//  1. 内容去空白后长度需在 1~2000；
//  2. 标题去空白后长度需在 1~100；
func (s *AnnouncementService) Create(adminID uint64, input AnnouncementInput) (*model.Announcement, error) {
	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 2000 {
		return nil, apperror.ParamError
	}
	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) == 0 || len(input.Title) > 100 {
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

func (s *AnnouncementService) DeleteAnnouncement(announcementID uint64) error {
	return s.repository.Delete(announcementID)
}

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
