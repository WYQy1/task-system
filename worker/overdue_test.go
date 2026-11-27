package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-system/model"
)

// fakeStore 模拟任务表：MarkReminded 之后任务就从"待提醒"里消失，
// 和真实 SQL 的 WHERE reminded_at IS NULL 行为一致。
type fakeStore struct {
	mu         sync.Mutex
	pending    []model.Task
	reminded   map[int]int
	scanErr    error
	markCtxErr []error
}

func newFakeStore(tasks int) *fakeStore {
	pending := make([]model.Task, 0, tasks)
	due := time.Now().Add(-time.Hour)
	for i := 1; i <= tasks; i++ {
		pending = append(pending, model.Task{ID: i, UserID: 1, Title: "超期任务", DueAt: &due})
	}
	return &fakeStore{pending: pending, reminded: make(map[int]int)}
}

func (f *fakeStore) ScanOverdue(ctx context.Context, limit int) ([]model.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.scanErr != nil {
		return nil, f.scanErr
	}
	if len(f.pending) == 0 {
		return nil, nil
	}
	n := limit
	if n > len(f.pending) {
		n = len(f.pending)
	}
	return append([]model.Task(nil), f.pending[:n]...), nil
}

func (f *fakeStore) MarkReminded(ctx context.Context, id int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// 记下调用这一刻 ctx 的状态：真实 SQL 在 ctx 已取消时会直接返回错误
	f.markCtxErr = append(f.markCtxErr, ctx.Err())
	f.reminded[id]++
	for i, task := range f.pending {
		if task.ID == id {
			f.pending = append(f.pending[:i], f.pending[i+1:]...)
			break
		}
	}
	return 1, nil
}

func (f *fakeStore) markCtxErrors() []error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]error(nil), f.markCtxErr...)
}

func (f *fakeStore) remindedCount() map[int]int {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[int]int, len(f.reminded))
	for k, v := range f.reminded {
		out[k] = v
	}
	return out
}

// fakeNotifier 记录提醒了哪些任务，并统计并发峰值。
type fakeNotifier struct {
	mu          sync.Mutex
	notified    []int
	fail        bool
	delay       time.Duration
	inFlight    int32
	maxInFlight int32
	onNotify    chan<- int
}

func (n *fakeNotifier) Notify(ctx context.Context, task model.Task) error {
	cur := atomic.AddInt32(&n.inFlight, 1)
	for {
		old := atomic.LoadInt32(&n.maxInFlight)
		if cur <= old || atomic.CompareAndSwapInt32(&n.maxInFlight, old, cur) {
			break
		}
	}
	defer atomic.AddInt32(&n.inFlight, -1)

	if n.delay > 0 {
		time.Sleep(n.delay)
	}
	if n.onNotify != nil {
		select {
		case n.onNotify <- task.ID:
		default:
		}
	}
	if n.fail {
		return errors.New("通知渠道不可用")
	}
	n.mu.Lock()
	n.notified = append(n.notified, task.ID)
	n.mu.Unlock()
	return nil
}

func (n *fakeNotifier) notifiedIDs() []int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]int(nil), n.notified...)
}

// 收尾：取消 ctx 后 Run 必须在限定时间内返回（不泄漏 goroutine、不卡住关闭流程）。
func runAndStop(t *testing.T, scanner *Scanner) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("ctx 取消后扫描器没有在 2s 内退出")
		}
	})
	// 把 ctx 交给调用方控制
	t.Cleanup(cancel)
}

// 每条超期任务只提醒一次：提醒成功后写 reminded_at，下一轮扫描就不会再捞到它。
func TestScannerNotifiesEachOverdueTaskOnce(t *testing.T) {
	store := newFakeStore(3)
	notified := make(chan int, 10)
	notifier := &fakeNotifier{onNotify: notified}
	scanner := NewScanner(store, notifier, Config{Interval: 10 * time.Millisecond, Workers: 2, BatchSize: 10})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	got := make(map[int]bool)
	deadline := time.After(2 * time.Second)
	for len(got) < 3 {
		select {
		case id := <-notified:
			got[id] = true
		case <-deadline:
			t.Fatalf("2s 内只收到 %d 条提醒，期望 3 条", len(got))
		}
	}

	// 再等两轮，确认没有重复提醒
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	assert.Len(t, got, 3, "三条超期任务都应被提醒")
	reminded := store.remindedCount()
	require.Len(t, reminded, 3)
	for id, times := range reminded {
		assert.Equal(t, 1, times, "任务 %d 被重复标记", id)
	}
	assert.ElementsMatch(t, []int{1, 2, 3}, notifier.notifiedIDs())
}

