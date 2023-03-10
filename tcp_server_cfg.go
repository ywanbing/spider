package spider

import (
	"crypto/tls"
	"time"
)

var defaultConnConfig = ConnConfig{
	maxSendMsgNum:     1000,
	maxRecvMsgNum:     10000,
	binaryPoolMinSize: 512,
	binaryPoolMaxSize: 512 * 1024,
	readTimeout:       3 * time.Second,
	writeTimeout:      3 * time.Second,
	onConnHandle: func(conn TcpConn) bool {
		return true
	},
}

type ConnConfig struct {
	// 发送消息缓冲区最大消息数量。默认值为1000。
	maxSendMsgNum int32
	// 接收消息缓冲区最大消息数量。默认值：10000。
	maxRecvMsgNum int32

	// 二进制数组对象池最小值。默认值：512 (0.5K)。
	binaryPoolMinSize int
	// 二进制数组对象池最大值。默认值：512 * 1024（512K) 。
	// NOTE：请根据实际的观测情况进行设置，以避免过多的内存占用。
	binaryPoolMaxSize int

	// 心跳控制 TODO
	HeartBeatOn        bool
	HeartBeatMessageID int32
	HeartBeatInterval  time.Duration

	// 创建连接是否允许的处理程序,
	// 如果返回false，则不允许创建连接；
	// 在这个函数中，可以对消息进行读取和写入。
	// 可用用于实现连接创建认证。
	onConnHandle func(conn TcpConn) bool

	// 读取和写入超时时间。默认值：0（不超时）。
	readTimeout  time.Duration
	writeTimeout time.Duration

	// If you want your tcp server using certs, using this field
	tlSConfig *tls.Config
}

type ConnConfigOption func(ConnConfig) ConnConfig

// WithMaxMsgNum sets the max [send|recv] message number.
// default: send=1000, recv=10000
func WithMaxMsgNum(send, recv int32) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		if send > 0 {
			cfg.maxSendMsgNum = send
		}
		if recv > 0 {
			cfg.maxRecvMsgNum = recv
		}
		return cfg
	}
}

// WithBinaryPoolSize sets the binary pool size.
// default: min=512, max=512*1024
func WithBinaryPoolSize(min, max int) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		if min > 0 {
			cfg.binaryPoolMinSize = min
		}
		if max > 0 {
			cfg.binaryPoolMaxSize = max
		}
		return cfg
	}
}

// WithOnConnHandle sets the onConnHandle.
func WithOnConnHandle(onConnHandle func(conn TcpConn) bool) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		cfg.onConnHandle = onConnHandle
		return cfg
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(d time.Duration) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		if d > 0 {
			cfg.readTimeout = d
		}
		return cfg
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(d time.Duration) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		if d > 0 {
			cfg.writeTimeout = d
		}
		return cfg
	}
}

// WithTLSConfig sets the tls config.
func WithTLSConfig(tlsConfig *tls.Config) ConnConfigOption {
	return func(cfg ConnConfig) ConnConfig {
		if tlsConfig != nil {
			cfg.tlSConfig = tlsConfig
		}
		return cfg
	}
}
