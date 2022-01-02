package server

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
	"bufio"
	"errors"
	"strings"
	"crypto/tls"
	"crypto/rand"
	"sync/atomic"
	"encoding/json"

	"github.com/GoTyro/pool-proxy/util"
)

type ProxyServer struct {
	config     *Config
	pool       int32                 //当前使用的矿池索引
	pools      []*PoolClient         //配置文件中获取的矿池列表
	sessions   map[*Session]struct{}
	timeout    time.Duration
	sessionsMu sync.RWMutex
}

type Session struct {
	sync.Mutex           //互斥锁
	ip     string        //矿机IP
	enc    *json.Encoder //json处理
	conn   net.Conn      //矿机连接标识
	pool   net.Conn      //矿池连接标识
	login  string        //矿机挖矿钱包账号
}

const (
	MaxReqSize = 1024
)

func NewProxy(cfg *Config) *ProxyServer{
	if len(cfg.Name) == 0 {
		log.Fatal("配置文件未配置'name'主键")
	}

	proxy := &ProxyServer{config: cfg}

	//矿池读取配置
	proxy.pools = make([]*PoolClient, len(cfg.Pool))
	for i, v := range cfg.Pool {
		log.Printf("读取矿池配置: %s => %s", v.Name, v.Address)

		protoAry := strings.Split(v.Address, "+")
		protocol := protoAry[0]
		if protocol != ProtoStratum {
			log.Fatalf("矿池地址 %s 请以 %s 开头", v.Address, ProtoStratum)
		}

		transAry := strings.Split(protoAry[1], "://")
		transport := transAry[0]

		if transport != TransportTCP && transport != TransportSSL {
			log.Fatalf("矿池协议类型仅支持 %s 或 %s", TransportTCP, TransportSSL)
		}

		hostAry := strings.Split(transAry[1], ":")
		host := hostAry[0]//地址
		port := hostAry[1]//端口

		//判断矿池地址是否正确
		_, err := net.ResolveIPAddr("ip4", host)
		if err != nil {
			log.Fatalf("矿池地址 %s 解析失败", host)
		}

		//去掉斜杠小尾巴
		if strings.HasSuffix(port, "/") {
			port = port[:len(port) - 1]
		}

		//保存上(标识,地址,端口,协议,超时)
		proxy.pools[i] = NewPoolClient(v.Name, host, port, transport, v.Timeout, proxy.config.Debug)
	}
	log.Printf("设置矿池连接: %s => %s", proxy.rpc().Name, proxy.rpc().Address)

	//session
	proxy.sessions = make(map[*Session]struct{})

	return proxy
}

//启动服务端
func (s *ProxyServer) Start() {
	var err error
	var listen net.Listener
	setKeepAlive := func(net.Conn) {}

	//配置超时
	s.timeout = util.MustParseDuration(s.config.Server.Timeout)

	if s.config.Server.TLS {
		//读入ssl数字证书文件
		crt, err := tls.LoadX509KeyPair(s.config.Server.CertFile, s.config.Server.KeyFile)
		if err != nil {
			log.Fatalln(err.Error())
		}

		//定义ssl配置
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = []tls.Certificate{crt}
		tlsConfig.Time = time.Now
		tlsConfig.Rand = rand.Reader

		//启动SSL侦听
		listen, err = tls.Listen("tcp", s.config.Server.Listen, tlsConfig)
	} else {
		//启动TCP侦听
		listen, err = net.Listen("tcp", s.config.Server.Listen)
		setKeepAlive = func(conn net.Conn) {
			conn.(*net.TCPConn).SetKeepAlive(true)
		}
	}

	if err != nil {
		log.Fatalf("代理程序启动失败,详情: %s", err)
	}
	log.Printf("代理程序已启动,侦听: %s", s.config.Server.Listen)

	defer listen.Close()
	var accept = make(chan int, s.config.Server.MaxConn)//有缓冲的channel
	n := 0//矿机连接数量

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("建立连接错误: %s", err)
			continue
		}

		setKeepAlive(conn)

		//连接矿池
		pool, err := s.rpc().ConnectionPool()
		if err != nil {
			log.Printf("矿池 %s 连接失败,详情: %s", s.rpc().Address, err.Error())
			continue
		}

		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())//获取客户端IP
		n += 1//连接数量+1
		cs := &Session{conn: conn, pool: pool, ip: ip}//矿机连接session

		//go s.getResponse(cs)
		go func(cs *Session) {
			s.getResponse(cs)
		}(cs)

		accept <- n//写channel
		go func(cs *Session) {
			err = s.handleMinerClient(cs)
			if err != nil {
				s.removeSession(cs)//销毁session
				conn.Close()//关闭矿机连接
				pool.Close()//关闭矿池连接
			}
			<-accept
		}(cs)
	}
}

