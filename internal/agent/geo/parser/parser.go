package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/solariswu/peanut/internal/agent/geo/models"
)

// ParseOptimizationReport 从 LLM 响应中解析完整的优化报告
func ParseOptimizationReport(content, url string) *models.OptimizationReport {
	report := &models.OptimizationReport{
		URL:        url,
		Timestamp:  time.Now(),
		Title:      ExtractTitle(content),
		MainQuery:  ExtractMainQuery(content),
		OverallScore: ExtractOverallScore(content),
	}

	// 提取对比表格
	report.ComparisonTable = extractComparisonTable(content)

	// 提取内容差距
	report.ContentGaps = extractContentGaps(content)

	// 提取优化建议
	report.OptimizationSuggestions = extractOptimizationSuggestions(content)

	// 如果没有提取到标题，使用 URL 生成一个
	if report.Title == "" {
		report.Title = generateTitleFromURL(url)
	}

	return report
}

// ExtractTitle 从内容中提取网页标题
func ExtractTitle(content string) string {
	// 查找 "标题:" 或 "Title:" 模式
	patterns := []string{
		`标题[:：]\s*([^\n]+)`,
		`Title[:：]\s*([^\n]+)`,
		`网页标题[:：]\s*([^\n]+)`,
		`H1[:：]\s*([^\n]+)`,
		`#{1}\s+([^\n]+)`, // Markdown 标题
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 如果没有找到明确的标题，查找第一行
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "|") {
			// 过滤掉一些明显的非标题行
			if !strings.Contains(line, "分析") && !strings.Contains(line, "报告") {
				return line
			}
		}
	}

	return ""
}

