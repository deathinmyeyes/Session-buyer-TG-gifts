package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	config := &AppConfig{
		LoggerLevel: "info",
		SoftConfig: SoftConfig{
			UpdateTicker: 30,
			RepoOwner:    "test-owner",
			RepoName:     "test-repo",
			ApiLink:      "https://api.github.com",
			TgSettings: TgSettings{
				AppId:              123456,
				ApiHash:            "test_api_hash",
				Phone:              "+1234567890",
				Password:           "test_password",
				TgBotKey:           "test_bot_key",
				NotificationChatID: 987654321,
			},
			GiftParam: GiftParam{
				TotalStarCap:  10000,
				TestMode:      false,
				LimitedStatus: true,
				ReleaseBy:     false,
			},
			Criterias: []Criterias{
				{
					MinPrice:    100,
					MaxPrice:    1000,
					TotalSupply: 50,
					Count:       5,
				},
			},
			Receiver: ReceiverParams{
				UserReceiverID:    []string{"987654321"},
				ChannelReceiverID: []string{"123456789"},
			},
			Ticker:               30.0,
			RetryCount:           3,
			RetryDelay:           5,
			MaxBuyCount:          10,
			ConcurrencyGiftCount: 5,
			ConcurrentOperations: 3,
			RPCRateLimit:         10,
		},
	}

	// Write config to file
	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Test loading config
	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, config.LoggerLevel, loadedConfig.LoggerLevel)
	assert.Equal(t, config.SoftConfig.UpdateTicker, loadedConfig.SoftConfig.UpdateTicker)
	assert.Equal(t, config.SoftConfig.RepoOwner, loadedConfig.SoftConfig.RepoOwner)
	assert.Equal(t, config.SoftConfig.RepoName, loadedConfig.SoftConfig.RepoName)
	assert.Equal(t, config.SoftConfig.ApiLink, loadedConfig.SoftConfig.ApiLink)
	assert.Equal(t, config.SoftConfig.TgSettings, loadedConfig.SoftConfig.TgSettings)
	assert.Equal(t, config.SoftConfig.GiftParam.TotalStarCap, loadedConfig.SoftConfig.GiftParam.TotalStarCap)
	assert.Equal(t, len(config.SoftConfig.Criterias), len(loadedConfig.SoftConfig.Criterias))
	assert.Equal(t, config.SoftConfig.Criterias[0], loadedConfig.SoftConfig.Criterias[0])
	assert.Equal(t, config.SoftConfig.Receiver, loadedConfig.SoftConfig.Receiver)
	assert.Equal(t, config.SoftConfig.Ticker, loadedConfig.SoftConfig.Ticker)
	assert.Equal(t, config.SoftConfig.RetryCount, loadedConfig.SoftConfig.RetryCount)
	assert.Equal(t, config.SoftConfig.RetryDelay, loadedConfig.SoftConfig.RetryDelay)
	assert.Equal(t, config.SoftConfig.GiftParam.TestMode, loadedConfig.SoftConfig.GiftParam.TestMode)
	assert.Equal(t, config.SoftConfig.MaxBuyCount, loadedConfig.SoftConfig.MaxBuyCount)
	assert.Equal(t, config.SoftConfig.GiftParam.LimitedStatus, loadedConfig.SoftConfig.GiftParam.LimitedStatus)
	assert.Equal(t, config.SoftConfig.ConcurrencyGiftCount, loadedConfig.SoftConfig.ConcurrencyGiftCount)
	assert.Equal(t, config.SoftConfig.ConcurrentOperations, loadedConfig.SoftConfig.ConcurrentOperations)
	assert.Equal(t, config.SoftConfig.RPCRateLimit, loadedConfig.SoftConfig.RPCRateLimit)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.json"

	config, err := LoadConfig(nonExistentPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.json")

	// Write invalid JSON to file
	invalidJSON := `{
		"logger_level": "info",
		"soft_config": {
			"tg_settings": {
				"app_id": 123456,
				"api_hash": "test_hash"
				// missing comma and closing braces
	`
	err := os.WriteFile(configPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty_config.json")

	// Write empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_MinimalConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "minimal_config.json")

	// Create minimal valid config
	minimalConfig := map[string]interface{}{
		"logger_level": "debug",
		"soft_config": map[string]interface{}{
			"tg_settings": map[string]interface{}{
				"app_id":   123456,
				"api_hash": "test_hash",
			},
			"ticker": 60.0,
		},
	}

	data, err := json.Marshal(minimalConfig)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "debug", config.LoggerLevel)
	assert.Equal(t, 123456, config.SoftConfig.TgSettings.AppId)
	assert.Equal(t, "test_hash", config.SoftConfig.TgSettings.ApiHash)
	assert.Equal(t, 60.0, config.SoftConfig.Ticker)
	// Other fields should have zero values
	assert.Equal(t, "", config.SoftConfig.TgSettings.Password)
	assert.Empty(t, config.SoftConfig.Criterias)
}