// worker 数量是上限：并发提醒的峰值不能超过配置值。
func TestScannerBoundsConcurrency(t *testing.T) {
	const workers = 4
	store := newFakeStore(20)
	notifier := &fakeNotifier{delay: 30 * time.Millisecond}
	scanner := NewScanner(store, notifier, Config{
		Interval: 10 * time.Millisecond, Workers: workers, BatchSize: 20,
	})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	require.Eventually(t, func() bool {
		return len(store.remindedCount()) == 20
	}, 3*time.Second, 20*time.Millisecond, "20 条任务应全部提醒完")

	cancel()
	<-done

	peak := atomic.LoadInt32(&notifier.maxInFlight)
	assert.LessOrEqual(t, peak, int32(workers), "并发提醒数超过了 worker 上限")
	assert.Greater(t, peak, int32(1), "worker 池没有并发起来，配置没生效")
}

// 提醒失败不能标记 reminded_at，否则这条任务就永远被漏掉。
func TestScannerKeepsTaskWhenNotifyFails(t *testing.T) {
	store := newFakeStore(2)
	notified := make(chan int, 10)
	notifier := &fakeNotifier{fail: true, onNotify: notified}
	scanner := NewScanner(store, notifier, Config{Interval: 10 * time.Millisecond, Workers: 2, BatchSize: 10})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	select {
	case <-notified:
	case <-time.After(2 * time.Second):
		t.Fatal("2s 内没有触发提醒")
	}
	cancel()
	<-done

	assert.Empty(t, store.remindedCount(), "提醒失败时不应写 reminded_at")
}

// 扫描出错不能把整个后台循环带走：记日志，下一轮继续。
func TestScannerSurvivesScanError(t *testing.T) {
	store := newFakeStore(1)
	store.scanErr = errors.New("数据库连接抖动")
	notifier := &fakeNotifier{}
	scanner := NewScanner(store, notifier, Config{Interval: 10 * time.Millisecond, Workers: 1})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond) // 跑过好几轮报错
	select {
	case <-done:
		t.Fatal("扫描报错后后台循环不应退出")
	default:
	}

	// 数据库恢复后应能继续工作
	store.mu.Lock()
	store.scanErr = nil
	store.mu.Unlock()

	require.Eventually(t, func() bool {
		return len(notifier.notifiedIDs()) == 1
	}, 2*time.Second, 20*time.Millisecond, "数据库恢复后应继续提醒")

	cancel()
	<-done
}

// 空任务表 + 立刻取消：Run 必须马上返回，用于验证优雅退出不会卡在通道上。
func TestScannerStopsPromptlyOnCancel(t *testing.T) {
	store := newFakeStore(0)
	scanner := NewScanner(store, &fakeNotifier{}, Config{Interval: time.Hour, Workers: 2})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("取消后 1s 内没有退出")
	}
}

// blockingNotifier 卡在 Notify 里，直到测试放行；用来在"提醒已开始、尚未写回标记"的瞬间发退出信号。
type blockingNotifier struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (b *blockingNotifier) Notify(ctx context.Context, task model.Task) error {
	b.once.Do(func() { close(b.started) })
	<-b.release
	return nil
}

// 提醒已经发出去了，此时进程收到退出信号：reminded_at 仍必须写成功。
// 否则进程重启后会重新扫到同一条任务，用户收到重复提醒。
func TestMarkRemindedSurvivesCancel(t *testing.T) {
	store := newFakeStore(1)
	notifier := &blockingNotifier{started: make(chan struct{}), release: make(chan struct{})}
	scanner := NewScanner(store, notifier, Config{Interval: time.Hour, Workers: 1, BatchSize: 1})
	scanner.logf = t.Logf

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		scanner.Run(ctx)
		close(done)
	}()

	select {
	case <-notifier.started:
	case <-time.After(2 * time.Second):
		t.Fatal("扫描器没有在启动后立刻扫描")
	}

	cancel()                // 模拟收到 SIGTERM
	close(notifier.release) // 提醒动作此刻完成

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("扫描器没有在取消后退出")
	}

	require.Len(t, store.markCtxErrors(), 1)
	assert.NoError(t, store.markCtxErrors()[0], "写回 reminded_at 用的 ctx 不应是已取消的")
	assert.Equal(t, map[int]int{1: 1}, store.remindedCount())
}
