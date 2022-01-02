## pool-proxy

### 本转发程序必需在可以正常访问外网的服务器上面运行,如需测试或使用,请保证您当前运行本程序的机器可以正常访问矿池服务器


[Telegram讨论组](https://t.me/PoolProxy)


#### 环境配置

* 更新系统

CentOS:

    yum update -y

Ubuntu:

    sudo apt-get update && sudo apt-get dist-upgrade -y

* 安装常用软件

CentOS:

    yum install curl git screen unzip wget ntp ntpdate -y

Ubuntu:

    sudo apt-get install curl git screen unzip wget ntp ntpdate -y

* 同步时间

CentOS:

    ntpdate time.nist.gov
    hwclock --systohc

Ubuntu:

    sudo ntpdate -s time.nist.gov
    sudo hwclock --systohc

* 安装Golang

判断Go是否已安装

    go version

如果有输出版本信息, 请跳过Golang安装这一步

CentOS:

    yum install golang -y

Ubuntu:

    sudo add-apt-repository ppa:longsleep/golang-backports
    sudo apt-get update
    sudo apt-get install golang-go -y

#### 安装

    git config --global http.sslVerify false && git config --global http.postBuffer 1048576000 && git config --global http.https://gopkg.in.followredirects true
    git clone https://github.com/GoTyro/pool-proxy.git && cd pool-proxy

#### 运行

直接运行:

    go run main.go

编译运行:

    go build .
    ./pool-proxy

注: 两种运行方式任选其一

#### 矿机端

    矿池地址改为: "stratum+tcp://你服务器的IP或你的域名:你服务器设置的侦听端口"
    注意: "你服务器的IP或你的域名:你服务器设置的侦听端口" 中间有个英文的冒号,不要填成中文全角符号,且格式为 "stratum+tcp://IP:端口" 中间不要加空格

#### 开启SSL
    
    本程序默认没有开启ssl, 如果需要开启ssl功能, 您需要免费自签一个SSL数字证书, 证书生成完毕后, 把 "XXXX.crt" 和 "XXXX.key" 这两个文件复制到 "certs" 文件夹, 并修改根目录里的 "config.json" 文件, 将 "tls" 选项设置为 true
    注意: 开启SSL后, 您的矿机端需要把矿池地址修改为 "stratum+ssl://你服务器的IP或你的域名:你服务器设置的侦听端口"
