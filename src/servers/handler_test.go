package servers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bililive-go/bililive-go/src/configs"
	"github.com/bililive-go/bililive-go/src/instance"
)

func TestGetSoopLiveAuthConfigDoesNotExposeSavedPassword(t *testing.T) {
	cfg := configs.NewConfig()
	cfg.SoopLiveAuth.Username = "tester"
	cfg.SoopLiveAuth.Password = "secret"
	configs.SetCurrentConfig(cfg)

	recorder := httptest.NewRecorder()
	getSoopLiveAuthConfig(recorder, nil)

	assert.Equal(t, 200, recorder.Code)

	var resp commonResp
	err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "tester", data["username"])
	assert.Equal(t, true, data["has_saved_credentials"])
	_, exists := data["password"]
	assert.False(t, exists)
}

func TestGetPlatformStatsReportsDisabledRateLimit(t *testing.T) {
	cfg := configs.NewConfig()
	cfg.Interval = 1
	cfg.PlatformConfigs[configs.PlatformKeyDouyin] = configs.PlatformConfig{
		Name:                 "抖音",
		MinAccessIntervalSec: 60,
	}
	cfg.LiveRooms = []configs.LiveRoom{{
		Url:         "https://live.douyin.com/123456",
		IsListening: true,
	}}
	cfg.RefreshLiveRoomIndexCache()
	configs.SetCurrentConfig(cfg)
	t.Cleanup(func() { configs.SetCurrentConfig(configs.NewConfig()) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/config/platforms", nil)
	request = request.WithContext(context.WithValue(request.Context(), instance.Key, &instance.Instance{}))
	getPlatformStats(recorder, request)

	assert.Equal(t, 200, recorder.Code)
	var response struct {
		PlatformRateLimitEnabled bool `json:"platform_rate_limit_enabled"`
		Platforms                []struct {
			PlatformKey    string `json:"platform_key"`
			WarningMessage string `json:"warning_message"`
		} `json:"platforms"`
	}
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.False(t, response.PlatformRateLimitEnabled)
	if assert.Len(t, response.Platforms, 1) {
		assert.Equal(t, configs.PlatformKeyDouyin, response.Platforms[0].PlatformKey)
		assert.Empty(t, response.Platforms[0].WarningMessage)
	}
}

func TestNormalizeLiveURLInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "抖音完整分享文案",
			input: "4- 长按复制此条消息，打开抖音搜索，查看TA的更多作品。 https://v.douyin.com/DhWHNGVCxYk/7@5.com :5pm",
			want:  "https://v.douyin.com/DhWHNGVCxYk/",
		},
		{
			name:  "不带协议的抖音短链",
			input: "复制 v.douyin.com/AbC_123/ 后打开抖音",
			want:  "https://v.douyin.com/AbC_123/",
		},
		{
			name:  "普通分享文案",
			input: "直播地址：https://live.bilibili.com/12345。欢迎观看",
			want:  "https://live.bilibili.com/12345",
		},
		{
			name:  "普通网址保持不变",
			input: "https://www.douyu.com/67890",
			want:  "https://www.douyu.com/67890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeLiveURLInput(tt.input))
		})
	}
}
