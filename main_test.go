package main

import "testing"

func TestOperatorMessagesIncludeTraditionalChinese(t *testing.T) {
	if got := operatorMessage("en-US", "healthcheck-exit"); got == "" {
		t.Fatal("English operator message is empty")
	}
	if got := operatorMessage("zh-TW", "healthcheck-exit"); got != "健康檢查因錯誤而停止" {
		t.Fatalf("unexpected Traditional Chinese operator message: %q", got)
	}
}
