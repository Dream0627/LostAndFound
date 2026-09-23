// 本文件用“泛型”统一各列表接口的返回结构(list + 分页信息)，
// 避免为每种资源重复定义字段完全相同的 DTO。
// 约束采用“类型联集”而不是 any：只有项目内的这四类列表资源可用作元素类型，
// 既能复用同一套分页结构，又能在编译期挡住误用(例如把某个无关类型塞进来)。
package service

import "LAF/internal/model"

// PageItem 约束可分页列表的元素类型：目前为帖子/评论/对话/消息四类资源(均为指针)。
type PageItem interface {
	*model.Post | *model.Comment | *model.Conversation | *model.Message
}

// PageResult 是通用的分页返回结构，T 为列表元素类型(受 PageItem 约束)。
type PageResult[T PageItem] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
