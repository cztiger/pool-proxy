## pool-proxy
简单的矿池代理程序,支持ssl和tcp

**更新系统**

CentOS: 

    yum update -y
    
Ubuntu:

    sudo apt-get update && sudo apt-get dist-upgrade -y

**安装常用软件**

CentOS:

    yum install curl git screen unzip wget ntp ntpdate -y

Ubuntu:

    sudo apt-get install curl git screen unzip wget ntp ntpdate -y
   
**同步时间**

CentOS:

    ntpdate time.nist.gov
    hwclock --systohc

Ubuntu:

    sudo ntpdate -s time.nist.gov
    sudo hwclock --systohc
    
