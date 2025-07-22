package e2

type E2Config struct {
	SCTPPort       int    `yaml:"sctp_port"`
	LocalAddress   string `yaml:"local_address"`
	MaxConnections int    `yaml:"max_connections"`
	HeartbeatTimer int    `yaml:"heartbeat_timer"`
}