func TestLoadConfig_MultipleCriterias(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "multi_criteria_config.json")

	config := &AppConfig{
		LoggerLevel: "warn",
		SoftConfig: SoftConfig{
			TgSettings: TgSettings{
				AppId:    789012,
				ApiHash:  "multi_test_hash",
				Phone:    "+9876543210",
				Password: "multi_password",
			},
			Criterias: []Criterias{
				{
					MinPrice:    50,
					MaxPrice:    500,
					TotalSupply: 25,
					Count:       2,
				},
				{
					MinPrice:    1000,
					MaxPrice:    5000,
					TotalSupply: 100,
					Count:       10,
				},
				{
					MinPrice:    10000,
					MaxPrice:    50000,
					TotalSupply: 10,
					Count:       1,
				},
			},
			Receiver: ReceiverParams{
				UserReceiverID: []string{"111222333"},
			},
			Ticker: 15.0,
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, 3, len(loadedConfig.SoftConfig.Criterias))

	// Check first criteria
	assert.Equal(t, int64(50), loadedConfig.SoftConfig.Criterias[0].MinPrice)
	assert.Equal(t, int64(500), loadedConfig.SoftConfig.Criterias[0].MaxPrice)
	assert.Equal(t, int64(25), loadedConfig.SoftConfig.Criterias[0].TotalSupply)
	assert.Equal(t, int64(2), loadedConfig.SoftConfig.Criterias[0].Count)

	// Check second criteria
	assert.Equal(t, int64(1000), loadedConfig.SoftConfig.Criterias[1].MinPrice)
	assert.Equal(t, int64(5000), loadedConfig.SoftConfig.Criterias[1].MaxPrice)
	assert.Equal(t, int64(100), loadedConfig.SoftConfig.Criterias[1].TotalSupply)
	assert.Equal(t, int64(10), loadedConfig.SoftConfig.Criterias[1].Count)

	// Check third criteria
	assert.Equal(t, int64(10000), loadedConfig.SoftConfig.Criterias[2].MinPrice)
	assert.Equal(t, int64(50000), loadedConfig.SoftConfig.Criterias[2].MaxPrice)
	assert.Equal(t, int64(10), loadedConfig.SoftConfig.Criterias[2].TotalSupply)
	assert.Equal(t, int64(1), loadedConfig.SoftConfig.Criterias[2].Count)
}

func TestLoadConfig_ZeroValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zero_config.json")

	config := &AppConfig{
		LoggerLevel: "",
		SoftConfig: SoftConfig{
			Criterias: []Criterias{
				{
					MinPrice:    0,
					MaxPrice:    0,
					TotalSupply: 0,
					Count:       0,
				},
			},
			Receiver: ReceiverParams{
				UserReceiverID: []string{"0"},
			},
			Ticker: 0.0,
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, "", loadedConfig.LoggerLevel)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].MinPrice)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].MaxPrice)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].TotalSupply)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].Count)
	assert.Equal(t, 0.0, loadedConfig.SoftConfig.Ticker)
}

