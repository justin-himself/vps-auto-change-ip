# vps auto change ip

这是一个简单的 python 脚本, 目的是在 vps ip 被墙的时候自动切换 ip 并同步更新面板里对应的 ip.

_目前只支持 oracle 和 v2board. 其他提供商和面板的支持等待开发._

脚本会隔一段时间 ping 一下所有的 server, 如果发现 ping 不通, 那么把 ip 换掉. 因此 ping 的流量必须过墙.

如果脚本本身就在国内的话, 脚本自身就可以负责 ping, 那么直接运行脚本即可.

```bash
python3 app/main.py
```

如果脚本本身部署在墙外, 那么必须要部署一个 ping_server 部署在国内负责 ping.

```bash
# 部署在国内
uvicorn app/ping_server.py

# 部署在国外
python3 app/main.py
```

## docker

```bash
git clone https://github.com/justin-himself/vps-auto-change-ip
cd vps-auto-change-ip
mv config.example config
docker build -t justinhimself/vps-auto-change-ip .
```

> 部署程序本体:

```bash
docker run -d \
  --name vps_auto_change_ip \
  --restart unless-stopped \
  -v $PWD/config:/config \
  justinhimself/vps-auto-change-ip
```

> 部署 ping_server (optional)

```bash
docker run -d \
  --name vps_auto_change_ip_ping_server \
  --restart unless-stopped \
  -p 8000:8000 \
  -v $PWD/config:/config \
  justinhimself/vps-auto-change-ip \
  uvicorn app.ping_server:app --host 0.0.0.0
```

ping_server 的 config 只需要保留 ping 的内容就好.
