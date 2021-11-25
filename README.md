# wormhole
wormhole  (超高性能 安全带内网穿透)

![](./doc/Wormhole.png)

## 自签SSL证书

``` 
openssl req -newkey rsa:4096 \
            -x509 \
            -sha256 \
            -days 3650 \
            -nodes \
            -out proxy.crt \
            -keyout proxy.key
```

## 使用:

server:
``` 
./server -l client连接端口 -r 外网保留端口 -c proxy.crt -k proxy.key
```

client:
``` 
./client -l 本地地址 -r 内网穿透服务地址  -c proxy.crt -k proxy.key
```
