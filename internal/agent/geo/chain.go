package geo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/solariswu/peanut/internal/agent/geo/agents"
	"github.com/solariswu/peanut/internal/agent/geo/llm"
	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/parser"
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// NewChain 创建新的 GEO Chain（使用 Eino ADK）
func NewChain() (adk.Agent, error) {
	// 初始化工具
	scraper := tools.NewHTTPScraper()

	// 使用免费的 DuckDuckGo 搜索器
	searcher, err := tools.NewDuckDuckGoSearcher()
	if err != nil {
		return nil, fmt.Errorf("创建 DuckDuckGo 搜索器失败: %w", err)
	}
	fmt.Println("✅ 使用 DuckDuckGo 免费搜索")

	// 优先使用豆包模型，如果配置了 API key，则使用 DoubaoSERPProvider
	var serp tools.SERPProvider
	if doubaoSERP, err := tools.NewDoubaoSERPProvider(); err == nil {
		serp = doubaoSERP
		fmt.Println("✅ 使用豆包模型生成豆包元宝 AI 摘要")
	} else {
		// 回退到 DuckDuckGo
		ddgSERP, err := tools.NewDuckDuckGoSERPProvider()
		if err != nil {
			return nil, fmt.Errorf("创建 SERP provider 失败: %w", err)
		}
		serp = ddgSERP
		fmt.Println("✅ 使用 DuckDuckGo 免费搜索生成豆包元宝 AI 摘要")
		fmt.Printf("⚠️  未配置豆包 API key (%v)\n", err)
	}

	// 尝试初始化豆包 LLM
	var llmModel model.ToolCallingChatModel
	if arkModel, err := llm.NewArkChatModel(); err == nil {
		llmModel = arkModel
		fmt.Println("✅ 使用豆包模型进行智能分析")
	} else {
		fmt.Printf("⚠️  未配置豆包 LLM，使用默认模型 (%v)\n", err)
		llmModel = nil // 不设置模型，让 Eino 使用默认配置
	}

	// 初始化 7 个 Agent（包含内容重写）
	subAgents := []adk.Agent{
		agents.NewTitleScraperAgentWithModel(scraper, llmModel),
		agents.NewQueryFanoutResearcherAgentWithModel(searcher, llmModel),
		agents.NewMainQueryExtractorAgent(llmModel),
		agents.NewAIOverviewRetrieverAgentWithModel(serp, llmModel),
		agents.NewQueryFanoutSummarizerAgent(llmModel),
		agents.NewAIContentOptimizerAgent(llmModel),
		agents.NewContentRewriterAgent(llmModel),
	}

	// 创建 Sequential Agent
	agent, err := adk.NewSequentialAgent(context.Background(), &adk.SequentialAgentConfig{
		Name:        "GEOAgent",
		Description: "豆包元宝生成式引擎优化（GEO）智能体，7 步流程：分析→优化→内容重写",
		SubAgents:   subAgents,
	})

	if err != nil {
		return nil, fmt.Errorf("创建 Sequential Agent 失败: %w", err)
	}

	return agent, nil
}

// OutputDir 输出目录
const OutputDir = "output"

// ensureOutputDir 确保输出目录存在
func ensureOutputDir() error {
	return os.MkdirAll(OutputDir, 0755)
}

// saveIntermediateOutput 保存中间步骤输出到文件
func saveIntermediateOutput(step int, agentName string, content string, url string) error {
	if err := ensureOutputDir(); err != nil {
		return err
	}

	var filename string
	var title string

	// 根据 agent 名称确定输出文件名
	switch agentName {
	case "query_fanout_researcher":
		filename = "query_fanout.md"
		title = "# 查询发散结果\n\n"
	case "ai_overview_retriever":
		filename = "ai_overview.md"
		title = "# 豆包元宝 AI 摘要\n\n"
	case "query_fanout_summarizer":
		filename = "query_fanout_summary.md"
		title = "# 查询发散总结\n\n"
	case "ai_content_optimizer":
		filename = "report.md"
		title = fmt.Sprintf("# 豆包元宝 GEO 优化报告\n\n**URL**: %s\n\n**生成时间**: %s\n\n", url, time.Now().Format("2006-01-02 15:04:05"))
	case "content_rewriter":
		filename = "optimized_article.md"
		title = fmt.Sprintf("# 优化后的文章\n\n**原文 URL**: %s\n\n**生成时间**: %s\n\n", url, time.Now().Format("2006-01-02 15:04:05"))
	default:
		// 其他步骤不保存文件
		return nil
	}

	// 构建完整的 Markdown 内容
	markdown := title + content

	// 写入文件
	filepath := filepath.Join(OutputDir, filename)
	if err := os.WriteFile(filepath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("写入文件失败 %s: %w", filepath, err)
	}

	fmt.Printf("  📄 已保存: %s\n", filepath)
	return nil
}

// NewDefaultChain 使用默认配置创建 Chain
func NewDefaultChain() (adk.Agent, error) {
	return NewChain()
}

