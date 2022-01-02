package server

import ()

type Config struct {
	Name              string  `json:"name"`
	Threads           int     `json:"threads"`
	PoolCheckInterval string  `json:"poolCheckInterval"`
	Server            Server  `json:"server"`
	Pool              []Pool  `json:"pool"`
	Debug             bool    `json:"debug"`
}

type Server struct {
	Listen   string `json:"listen"`
	Timeout  string `json:"timeout"`
	MaxConn  int    `json:"maxConn"`
	TLS      bool   `json:"tls"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

type Pool struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Timeout  string `json:"timeout"`
}