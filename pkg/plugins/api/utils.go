package api

import (
	"net/http"
	"strings"
)

// extractHeaders 从HTTP头中提取字段
func extractHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// containsAny 检查字符串是否包含任意一个子串
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(strings.ToLower(s), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
