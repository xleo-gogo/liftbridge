package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	client "github.com/liftbridge-io/liftbridge-api/go"
	proto "github.com/liftbridge-io/liftbridge/server/protocol"
)

// Ensure NewConfig properly parses config files.
func TestNewConfigFromFile(t *testing.T) {
	config, err := NewConfig("configs/full.yaml")
	require.NoError(t, err)

	require.Equal(t, "localhost", config.Listen.Host)
	require.Equal(t, 9293, config.Listen.Port)
	require.Equal(t, "0.0.0.0", config.Host)
	require.Equal(t, 5050, config.Port)
	require.Equal(t, uint32(5), config.LogLevel)
	require.True(t, config.LogRecovery)
	require.True(t, config.LogRaft)
	require.Equal(t, "/foo", config.DataDir)
	require.Equal(t, 10, config.BatchMaxMessages)
	require.Equal(t, time.Second, config.BatchMaxTime)
	require.Equal(t, time.Minute, config.MetadataCacheMaxAge)

	require.Equal(t, int64(1024), config.Streams.RetentionMaxBytes)
	require.Equal(t, int64(100), config.Streams.RetentionMaxMessages)
	require.Equal(t, time.Hour, config.Streams.RetentionMaxAge)
	require.Equal(t, time.Minute, config.Streams.CleanerInterval)
	require.Equal(t, int64(64), config.Streams.SegmentMaxBytes)
	require.Equal(t, time.Minute, config.Streams.SegmentMaxAge)
	require.True(t, config.Streams.Compact)
	require.Equal(t, 2, config.Streams.CompactMaxGoroutines)

	require.Equal(t, "foo", config.Clustering.ServerID)
	require.Equal(t, "bar", config.Clustering.Namespace)
	require.Equal(t, 10, config.Clustering.RaftSnapshots)
	require.Equal(t, uint64(100), config.Clustering.RaftSnapshotThreshold)
	require.Equal(t, 5, config.Clustering.RaftCacheSize)
	require.Equal(t, []string{"a", "b"}, config.Clustering.RaftBootstrapPeers)
	require.Equal(t, time.Minute, config.Clustering.ReplicaMaxLagTime)
	require.Equal(t, 30*time.Second, config.Clustering.ReplicaMaxLeaderTimeout)
	require.Equal(t, 2*time.Second, config.Clustering.ReplicaMaxIdleWait)
	require.Equal(t, 3*time.Second, config.Clustering.ReplicaFetchTimeout)
	require.Equal(t, 1, config.Clustering.MinISR)

	require.Equal(t, true, config.ActivityStream.Enabled)
	require.Equal(t, time.Minute, config.ActivityStream.PublishTimeout)
	require.Equal(t, client.AckPolicy_LEADER, config.ActivityStream.PublishAckPolicy)

	require.Equal(t, []string{"nats://localhost:4222"}, config.NATS.Servers)
	require.Equal(t, "user", config.NATS.User)
	require.Equal(t, "pass", config.NATS.Password)
}

// Ensure that default config is loaded.
func TestNewConfigDefault(t *testing.T) {
	config, err := NewConfig("")
	require.NoError(t, err)
	require.Equal(t, 512, config.Clustering.RaftCacheSize)
	require.Equal(t, "liftbridge-default", config.Clustering.Namespace)
	require.Equal(t, 1024, config.BatchMaxMessages)
}

// Ensure that both config file and default configs are loaded.
func TestNewConfigDefaultAndFile(t *testing.T) {
	config, err := NewConfig("configs/simple.yaml")
	require.NoError(t, err)
	// Ensure custom configs are loaded
	require.Equal(t, true, config.LogRecovery)
	require.Equal(t, int64(1024), config.Streams.RetentionMaxBytes)

	// Ensure also default values are loaded at the same time
	require.Equal(t, 512, config.Clustering.RaftCacheSize)
	require.Equal(t, "liftbridge-default", config.Clustering.Namespace)
	require.Equal(t, 1024, config.BatchMaxMessages)
}

// Ensure we can properly parse NATS username and password from a config file.
func TestNewConfigNATSAuth(t *testing.T) {
	config, err := NewConfig("configs/nats-auth.yaml")
	require.NoError(t, err)
	require.Equal(t, "admin", config.NATS.User)
	require.Equal(t, "password", config.NATS.Password)
}

// Ensure parsing host and listen.
func TestNewConfigListen(t *testing.T) {
	config, err := NewConfig("configs/listen-host.yaml")
	require.NoError(t, err)
	require.Equal(t, "192.168.0.1", config.Listen.Host)
	require.Equal(t, int(4222), config.Listen.Port)
	require.Equal(t, "my-host", config.Host)
	require.Equal(t, int(4333), config.Port)
}

