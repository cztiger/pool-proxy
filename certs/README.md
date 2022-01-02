# 数字证书放在这个目录

把你生成的证书文件"XXX.crt"和"xxx.key"放在这个目录里

### Linux生成自签数字证书教程

请保证您的服务器已正确安装Golang且版本为`1.13+`, 如需查看本机Golang版本请使用 `go version`

依次执行
```
git clone https://github.com/square/certstrap.git
cd certstrap
go build
```
* 1.生成自信任的CA认证证书
```
certstrap init --common-name "ca" --expires "3 years"
```
命令完成后，会在当前目录下创建一个新的`out`目录，生成的证书都在该目录下, 文件分别为`ca.key` `ca.crl` `ca.crt`

* 2.生成服务端证书

执行
```
certstrap request-cert -cn server -ip 127.0.0.1 -domain "*.example.com"
```
然后按两次回车

多个ip或域名请用英文逗号分割,如:

`certstrap request-cert -cn server -ip "127.0.0.1,192.168.1.88" -domain "*.example.com,*.baidu.com"`

* 3.CA证书签名
