package model

import "time"

type Task struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	DueAt       *time.Time `json:"due_at,omitempty"`      // 截止时间，为空表示不设截止
	RemindedAt  *time.Time `json:"reminded_at,omitempty"` // 超期提醒时间，作为去重标记
}

// IsOverdue 判断任务在 at 时刻是否已超期：未完成、有截止时间、且截止时间早于 at。
func (t Task) IsOverdue(at time.Time) bool {
	if t.Status == "completed" || t.DueAt == nil {
		return false
	}
	return t.DueAt.Before(at)
}