// Ensure parsing TLS config.
func TestNewConfigTLS(t *testing.T) {
	config, err := NewConfig("configs/tls.yaml")
	require.NoError(t, err)
	require.Equal(t, "./configs/certs/server.key", config.TLSKey)
	require.Equal(t, "./configs/certs/server.crt", config.TLSCert)
}

// Ensure error is raised when given config file not found.
func TestNewConfigFileNotFound(t *testing.T) {
	_, err := NewConfig("somefile.yaml")
	require.Error(t, err)
}

// Ensure an error is returned when there is invalid configuration in listen.
func TestNewConfigInvalidClusteringSetting(t *testing.T) {
	_, err := NewConfig("configs/invalid-listen.yaml")
	require.Error(t, err)
}

// Ensure an error is returned when there is an unknown setting in the file.
func TestNewConfigUnknownSetting(t *testing.T) {
	_, err := NewConfig("configs/unknown-setting.yaml")
	require.Error(t, err)
}

// Ensure custom's StreamConfig can be parsed correctly
// if a given value is present in the custom's StreamConfig
// it should be set, otherwise, default values should be kept
func TestParseCustomStreamConfig(t *testing.T) {
	// Given custom stream config
	// duration configuration is in millisecond
	customStreamConfig := &proto.CustomStreamConfig{
		SegmentMaxBytes:      1024,
		SegmentMaxAge:        1000000,
		RetentionMaxBytes:    2048,
		RetentionMaxMessages: 1000,
		RetentionMaxAge:      1000000,
		CleanerInterval:      1000000,
		CompactMaxGoroutines: 10,
	}
	streamConfig := StreamsConfig{}

	streamConfig.ParseCustomStreamConfig(customStreamConfig)

	s, _ := time.ParseDuration("1000s")

	// Expect custom stream config overwrites default stream config
	require.Equal(t, int64(1024), streamConfig.SegmentMaxBytes)
	require.Equal(t, s, streamConfig.SegmentMaxAge)
	require.Equal(t, int64(2048), streamConfig.RetentionMaxBytes)
	require.Equal(t, int64(1000), streamConfig.RetentionMaxMessages)
	require.Equal(t, s, streamConfig.RetentionMaxAge)
	require.Equal(t, s, streamConfig.CleanerInterval)
	require.Equal(t, 10, streamConfig.CompactMaxGoroutines)

}

// Ensure default stream configs are always present,
// this should be the case when custom's stream configs are not set
func TestDefaultCustomStreamConfig(t *testing.T) {
	s, _ := time.ParseDuration("1000s")
	// Given a default stream config
	streamConfig := StreamsConfig{SegmentMaxBytes: 2048, SegmentMaxAge: s}

	// Given custom configs
	customStreamConfig := &proto.CustomStreamConfig{
		RetentionMaxBytes:    1024,
		RetentionMaxMessages: 1000,
		RetentionMaxAge:      1000000,
		CleanerInterval:      1000000,
		CompactMaxGoroutines: 10,
	}

	streamConfig.ParseCustomStreamConfig(customStreamConfig)

	// Ensure that in case of non-overlap values, default configs
	// remain present
	require.Equal(t, int64(2048), streamConfig.SegmentMaxBytes)
	require.Equal(t, s, streamConfig.SegmentMaxAge)

	// Ensure values from custom configs overwrite default configs
	require.Equal(t, int64(1024), streamConfig.RetentionMaxBytes)
	require.Equal(t, int64(1000), streamConfig.RetentionMaxMessages)
	require.Equal(t, s, streamConfig.RetentionMaxAge)
	require.Equal(t, s, streamConfig.CleanerInterval)
	require.Equal(t, 10, streamConfig.CompactMaxGoroutines)

}

// Ensure compact activation is correctly parsed
func TestCompactEnabledInCustomStreamConfig(t *testing.T) {
	// Given a default stream config
	streamConfig := StreamsConfig{}

	// Given custom configs with option to disable compact
	customStreamConfig := &proto.CustomStreamConfig{
		CompactEnabled: 2,
	}

	streamConfig.ParseCustomStreamConfig(customStreamConfig)

	// Ensure that stream config correctly disable compact option
	require.Equal(t, false, streamConfig.Compact)

	// Given a default stream config
	streamConfig2 := StreamsConfig{}
	// Given custom configs with option to disable compact
	customStreamConfig2 := &proto.CustomStreamConfig{
		CompactEnabled: 1,
	}

	streamConfig2.ParseCustomStreamConfig(customStreamConfig2)

	// Ensure that stream config correctly disable compact option
	require.Equal(t, true, streamConfig2.Compact)

	// Given a default stream config with default compaction disabled
	streamConfig3 := StreamsConfig{}

	// Given custom configs with NO option to configure compact
	customStreamConfig3 := &proto.CustomStreamConfig{}

	streamConfig3.ParseCustomStreamConfig(customStreamConfig3)

	// Ensure that stream default config is retained (by default compact.enabled is set
	// to true)
	require.Equal(t, true, streamConfig2.Compact)
}