// ExtractMainQuery 提取主查询词
func ExtractMainQuery(content string) string {
	patterns := []string{
		`主查询[:：]\s*([^\n]+)`,
		`Main Query[:：]\s*([^\n]+)`,
		`核心查询[:：]\s*([^\n]+)`,
		`搜索词[:：]\s*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 尝试从 "查询:" 提取
	if strings.Contains(content, "查询:") {
		re := regexp.MustCompile(`查询[:：]\s*([^\n]+)`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// ExtractOverallScore 提取总体评分
func ExtractOverallScore(content string) int {
	patterns := []string{
		`总体评分[:：]\s*(\d+)`,
		`Overall Score[:：]\s*(\d+)`,
		`评分[:：]\s*(\d+)`,
		`Score[:：]\s*(\d+)`,
		`(\d+)/100`, // 匹配 "XX/100" 格式
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			var score int
			if _, err := regexp.MatchString(pattern, content); err == nil {
				// 提取数字
				scoreRe := regexp.MustCompile(`\d+`)
				scoreMatches := scoreRe.FindStringSubmatch(matches[1])
				if len(scoreMatches) > 0 {
					fmt.Sscanf(matches[1], "%d", &score)
					if score >= 0 && score <= 100 {
						return score
					}
				}
			}
		}
	}

	// 尝试从文本中提取所有 2-3 位数字，寻找可能是评分的
	re := regexp.MustCompile(`\b([5-9]\d|100)\b`)
	matches := re.FindAllString(content, -1)
	if len(matches) > 0 {
		// 取最后一个匹配（通常评分在最后）
		var score int
		fmt.Sscanf(matches[len(matches)-1], "%d", &score)
		if score >= 50 && score <= 100 {
			return score
		}
	}

	// 默认评分
	return 75
}

// extractComparisonTable 提取对比表格
func extractComparisonTable(content string) []models.ComparisonItem {
	var items []models.ComparisonItem

	// 查找 Markdown 表格
	lines := strings.Split(content, "\n")
	inTable := false
	headers := []string{}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 检测表格开始
		if strings.HasPrefix(line, "|") && strings.Contains(line, "维度") || strings.Contains(line, "Aspect") {
			inTable = true
			headers = strings.Split(strings.Trim(line, "|"), "|")
			// 清理表头
			for j, h := range headers {
				headers[j] = strings.TrimSpace(h)
			}
			continue
		}

		// 表格分隔行
		if inTable && strings.HasPrefix(line, "|") && strings.Contains(line, "---") {
			continue
		}

		// 表格数据行
		if inTable && strings.HasPrefix(line, "|") {
			// 检查是否是表格结束
			if !strings.Contains(line, "|") || strings.Trim(line, "|") == "" {
				inTable = false
				continue
			}

			cells := strings.Split(strings.Trim(line, "|"), "|")
			if len(cells) >= 5 { // 至少需要 5 列
				item := models.ComparisonItem{
					Dimension:   strings.TrimSpace(cells[0]),
					YourContent: strings.TrimSpace(cells[1]),
					AIOverview:  strings.TrimSpace(cells[2]),
					Similarity:  strings.TrimSpace(cells[3]),
					Difference:  strings.TrimSpace(cells[4]),
				}
				items = append(items, item)
			}
		} else if inTable && !strings.HasPrefix(line, "|") {
			inTable = false
		}

		// 如果连续 3 行都不是表格行，认为表格结束
		if inTable && i > 0 && !strings.HasPrefix(line, "|") {
			consecutiveNonTable := 0
			for j := i; j < len(lines) && j < i+3; j++ {
				if !strings.HasPrefix(strings.TrimSpace(lines[j]), "|") {
					consecutiveNonTable++
				}
			}
			if consecutiveNonTable >= 3 {
				inTable = false
			}
		}
	}

	return items
}

// extractContentGaps 提取内容差距
func extractContentGaps(content string) []string {
	var gaps []string

	// 查找 "内容差距" 部分的开始和结束
	startIdx := strings.Index(content, "## 内容差距")
	if startIdx == -1 {
		startIdx = strings.Index(content, "### 内容差距分析")
	}
	if startIdx == -1 {
		startIdx = strings.Index(content, "差距分析")
	}

	if startIdx != -1 {
		// 查找下一个 ## 或文档结尾
		endIdx := strings.Index(content[startIdx+3:], "\n##")
		if endIdx == -1 {
			endIdx = len(content)
		} else {
			endIdx += startIdx + 3
		}

		section := content[startIdx:endIdx]

		// 提取列表项
		lines := strings.Split(section, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// 匹配列表项（1. 或 - 或 *）
			if matched, _ := regexp.MatchString(`^\d+\.\s+`, line); matched {
				item := regexp.MustCompile(`^\d+\.\s+`).ReplaceAllString(line, "")
				item = strings.TrimSpace(item)
				if item != "" {
					gaps = append(gaps, item)
				}
			}
		}
	}

	return gaps
}

// extractOptimizationSuggestions 提取优化建议
func extractOptimizationSuggestions(content string) []models.OptimizationSuggestion {
	var suggestions []models.OptimizationSuggestion

	// 查找 "优化建议" 部分的开始
	startIdx := strings.Index(content, "## 优化建议")
	if startIdx == -1 {
		startIdx = strings.Index(content, "### 优化建议")
	}
	if startIdx == -1 {
		startIdx = strings.Index(content, "## 行动项")
	}

	if startIdx != -1 {
		// 查找下一个同级标题或文档结尾
		restContent := content[startIdx+3:] // 跳过 "## "
		endIdx := len(restContent)

		// 查找下一个 ## 标题（但不包括 ###）
		for _, prefix := range []string{"\n## ", "\n##\n"} {
			if idx := strings.Index(restContent, prefix); idx != -1 && idx < endIdx {
				// 确保不是 ###
				if idx+4 < len(restContent) && restContent[idx+4] != '#' {
					endIdx = idx
				}
			}
		}

		section := restContent[:endIdx]

		// 解析优先级（🔴 高优先级, 🟡 中优先级, 🟢 低优先级）
		lines := strings.Split(section, "\n")
		currentPriority := "medium"

		for _, line := range lines {
			line = strings.TrimSpace(line)

			// 检测优先级
			if strings.Contains(line, "🔴") || strings.Contains(line, "高优先级") {
				currentPriority = "high"
				continue
			}
			if strings.Contains(line, "🟡") || strings.Contains(line, "中优先级") {
				currentPriority = "medium"
				continue
			}
			if strings.Contains(line, "🟢") || strings.Contains(line, "低优先级") {
				currentPriority = "low"
				continue
			}

			// 提取建议项
			if matched, _ := regexp.MatchString(`^\d+\.\s+`, line); matched {
				item := regexp.MustCompile(`^\d+\.\s+`).ReplaceAllString(line, "")
				item = strings.TrimSpace(item)
				if item != "" {
					suggestion := models.OptimizationSuggestion{
						Priority:   currentPriority,
						Category:   "general",
						Issue:      "",
						Suggestion: item,
					}
					suggestions = append(suggestions, suggestion)
				}
			}
		}
	}

	return suggestions
}

// generateTitleFromURL 从 URL 生成标题
func generateTitleFromURL(url string) string {
	// 移除协议和 www
	url = strings.ReplaceAll(url, "https://", "")
	url = strings.ReplaceAll(url, "http://", "")
	url = strings.ReplaceAll(url, "www.", "")

	// 移除文件扩展名
	if idx := strings.LastIndex(url, "."); idx > 0 {
		url = url[:idx]
	}

	// 替换连字符和下划线为空格
	url = strings.ReplaceAll(url, "-", " ")
	url = strings.ReplaceAll(url, "_", " ")
	url = strings.ReplaceAll(url, "/", " / ")

	// 转换为标题格式（每个单词首字母大写）
	words := strings.Fields(url)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	title := strings.Join(words, " ")
	if len(title) > 60 {
		title = title[:60] + "..."
	}

	return title
}

// ExtractSources 提取来源列表
func ExtractSources(content string) []string {
	var sources []string

	// 查找 "来源:" 或 "Sources:" 部分的开始
	startIdx := strings.Index(content, "来源:")
	if startIdx == -1 {
		startIdx = strings.Index(content, "来源：")
	}
	if startIdx == -1 {
		startIdx = strings.Index(content, "Sources:")
	}

	if startIdx != -1 {
		// 查找下一个 ## 或两个连续换行或文档结尾
		restContent := content[startIdx:]
		endIdx := len(restContent)

		// 查找结束标记
		for _, marker := range []string{"\n\n", "\n##", "\n###"} {
			if idx := strings.Index(restContent, marker); idx != -1 && idx < endIdx {
				endIdx = idx
			}
		}

		section := restContent[:endIdx]

		// 提取方括号中的内容
		re := regexp.MustCompile(`\[([^\]]+)\]`)
		matches := re.FindStringSubmatch(section)
		if len(matches) > 1 {
			sourcesStr := matches[1]
			// 按逗号分割
			reSplit := regexp.MustCompile(`[,，]+`)
			parts := reSplit.Split(sourcesStr, -1)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" && strings.HasPrefix(part, "http") {
					sources = append(sources, part)
				}
			}
		}
	}

	// 如果没有找到明确的来源，尝试提取所有 URL
	if len(sources) == 0 {
		re := regexp.MustCompile(`https?://[^\s\])]+`)
		urlMatches := re.FindAllString(content, -1)
		sources = append(sources, urlMatches...)
	}

	return sources
}

