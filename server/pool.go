package server

import (
	"io"
	"fmt"
	"log"
	"net"
	"time"
	"sync"
	"bufio"
	"errors"
	"crypto/tls"
	"encoding/json"

	"github.com/GoTyro/pool-proxy/util"
)

const (
	ProtoStratum = "stratum"
	TransportTCP = "tcp"
	TransportSSL = "ssl"
)

type PoolClient struct {
	sync.RWMutex
	Name        string    //标识
	Address     string    //地址
	Transport   string    //协议,tcp/ssl
	Port        string    //端口
	Timeout     string    //超时
	Debug       bool      //调试模式
}

func NewPoolClient(name, host string, port string, proto string, timeout string, debug bool) *PoolClient {
	poolClient := &PoolClient{Name: name, Address: host, Port: port, Transport: proto, Timeout: timeout, Debug: debug}
	return poolClient
}

//连接矿池
func (r *PoolClient) ConnectionPool() (net.Conn, error) {
	var (
		err error
		conn net.Conn
	)

	//拼接矿池地址
	url := fmt.Sprintf("%s:%s", r.Address, r.Port)

	setKeepAlive := func(net.Conn) {}

	if (r.Transport == TransportSSL) {
		cfg :=  &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", url, cfg)
	} else {
		conn, err = net.Dial("tcp", url)
		setKeepAlive = func(conn net.Conn) {
			conn.(*net.TCPConn).SetKeepAlive(true)
		}
	}

	if err != nil {
		return nil, err
	}

	setKeepAlive(conn)

	return conn, nil
}

//发送矿机请求数据给矿池
func (r *PoolClient) SendRequest(conn net.Conn, params interface{}) error {
	Timeout := util.MustParseDuration(r.Timeout)
	if err := conn.SetWriteDeadline(time.Now().Add(Timeout)); err != nil {
		return err
	}

	enc := json.NewEncoder(conn)
	if err := enc.Encode(&params); err != nil {
		return err
	}

	return nil
}

//获取矿池反馈的消息
func (r *PoolClient) ReadResponse(conn net.Conn) (*PoolResp, error) {
	Timeout := util.MustParseDuration(r.Timeout)
	if err := conn.SetReadDeadline(time.Now().Add(Timeout)); err != nil {
		return nil, err
	}

	buffer := bufio.NewReaderSize(conn, MaxReqSize)
	data, isPrefix, err := buffer.ReadLine()
	if isPrefix {
		//log.Printf("矿池连接频率异常")
		return nil, err
	} else if err == io.EOF {
		//log.Printf("矿池断开连接")
		return nil, err
	} else if err != nil {
		//log.Printf("矿池通信失败,详情: %s", err)
		return nil, err
	}

	if len(data) > 1 {
		var resp *PoolResp
		err = json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("矿池异常反馈数据,详情: %s", err)
			return nil, err
		}

		if r.Debug {
			log.Printf("矿池反馈 %s", string(data))
		}

		if resp.Error != nil {
			return nil, errors.New(resp.Error["message"].(string))
		}

		return resp, err
	}
	return nil, errors.New("获取矿池消息失败")
}
