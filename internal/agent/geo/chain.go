package geo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/solariswu/peanut/internal/agent/geo/agents"
	"github.com/solariswu/peanut/internal/agent/geo/llm"
	"github.com/solariswu/peanut/internal/agent/geo/models"
	"github.com/solariswu/peanut/internal/agent/geo/parser"
	"github.com/solariswu/peanut/internal/agent/geo/tools"
)

// NewChain 创建新的 GEO Chain（使用 Eino ADK）
// 支持多平台：doubao/wechat/zhihu/xiaohongshu/wenxin/yuanbao
func NewChain(platform string) (adk.Agent, error) {
	ctx := context.Background()

	// 验证并设置默认平台
	platformType := models.PlatformType(platform)
	if !models.IsValidPlatform(platformType) {
		platformType = models.PlatformDoubao
	}
	fmt.Printf("🎯 目标平台: %s\n", models.GetPlatformName(platformType))

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
		fmt.Println("✅ 使用豆包模型生成 AI 摘要")
	} else {
		// 回退到 DuckDuckGo
		ddgSERP, err := tools.NewDuckDuckGoSERPProvider()
		if err != nil {
			return nil, fmt.Errorf("创建 SERP provider 失败: %w", err)
		}
		serp = ddgSERP
		fmt.Println("✅ 使用 DuckDuckGo 免费搜索生成 AI 摘要")
		fmt.Printf("⚠️  未配置豆包 API key (%v)\n", err)
	}

	// 使用统一的 Model 初始化
	llmModel, err := llm.NewChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("初始化 LLM 失败: %w", err)
	}
	fmt.Println("✅ 使用豆包模型进行智能分析")

	// 初始化 7 个 Agent（分析→优化→重写）
	subAgents := make([]adk.Agent, 0, 7)

	agent1, err := agents.NewTitleScraperAgentWithModel(scraper, llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 title_scraper agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent1)

	agent2, err := agents.NewQueryFanoutResearcherAgentWithModel(searcher, llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 query_fanout_researcher agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent2)

	agent3, err := agents.NewMainQueryExtractorAgent(llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 main_query_extractor agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent3)

	agent4, err := agents.NewAIOverviewRetrieverAgentWithModel(serp, llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 ai_overview_retriever agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent4)

	agent5, err := agents.NewQueryFanoutSummarizerAgent(llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 query_fanout_summarizer agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent5)

	agent6, err := agents.NewAIContentOptimizerAgent(llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 ai_content_optimizer agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent6)

	agent7, err := agents.NewContentRewriterAgent(llmModel)
	if err != nil {
		return nil, fmt.Errorf("创建 content_rewriter agent 失败: %w", err)
	}
	subAgents = append(subAgents, agent7)

	// 创建 Sequential Agent
	agent, err := adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        "GEOAgent",
		Description: fmt.Sprintf("%s 生成式引擎优化（GEO）智能体，7 步流程：爬取→搜索→提取→摘要→总结→优化→重写", models.GetPlatformName(platformType)),
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
	case "content_validator":
		filename = "validation_result.md"
		title = fmt.Sprintf("# 优化效果验证\n\n**原文 URL**: %s\n\n**验证时间**: %s\n\n", url, time.Now().Format("2006-01-02 15:04:05"))
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

// NewDefaultChain 使用默认配置创建 Chain（豆包元宝平台）
func NewDefaultChain() (adk.Agent, error) {
	return NewChain("doubao")
}

