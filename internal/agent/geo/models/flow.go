/*
 * Copyright 2025 Peanut Authors
 *
 * GEO Flow State - Flow 模式状态定义
 */

package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/compose"
)

func init() {
	_ = compose.RegisterSerializableType[FlowState]("GEOFlowState")
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(step int, total int, agentName string, message string)

// FlowState GEO Flow 状态
type FlowState struct {
	// 输入参数
	URL          string `json:"url,omitempty"`
	PlatformType string `json:"platform_type,omitempty"`

	// 步骤 1: 网页爬取结果
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`

	// 步骤 2: 查询发散结果
	QueryFanout   []string       `json:"query_fanout,omitempty"`
	SearchResults []SearchResult `json:"search_results,omitempty"`

	// 步骤 3: 主查询提取
	MainQuery    string   `json:"main_query,omitempty"`
	Keywords     []string `json:"keywords,omitempty"`
	SearchIntent string   `json:"search_intent,omitempty"`

	// 步骤 4: AI 摘要
	AIOverview string   `json:"ai_overview,omitempty"`
	Sources    []string `json:"sources,omitempty"`

	// 步骤 5: 查询总结
	QuerySummary string `json:"query_summary,omitempty"`

	// 步骤 6: 优化报告
	Report *OptimizationReport `json:"report,omitempty"`

	// 步骤 7: 重写后的文章
	OptimizedArticle string `json:"optimized_article,omitempty"`

	// 流程控制
	Goto       string `json:"goto,omitempty"`
	Step       int    `json:"step,omitempty"`
	MaxSteps   int    `json:"max_steps,omitempty"`
	TotalSteps int    `json:"total_steps,omitempty"` // 总步骤数（用于进度计算）

	// 错误处理
	LastError string `json:"last_error,omitempty"`

	// 进度回调（不序列化）
	OnProgress ProgressCallback `json:"-"`
}

// SearchResult 搜索结果条目
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// MarshalJSON 序列化
func (s *FlowState) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}

// UnmarshalJSON 反序列化
func (s *FlowState) UnmarshalJSON(b []byte) error {
	type Alias FlowState
	var tmp Alias
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	*s = FlowState(tmp)
	return nil
}

// GenFlowState 生成 Flow State 的工厂函数
func GenFlowState(ctx context.Context) *FlowState {
	fmt.Println("[GEO] GenFlowState 被调用")
	return &FlowState{
		Goto: "title_scraper", // 默认从 title_scraper 开始
	}
}

// GEOCheckPoint GEO Flow 的全局状态存储点
// 实现 CheckPointStore 接口，用 checkPointID 进行索引
type GEOCheckPoint struct {
	buf map[string][]byte
}

// Get 获取 checkpoint 数据
func (gc *GEOCheckPoint) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	data, ok := gc.buf[checkPointID]
	return data, ok, nil
}

// Set 设置 checkpoint 数据
func (gc *GEOCheckPoint) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	fmt.Printf("[GEOCheckPoint] Set called: id=%s, len=%d\n", checkPointID, len(checkPoint))
	gc.buf[checkPointID] = checkPoint
	return nil
}

// NewGEOCheckPoint 创建一个全局状态存储点实例
func NewGEOCheckPoint(ctx context.Context) compose.CheckPointStore {
	return &GEOCheckPoint{
		buf: make(map[string][]byte),
	}
}
