/*
 * Copyright 2025 Peanut Authors
 *
 * 工具函数
 */

package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetPromptTemplate 加载 prompt 模板文件
func GetPromptTemplate(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	possiblePaths := []string{
		filepath.Join(dir, "internal", "agent", "geo", "flow", "prompts", fmt.Sprintf("%s.md", name)),
		filepath.Join(dir, "prompts", fmt.Sprintf("%s.md", name)),
		filepath.Join(dir, "flow", "prompts", fmt.Sprintf("%s.md", name)),
	}

	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
	}

	return "", fmt.Errorf("prompt template not found: %s", name)
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}
