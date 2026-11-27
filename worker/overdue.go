// Package worker 放不跟 HTTP 请求同生命周期的事情。
// 目前只有一件：超期任务提醒扫描——定时把「过了截止时间、还没提醒过」的任务捞出来，
// 派给一组固定数量的 goroutine 并发提醒。
package worker

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"task-system/model"
)

// TaskStore 是扫描器需要的存储能力。用接口而不是直接拿 *sql.DB，
// 单测里塞个假实现就能跑，不用起 MySQL。
type TaskStore interface {
	// ScanOverdue 返回未完成 + 已过截止时间 + 还没提醒过的任务。
	ScanOverdue(ctx context.Context, limit int) ([]model.Task, error)
	// MarkReminded 写提醒时间（带 reminded_at IS NULL 条件，天然防重复）。
	MarkReminded(ctx context.Context, id int) (int64, error)
}

// Notifier 负责把「这条任务超期了」发出去：现在打到日志，将来换成邮件 / 企业微信 / webhook。
// 正因为提醒是外部 IO（可能几百毫秒），才需要并发 + 有界队列，而不是一条 UPDATE 就能替代。
type Notifier interface {
	Notify(ctx context.Context, task model.Task) error
}

// LogNotifier 是默认实现，只写日志。
type LogNotifier struct{}

func (LogNotifier) Notify(ctx context.Context, task model.Task) error {
	dueAt := ""
	if task.DueAt != nil {
		dueAt = task.DueAt.Format(time.DateTime)
	}
	log.Printf("[超期提醒] task_id=%d user_id=%d title=%q due_at=%s", task.ID, task.UserID, task.Title, dueAt)
	return nil
}

// Config 扫描参数。
type Config struct {
	Interval    time.Duration // 多久扫一次
	BatchSize   int           // 单次最多捞多少条（同时是 job 通道的缓冲上限）
	Workers     int           // 并发提醒的 goroutine 数：有界，防止一次捞上千条把下游打满
	ScanTimeout time.Duration // 单次扫描查库的超时
}

func (c Config) withDefaults() Config {
	if c.Interval <= 0 {
		c.Interval = time.Minute
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 50
	}
	if c.Workers <= 0 {
		c.Workers = 4
	}
	if c.ScanTimeout <= 0 {
		c.ScanTimeout = 5 * time.Second
	}
	return c
}

// Scanner 超期任务扫描器。
type Scanner struct {
	store    TaskStore
	notifier Notifier
	cfg      Config
	logf     func(format string, args ...any)
}

func NewScanner(store TaskStore, notifier Notifier, cfg Config) *Scanner {
	return &Scanner{
		store:    store,
		notifier: notifier,
		cfg:      cfg.withDefaults(),
		logf:     log.Printf,
	}
}

// Run 阻塞运行，直到 ctx 被取消才返回：
//  1. 先起 Workers 个 goroutine 消费 job 通道（固定并发，不会随任务数无限涨）；
//  2. 用 ticker 定时扫描，命中后把任务投进通道；
//  3. ctx 取消时先关通道，worker 把队列里剩下的处理完再退出——这就是优雅退出。
func (s *Scanner) Run(ctx context.Context) {
	jobs := make(chan model.Task, s.cfg.BatchSize)

	var wg sync.WaitGroup
	for i := 0; i < s.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs { // 通道关闭且排空后循环自然结束
				s.handle(ctx, task)
			}
		}()
	}

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	s.scanOnce(ctx, jobs) // 启动先扫一次，不用等第一个 tick

	for {
		select {
		case <-ctx.Done():
			close(jobs) // 不再投递新任务，让 worker 把在途的收尾
			wg.Wait()
			s.logf("超期扫描器已退出：%v", ctx.Err())
			return
		case <-ticker.C:
			s.scanOnce(ctx, jobs)
		}
	}
}

// scanOnce 扫一轮。投递用 select 兼顾 ctx：队列满且进程要退出时立刻放手，不卡住关闭流程。
func (s *Scanner) scanOnce(ctx context.Context, jobs chan<- model.Task) {
	scanCtx, cancel := context.WithTimeout(ctx, s.cfg.ScanTimeout)
	defer cancel()

	tasks, err := s.store.ScanOverdue(scanCtx, s.cfg.BatchSize)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		s.logf("扫描超期任务失败：%v", err)
		return
	}
	if len(tasks) == 0 {
		return
	}
	s.logf("本轮扫到 %d 条超期任务", len(tasks))

	for _, task := range tasks {
		select {
		case jobs <- task:
		case <-ctx.Done():
			return
		}
	}
}

// handle 提醒单条任务：提醒成功才写 reminded_at。
// 写失败或提醒失败都不标记，下一轮会重扫到同一条，宁可能重复也不漏提醒。
func (s *Scanner) handle(ctx context.Context, task model.Task) {
	if err := s.notifier.Notify(ctx, task); err != nil {
		s.logf("提醒失败 task_id=%d：%v", task.ID, err)
		return
	}
	if _, err := s.store.MarkReminded(ctx, task.ID); err != nil {
		s.logf("标记已提醒失败 task_id=%d：%v", task.ID, err)
	}
}