func TestLoadConfig_BooleanFields(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "boolean_config.json")

	config := &AppConfig{
		LoggerLevel: "info",
		SoftConfig: SoftConfig{
			GiftParam: GiftParam{
				TestMode:      true,
				LimitedStatus: false,
			},
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.True(t, loadedConfig.SoftConfig.GiftParam.TestMode)
	assert.False(t, loadedConfig.SoftConfig.GiftParam.LimitedStatus)
}

func TestLoadConfig_ComplexReceiverParams(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "receiver_config.json")

	config := &AppConfig{
		LoggerLevel: "info",
		SoftConfig: SoftConfig{
			Receiver: ReceiverParams{
				UserReceiverID:    []string{"111", "222", "333"},
				ChannelReceiverID: []string{"444", "555", "666"},
			},
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, []string{"111", "222", "333"}, loadedConfig.SoftConfig.Receiver.UserReceiverID)
	assert.Equal(t, []string{"444", "555", "666"}, loadedConfig.SoftConfig.Receiver.ChannelReceiverID)
}

func TestLoadConfig_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_permission_config.json")

	// Create file with content
	err := os.WriteFile(configPath, []byte(`{"logger_level": "info"}`), 0644)
	require.NoError(t, err)

	// Remove read permissions
	err = os.Chmod(configPath, 0000)
	require.NoError(t, err)

	// Restore permissions after test
	defer func() {
		os.Chmod(configPath, 0644)
	}()

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_AllFieldTypes(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "all_fields_config.json")

	config := &AppConfig{
		LoggerLevel: "error",
		SoftConfig: SoftConfig{
			UpdateTicker:         120,
			RepoOwner:            "github-user",
			RepoName:             "test-repo",
			ApiLink:              "https://custom-api.example.com",
			Ticker:               45.5,
			RetryCount:           7,
			RetryDelay:           10,
			MaxBuyCount:          25,
			ConcurrencyGiftCount: 8,
			ConcurrentOperations: 6,
			RPCRateLimit:         20,
			GiftParam: GiftParam{
				TotalStarCap:  50000,
				TestMode:      true,
				LimitedStatus: false,
				ReleaseBy:     true,
			},
			TgSettings: TgSettings{
				AppId:              999888,
				ApiHash:            "full_test_hash",
				Phone:              "+1122334455",
				Password:           "complex_password",
				TgBotKey:           "bot_key_123",
				NotificationChatID: -1001234567890,
			},
			Criterias: []Criterias{
				{
					MinPrice:    500,
					MaxPrice:    2000,
					TotalSupply: 75,
					Count:       15,
				},
				{
					MinPrice:    5000,
					MaxPrice:    10000,
					TotalSupply: 30,
					Count:       3,
				},
			},
			Receiver: ReceiverParams{
				UserReceiverID:    []string{"123", "456", "789"},
				ChannelReceiverID: []string{"111", "222"},
			},
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)

	// Verify all fields
	assert.Equal(t, config.LoggerLevel, loadedConfig.LoggerLevel)
	assert.Equal(t, config.SoftConfig.UpdateTicker, loadedConfig.SoftConfig.UpdateTicker)
	assert.Equal(t, config.SoftConfig.RepoOwner, loadedConfig.SoftConfig.RepoOwner)
	assert.Equal(t, config.SoftConfig.RepoName, loadedConfig.SoftConfig.RepoName)
	assert.Equal(t, config.SoftConfig.ApiLink, loadedConfig.SoftConfig.ApiLink)
	assert.Equal(t, config.SoftConfig.GiftParam.TotalStarCap, loadedConfig.SoftConfig.GiftParam.TotalStarCap)
	assert.Equal(t, config.SoftConfig.Ticker, loadedConfig.SoftConfig.Ticker)
	assert.Equal(t, config.SoftConfig.RetryCount, loadedConfig.SoftConfig.RetryCount)
	assert.Equal(t, config.SoftConfig.RetryDelay, loadedConfig.SoftConfig.RetryDelay)
	assert.Equal(t, config.SoftConfig.GiftParam.TestMode, loadedConfig.SoftConfig.GiftParam.TestMode)
	assert.Equal(t, config.SoftConfig.MaxBuyCount, loadedConfig.SoftConfig.MaxBuyCount)
	assert.Equal(t, config.SoftConfig.GiftParam.LimitedStatus, loadedConfig.SoftConfig.GiftParam.LimitedStatus)
	assert.Equal(t, config.SoftConfig.ConcurrencyGiftCount, loadedConfig.SoftConfig.ConcurrencyGiftCount)
	assert.Equal(t, config.SoftConfig.ConcurrentOperations, loadedConfig.SoftConfig.ConcurrentOperations)
	assert.Equal(t, config.SoftConfig.RPCRateLimit, loadedConfig.SoftConfig.RPCRateLimit)

	// Verify TgSettings
	assert.Equal(t, config.SoftConfig.TgSettings, loadedConfig.SoftConfig.TgSettings)

	// Verify Criterias
	assert.Equal(t, len(config.SoftConfig.Criterias), len(loadedConfig.SoftConfig.Criterias))
	for i, criteria := range config.SoftConfig.Criterias {
		assert.Equal(t, criteria, loadedConfig.SoftConfig.Criterias[i])
	}

	// Verify Receiver
	assert.Equal(t, config.SoftConfig.Receiver, loadedConfig.SoftConfig.Receiver)
}
