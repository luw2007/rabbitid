# rabbitid
Distributed ID Generate Service

ID生成器

为了保证多机房部署数据的一致性，需要一个全局ID生成器。


```shell
git clone github.com/luw2007/rabbitid
cd rabbitid
go run cmd/idRedis/main.go
```

redis store
---
```shell
redis-server
```

zk store
---
使用docker-compose启动zk集群

```shell
zkCli
create /rabbitid/0 'master data center'
create /rabbitid 'rabbit id root'
cd tools/zoo && docker-compose up -d
```

使用命令行启动
```shell
# 修改 tool/zoo/Makefile 中的 ZKSERVER
vim tools/zoo/Makefile
make start
```

zk需要基础目录 /rabbitid/{机房ID}/{项目}
```shell
zkCli
create /rabbitid/0 'master data center'
create /rabbitid 'rabbit id root'

cd tools/zoo && docker-compose up -d
```

idHttp
---
- /next
    获取下一个id `curl 'http://127.0.0.1:7000/next' -d '{"app":"ugc","db":"topic"}'`
- /last
    最后一个id `curl 'http://127.0.0.1:7000/last' -d '{"app":"ugc","db":"topic"}'`
- /max
    最大的id `curl 'http://127.0.0.1:7000/max' -d '{"app":"ugc","db":"topic"}'`
- /remainder
    剩余数量 `curl 'http://127.0.0.1:7000/remainder' -d '{"app":"ugc","db":"topic"}'`

文档
---
- [需要调研](doc/research.md)
- [segment设计](doc/segment.md)
- [store维护](doc/tree.md)

功能
---
- [x] 基于redis发号DEMO
- [x] 实现数字生成器
- [x] 使用etcd 作为发号的存储
- [x] 完善预取逻辑
- [x] 提供预览版本做性能和可用性测试
- [x] 使用zk作为发号的存储
- [x] redis协议支持
- [x] 使用锁替换channel 减少内存占用
- [x] toml管理配置
- [x] 可用性测试，zk/etcd 短时间故障，比如超时或者选主
- [x] 将/rabbitid/[dc]/[db] 增加层级/rabbitid/[dc]/[db]/[table]


感谢
---
- andrew-d_id
- bear
- dhetis
- distributed-unique-id
- flike_idgo
- fyllo
- get_uid
- go-id-alloc
- go-id-builder
- go-katsubushi
- go-unique
- goSnowFlake
- goimpulse
- hikenote_idgenerator
- id
- idCreator
- idGenerator
- id_publisher
- idalloc
- idgen
- idgo
- idleaf
- ids
- igener
- indigo
- kala
- libxx_id
- mazhaoyong_idgenerator
- numerical-id-generator
- redis-id-generator
- shopexguid
- the-anna-project_id
- tokenserver
- upid
- zkUid
