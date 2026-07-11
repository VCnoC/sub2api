// Package handler 测试对话广场的平台默认模型。
package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDefaultModelsForVideoPlatform(t *testing.T) {
	require.Equal(
		t,
		[]string{service.DefaultVideoPlatformModel, "grok-imagine-video-1.5-preview"},
		defaultModelsForPlatform(service.PlatformVideo),
	)
}
