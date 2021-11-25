# wormhole
wormhole  (安全带内网穿透)

![](./doc/Wormhole.png)

## 创建证书

``` 
1.创建根证书私钥长度为2048
openssl genrsa -out ca.key 2048

2.利用私钥创建根证书按照提示一路输入：
openssl req -new -x509 -days 36500 -key ca.key -out ca.crt

3.创建长度为2048的SSL证书私匙
openssl genrsa -out server.key 2048

4.利用刚才的私匙建立SSL证书请求一路向下：
openssl req -new -key server.key -out server.csr

5.当前文件夹下运行如下命令创建所需目录：
mkdir dir demoCA &&cd demoCA&&mkdir newcerts&&echo '10' > serial &&touch index.txt&&cd ..

6.用CA根证书签署SSL自建证书：
openssl ca -in server.csr -out server.crt -cert ca.crt -keyfile ca.key

7.查看证书
openssl x509 -in server.crt -noout -text
```

## 自签SSL证书

``` 
openssl genrsa -des3 -out server.key 2048

openssl rsa -in server.pass.key -out server.key  // 去除密码

openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt

or:

openssl req -newkey rsa:4096 \
            -x509 \
            -sha256 \
            -days 3650 \
            -nodes \
            -out proxy.crt \
            -keyout proxy.key
```

## client 链接方式

1. 首次链接 建立 任务链接
2. 接受任务 再次建立 网络链接

## 使用:

server:
``` 
./server -l client连接端口 -r 外网保留端口 -c proxy.crt -k proxy.key
```

client:
``` 
./client -l 本地地址 -r 内网穿透服务地址  -c proxy.crt -k proxy.key
```