// Execute 执行 Chain
func Execute(ctx context.Context, agent adk.Agent, url string) (*models.OptimizationReport, error) {
	// 创建输出目录
	if err := ensureOutputDir(); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 构建查询
	query := fmt.Sprintf(`请分析这个 URL 的豆包元宝 GEO 优化潜力: %s

请完整执行以下 7 个步骤：
1. 爬取网页标题
2. 基于国内搜索（DuckDuckGo中文搜索）进行相关查询发散
3. 提取主查询
4. 模拟豆包元宝 AI 摘要生成
5. 总结查询发散
6. 生成豆包元宝优化报告
7. 根据优化建议生成优化后的文章

请以 Markdown 格式返回完整的分析报告和优化文章。`, url)

	// 执行 Agent
	iter := runner.Query(ctx, query)

	stepCount := 0
	var finalMessage *schema.Message
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			return nil, fmt.Errorf("执行失败: %w", event.Err)
		}

		// 获取消息
		msg, agentInfo, err := adk.GetMessage(event)
		if err == nil {
			stepCount++

			// 提取 agent 名称
			agentName := ""
			if agentInfo != nil {
				agentName = agentInfo.AgentName
			}

			// 如果无法提取 agent 名称，根据步骤推断
			if agentName == "" {
				agentNames := []string{
					"title_scraper",              // 步骤 1
					"query_fanout_researcher",    // 步骤 2
					"main_query_extractor",       // 步骤 3
					"ai_overview_retriever",      // 步骤 4
					"query_fanout_summarizer",    // 步骤 5
					"ai_content_optimizer",       // 步骤 6
					"content_rewriter",           // 步骤 7
				}
				if stepCount <= len(agentNames) {
					agentName = agentNames[stepCount-1]
				}
			}

			// 保存中间步骤输出
			if err := saveIntermediateOutput(stepCount, agentName, msg.Content, url); err != nil {
				fmt.Printf("⚠️  保存中间输出失败: %v\n", err)
			}

			finalMessage = msg
		}
	}

	// 解析响应为报告
	report := parseReport(finalMessage, url)

	return report, nil
}

// ExecuteWithStreaming 流式执行（带进度回调）
func ExecuteWithStreaming(ctx context.Context, agent adk.Agent, url string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	// 创建输出目录
	if err := ensureOutputDir(); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}
	fmt.Printf("📁 输出目录: %s/\n\n", OutputDir)

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 构建完整的查询（与 Execute 相同）
	query := fmt.Sprintf(`请分析这个 URL 的豆包元宝 GEO 优化潜力: %s

请完整执行以下 6 个步骤：
1. 爬取网页标题
2. 基于国内搜索（DuckDuckGo中文搜索）进行相关查询发散
3. 提取主查询
4. 模拟豆包元宝 AI 摘要生成
5. 总结查询发散
6. 生成豆包元宝优化报告

请以 Markdown 格式返回完整的分析报告。`, url)

	// 执行并收集进度
	iter := runner.Query(ctx, query)

	stepCount := 0
	var finalMessage *schema.Message

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			if callback != nil {
				callback(stepCount, "", fmt.Sprintf("错误: %v", event.Err))
			}
			return nil, fmt.Errorf("执行失败: %w", event.Err)
		}

		// 获取消息并报告进度
		msg, agentInfo, err := adk.GetMessage(event)
		if err == nil {
			stepCount++

			// 尝试提取 agent 名称
			agentName := ""
			if agentInfo != nil {
				agentName = agentInfo.AgentName
			}

			// 如果无法提取 agent 名称，根据步骤推断
			if agentName == "" {
				agentNames := []string{
					"title_scraper",              // 步骤 1
					"query_fanout_researcher",    // 步骤 2
					"main_query_extractor",       // 步骤 3
					"ai_overview_retriever",      // 步骤 4
					"query_fanout_summarizer",    // 步骤 5
					"ai_content_optimizer",       // 步骤 6
					"content_rewriter",           // 步骤 7
				}
				if stepCount <= len(agentNames) {
					agentName = agentNames[stepCount-1]
				}
			}

			// 生成友好的消息
			stepMessages := []string{
				"✅ 爬取网页标题完成",
				"✅ 搜索相关查询完成",
				"✅ 提取主查询完成",
				"✅ 获取豆包元宝 AI 摘要完成",
				"✅ 总结查询发散完成",
				"✅ 生成优化报告完成",
				"✅ 生成优化后文章完成",
			}
			message := ""
			if stepCount <= len(stepMessages) {
				message = stepMessages[stepCount-1]
			} else {
				message = fmt.Sprintf("✅ 步骤 %d 完成", stepCount)
			}

			if callback != nil {
				callback(stepCount, agentName, message)
			}

			// 保存中间步骤输出到文件（对齐 Python 项目）
			if err := saveIntermediateOutput(stepCount, agentName, msg.Content, url); err != nil {
				fmt.Printf("⚠️  保存中间输出失败: %v\n", err)
			}

			finalMessage = msg
		}
	}

	// 返回报告
	report := parseReport(finalMessage, url)

	return report, nil
}

// parseReport 解析消息为报告
func parseReport(msg *schema.Message, url string) *models.OptimizationReport {
	if msg == nil {
		return &models.OptimizationReport{
			URL:          url,
			OverallScore: 75,
			Title:        "未获取到标题",
		}
	}

	// 使用新的解析器解析完整的报告
	report := parser.ParseOptimizationReport(msg.Content, url)

	// 如果解析失败，返回基本报告
	if report == nil {
		return &models.OptimizationReport{
			URL:          url,
			OverallScore: 75,
			Title:        parser.ExtractTitle(msg.Content),
		}
	}

	return report
}

// parseScoreFromContent 从内容中解析评分（已弃用，使用 parser 包）
// 保留此函数以保持向后兼容
func parseScoreFromContent(content string) int {
	return parser.ExtractOverallScore(content)
}