// FormatAsMarkdown 将报告格式化为 Markdown
func FormatAsMarkdown(report *models.OptimizationReport) string {
	var sb strings.Builder

	sb.WriteString("# GEO 优化报告\n\n")
	sb.WriteString(fmt.Sprintf("## 基本信息\n\n"))
	sb.WriteString(fmt.Sprintf("- **URL**: %s\n", report.URL))
	sb.WriteString(fmt.Sprintf("- **标题**: %s\n", report.Title))
	sb.WriteString(fmt.Sprintf("- **主查询**: %s\n", report.MainQuery))
	sb.WriteString(fmt.Sprintf("- **总体评分**: %d/100\n", report.OverallScore))
	sb.WriteString(fmt.Sprintf("- **生成时间**: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05")))

	if len(report.ComparisonTable) > 0 {
		sb.WriteString("## 对比分析\n\n")
		sb.WriteString("| 维度 | 查询发散总结 | AI Overview | 相似点 | 差异 |\n")
		sb.WriteString("|------|-------------|-------------|--------|------|\n")
		for _, item := range report.ComparisonTable {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				item.Dimension,
				item.YourContent,
				item.AIOverview,
				item.Similarity,
				item.Difference))
		}
		sb.WriteString("\n")
	}

	if len(report.ContentGaps) > 0 {
		sb.WriteString("## 内容差距\n\n")
		for i, gap := range report.ContentGaps {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, gap))
		}
		sb.WriteString("\n")
	}

	if len(report.OptimizationSuggestions) > 0 {
		sb.WriteString("## 优化建议\n\n")

		// 按优先级分组
		highPriority := []models.OptimizationSuggestion{}
		mediumPriority := []models.OptimizationSuggestion{}
		lowPriority := []models.OptimizationSuggestion{}

		for _, s := range report.OptimizationSuggestions {
			switch s.Priority {
			case "high":
				highPriority = append(highPriority, s)
			case "medium":
				mediumPriority = append(mediumPriority, s)
			case "low":
				lowPriority = append(lowPriority, s)
			}
		}

		if len(highPriority) > 0 {
			sb.WriteString("### 🔴 高优先级\n\n")
			for i, s := range highPriority {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Suggestion))
			}
			sb.WriteString("\n")
		}

		if len(mediumPriority) > 0 {
			sb.WriteString("### 🟡 中优先级\n\n")
			for i, s := range mediumPriority {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Suggestion))
			}
			sb.WriteString("\n")
		}

		if len(lowPriority) > 0 {
			sb.WriteString("### 🟢 低优先级\n\n")
			for i, s := range lowPriority {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Suggestion))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
