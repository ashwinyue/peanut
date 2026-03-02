package progress

import (
	"sync"
)

// Progress 进度信息
type Progress struct {
	AnalysisID int64  `json:"analysis_id"`
	Step       int    `json:"step"`
	Total      int    `json:"total"`
	AgentName  string `json:"agent_name"`
	Message    string `json:"message"`
	Status     string `json:"status"` // pending, processing, completed, failed
	Score      int    `json:"score,omitempty"`
}

// Manager 进度管理器
type Manager struct {
	mu      sync.RWMutex
	chans   map[int64][]chan Progress
	current map[int64]*Progress
}

// NewManager 创建进度管理器
func NewManager() *Manager {
	return &Manager{
		chans:   make(map[int64][]chan Progress),
		current: make(map[int64]*Progress),
	}
}

// Subscribe 订阅进度更新
func (m *Manager) Subscribe(analysisID int64) chan Progress {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Progress, 10)
	m.chans[analysisID] = append(m.chans[analysisID], ch)

	// 如果已有进度，立即发送
	if p, ok := m.current[analysisID]; ok {
		go func() {
			ch <- *p
		}()
	}

	return ch
}

// Unsubscribe 取消订阅
// 注意：不关闭 channel，因为 Complete/Fail 可能已经关闭了它
// 让 GC 回收 channel
func (m *Manager) Unsubscribe(analysisID int64, ch chan Progress) {
	m.mu.Lock()
	defer m.mu.Unlock()

	channels := m.chans[analysisID]
	for i, c := range channels {
		if c == ch {
			// 从列表中移除，但不关闭 channel（避免重复关闭 panic）
			m.chans[analysisID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
}

// Update 更新进度
func (m *Manager) Update(analysisID int64, step int, total int, agentName string, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := Progress{
		AnalysisID: analysisID,
		Step:       step,
		Total:      total,
		AgentName:  agentName,
		Message:    message,
		Status:     "processing",
	}

	m.current[analysisID] = &p

	// 广播给所有订阅者
	for _, ch := range m.chans[analysisID] {
		select {
		case ch <- p:
		default:
			// channel full, skip
		}
	}
}

// Complete 标记完成
func (m *Manager) Complete(analysisID int64, score int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := Progress{
		AnalysisID: analysisID,
		Status:     "completed",
		Score:      score,
	}

	m.current[analysisID] = &p

	// 广播给所有订阅者并关闭 channel
	for _, ch := range m.chans[analysisID] {
		select {
		case ch <- p:
		default:
			// channel full, skip
		}
		close(ch)
	}
	delete(m.chans, analysisID)
}

// Fail 标记失败
func (m *Manager) Fail(analysisID int64, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := Progress{
		AnalysisID: analysisID,
		Status:     "failed",
		Message:    errMsg,
	}

	m.current[analysisID] = &p

	// 广播给所有订阅者并关闭 channel
	for _, ch := range m.chans[analysisID] {
		select {
		case ch <- p:
		default:
			// channel full, skip
		}
		close(ch)
	}
	delete(m.chans, analysisID)
}

// Get 获取当前进度
func (m *Manager) Get(analysisID int64) *Progress {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.current[analysisID]; ok {
		copy := *p
		return &copy
	}
	return nil
}