// Execute 执行 Chain（支持指定平台）
func Execute(ctx context.Context, agent adk.Agent, url string, platform string) (*models.OptimizationReport, error) {
	// 创建输出目录
	if err := ensureOutputDir(); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 验证平台
	platformType := models.PlatformType(platform)
	if !models.IsValidPlatform(platformType) {
		platformType = models.PlatformDoubao
	}
	platformName := models.GetPlatformName(platformType)

	// 构建查询
	query := fmt.Sprintf(`请分析这个 URL 的 %s GEO 优化潜力: %s

**目标平台**: %s
**平台类型**: %s

请完整执行以下 7 个步骤：
1. 爬取网页标题
2. 基于国内搜索（DuckDuckGo中文搜索）进行相关查询发散并总结
3. 提取主查询（记住目标平台类型: %s）
4. 模拟 %s AI 摘要生成
5. 生成 %s 优化报告（使用平台特定权重）
6. 根据优化建议生成优化后的文章（遵循 %s 内容规范）

请以 Markdown 格式返回完整的分析报告和优化文章。`, platformName, url, platformName, platform, platform, platformName, platformName, platformName)

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
					"query_fanout_researcher",    // 步骤 2（合并了总结）
					"main_query_extractor",       // 步骤 3
					"ai_overview_retriever",      // 步骤 4
					"ai_content_optimizer",       // 步骤 5
					"content_rewriter",           // 步骤 6
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

// ExecuteWithStreaming 流式执行（带进度回调，支持指定平台）
func ExecuteWithStreaming(ctx context.Context, agent adk.Agent, url string, platform string, callback func(step int, agentName string, message string)) (*models.OptimizationReport, error) {
	// 创建输出目录
	if err := ensureOutputDir(); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}
	fmt.Printf("📁 输出目录: %s/\n\n", OutputDir)

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	// 验证平台
	platformType := models.PlatformType(platform)
	if !models.IsValidPlatform(platformType) {
		platformType = models.PlatformDoubao
	}
	platformName := models.GetPlatformName(platformType)

	// 构建完整的查询
	query := fmt.Sprintf(`请分析这个 URL 的 %s GEO 优化潜力: %s

**目标平台**: %s
**平台类型**: %s

请完整执行以下 7 个步骤：
1. 爬取网页标题
2. 基于国内搜索（DuckDuckGo中文搜索）进行相关查询发散
3. 提取主查询（记住目标平台类型: %s）
4. 模拟 %s AI 摘要生成
5. 总结查询发散
6. 生成 %s 优化报告（使用平台特定权重）
7. 根据优化建议生成优化后的完整文章（遵循 %s 内容规范）

请以 Markdown 格式返回完整的分析报告和优化文章。`, platformName, url, platformName, platform, platform, platformName, platformName, platformName)

	// 执行并收集进度
	iter := runner.Query(ctx, query)

	stepCount := 0
	// 收集每个步骤的输出，用于后续合并解析
	stepOutputs := make(map[string]string)
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
					"query_fanout_researcher",    // 步骤 2（合并了总结）
					"main_query_extractor",       // 步骤 3
					"ai_overview_retriever",      // 步骤 4
					"ai_content_optimizer",       // 步骤 5
					"content_rewriter",           // 步骤 6
				}
				if stepCount <= len(agentNames) {
					agentName = agentNames[stepCount-1]
				}
			}

			// 生成友好的消息
			stepMessages := []string{
				"✅ 爬取网页标题完成",
				"✅ 搜索并总结相关查询完成",
				"✅ 提取主查询完成",
				"✅ 获取 AI 摘要完成",
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

			// 收集每个步骤的输出
			if agentName != "" {
				stepOutputs[agentName] = msg.Content
			}

			finalMessage = msg
		}
	}

	// 合并所有步骤的输出，构建完整的报告内容
	fullContent := buildFullContent(stepOutputs, finalMessage, url)

	// 返回报告
	report := parseReportFullContent(fullContent, url)

	return report, nil
}

// buildFullContent 合并所有步骤的输出为完整内容
func buildFullContent(stepOutputs map[string]string, finalMessage *schema.Message, url string) string {
	var content strings.Builder

	// 添加标题
	content.WriteString(fmt.Sprintf("# GEO 优化分析报告\n\n**URL**: %s\n\n", url))

	// 按顺序添加各个步骤的输出
	stepOrder := []string{
		"title_scraper",
		"query_fanout_researcher",
		"main_query_extractor",
		"ai_overview_retriever",
		"query_fanout_summarizer",
		"ai_content_optimizer",
		"content_rewriter",
		"content_validator",
	}

	sectionTitles := map[string]string{
		"title_scraper":           "网页标题",
		"query_fanout_researcher": "查询发散",
		"main_query_extractor":    "主查询",
		"ai_overview_retriever":   "AI 摘要",
		"query_fanout_summarizer": "查询发散总结",
		"ai_content_optimizer":    "优化报告",
		"content_rewriter":        "优化后的文章",
		"content_validator":       "优化效果验证",
	}

	for _, agentName := range stepOrder {
		if output, ok := stepOutputs[agentName]; ok && output != "" {
			title := sectionTitles[agentName]
			content.WriteString(fmt.Sprintf("\n## %s\n\n", title))
			content.WriteString(output)
			content.WriteString("\n")
		}
	}

	// 如果没有收集到步骤输出，使用 finalMessage
	if len(stepOutputs) == 0 && finalMessage != nil {
		content.WriteString(finalMessage.Content)
	}

	return content.String()
}

// parseReportFullContent 从完整内容解析报告
func parseReportFullContent(content, url string) *models.OptimizationReport {
	if content == "" {
		return &models.OptimizationReport{
			URL:          url,
			OverallScore: 75,
			Title:        "未获取到标题",
		}
	}

	// 使用解析器解析完整的报告
	report := parser.ParseOptimizationReport(content, url)

	// 如果解析失败，返回基本报告
	if report == nil {
		return &models.OptimizationReport{
			URL:          url,
			OverallScore: 75,
			Title:        parser.ExtractTitle(content),
		}
	}

	return report
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