func (s *ProxyServer) getResponse(cs *Session) {
	for {
		poolResp, err := s.rpc().ReadResponse(cs.pool)
		if err != nil {
			break
		}

		err =  cs.handleResponse(s, poolResp)
		if err != nil {
			break
		}
	}
}

//矿机消息处理流程
func (s *ProxyServer) handleMinerClient(cs *Session) error {
	cs.enc = json.NewEncoder(cs.conn)
	//设置超时
	s.setDeadline(cs.conn)
	buffer := bufio.NewReaderSize(cs.conn, MaxReqSize)

	for {
		data, isPrefix, err := buffer.ReadLine()
		if isPrefix {
			log.Printf("矿机连接频率异常,IP: %s", cs.ip)
			return err
		} else if err == io.EOF {
			log.Printf("矿机断开连接,IP: %s", cs.ip)
			return err
		} else if err != nil {
			log.Printf("矿机通信失败,IP: %s,详情: %v", cs.ip, err)
			return err
		}

		if len(data) > 1 {
			var req MinerReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				log.Printf("矿机异常请求数据,IP: %s,详情: %v", cs.ip, err)
				return err
			}

			if s.config.Debug {
				log.Printf("矿机消息 %s", string(data))
			}

			s.setDeadline(cs.conn)//解决矿机通信i/o timeout的问题

			//把矿机消息转发给矿池
			err = cs.handleRequest(s, &req)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//把矿机消息转发给矿池
func (cs *Session) handleRequest(s *ProxyServer, req *MinerReq) error {
	var params interface{}
	switch req.Method {
		case "eth_submitLogin":
			var tmp []string
			err := json.Unmarshal(req.Params, &tmp)
			if err != nil {
				log.Printf("矿机submitLogin请求数据异常,IP: %s", cs.ip)
				return err
			}

			if len(tmp) == 0 {
				return errors.New("Request Params Error")
			}

			login := strings.ToLower(tmp[0])
			cs.login = login
			s.registerSession(cs)
			log.Printf("矿工连接成功,账户: %v,IP: %v", login, cs.ip)

			params = map[string]interface{}{"id": req.Id, "jsonrpc":"2.0", "method": req.Method, "params": req.Params, "worker": req.Worker}
		case "eth_getWork":
			params = map[string]interface{}{"id": req.Id, "jsonrpc":"2.0", "method": req.Method, "params": req.Params}
		case "eth_submitWork":
			params = map[string]interface{}{"id": req.Id, "jsonrpc":"2.0", "method": req.Method, "params": req.Params, "worker": req.Worker}
		case "eth_submitHashrate":
			params = map[string]interface{}{"id": req.Id, "jsonrpc":"2.0", "method": req.Method, "params": req.Params, "worker": req.Worker}
		default:
			return errors.New("Method not found")
	}

	return s.rpc().SendRequest(cs.pool, params)
}

//把矿池反馈的消息发给矿机
func (cs *Session) handleResponse(s *ProxyServer, resp *PoolResp) error {
	result := string(*resp.Result)

	if result == "true" {
		return cs.sendTCPResult(s, *resp.Id, true)
	} else if result == "false" {
		return cs.sendTCPError(s, *resp.Id, &ErrorReply{Code: -1, Message: resp.Error["message"].(string)})
	} else {
		var reply []string
		err := json.Unmarshal(*resp.Result, &reply)
		if err != nil {
			return err
		}

		return cs.sendTCPResult(s, *resp.Id, reply)
	}
	return nil
}

//向矿机发送成功消息
func (cs *Session) sendTCPResult(s *ProxyServer, id json.RawMessage, result interface{}) error {
	cs.Lock()
	defer cs.Unlock()

	message := MinerResp{Id: id, Jsonrpc: "2.0", Result: result}
	/*if s.config.Debug {
		b, err := json.Marshal(&message)
		if err != nil {
			log.Printf("JSON ERR: %s", err)
		}
		log.Printf("发送矿机 %s", string(b))
	}*/

	return cs.enc.Encode(&message)
}

//向矿机发送失败消息
func (cs *Session) sendTCPError(s *ProxyServer, id json.RawMessage, reply *ErrorReply) error {
	cs.Lock()
	defer cs.Unlock()

	message := MinerResp{Id: id, Jsonrpc: "2.0", Error: reply}
	/*if s.config.Debug {
		b, err := json.Marshal(&message)
		if err != nil {
			log.Printf("JSON ERR: %s", err)
		}
		log.Printf("发送矿机 %s", string(b))
	}*/

	err := cs.enc.Encode(&message)
	if err != nil {
		return err
	}
	return errors.New(reply.Message)
}

//获取当前连接的矿池信息
func (s *ProxyServer) rpc() *PoolClient {
	i := atomic.LoadInt32(&s.pool)
	return s.pools[i]
}

func (self *ProxyServer) setDeadline(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(self.timeout))
}

//创建session
func (s *ProxyServer) registerSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[cs] = struct{}{}
}

//销毁session
func (s *ProxyServer) removeSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, cs)
}