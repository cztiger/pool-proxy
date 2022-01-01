## pool-proxy
简单的矿池代理(转发)程序,支持ssl和tcp

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

CentOS:

    yum install golang -y

Ubuntu:

    sudo add-apt-repository ppa:longsleep/golang-backports
    sudo apt-get update
    sudo apt-get install golang-go -y

