# Canal ES 同步服务

## 架构说明

```
MySQL → Canal Server → Kafka → Canal Consumer → Elasticsearch
```

## 文件结构

```
service/canal/
├── canal.go                    # 主程序入口
├── etc/
│   └── canal.yaml             # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go          # 配置结构定义
│   ├── logic/
│   │   ├── kafka_consumer.go  # Kafka消费者
│   │   └── sync_handler.go    # ES同步处理器
│   ├── model/
│   │   ├── canal_message.go   # Canal消息结构
│   │   └── es_documents.go    # ES文档结构
│   └── svc/
│       └── servicecontext.go  # 服务上下文
└── README.md
```

## Canal Server 配置

### 1. 下载 Canal Server

```bash
# 下载 Canal Server
wget https://github.com/alibaba/canal/releases/download/canal-1.1.7/canal.deployer-1.1.7.tar.gz
tar -zxvf canal.deployer-1.1.7.tar.gz
cd canal.deployer-1.1.7
```

### 2. 修改 conf/canal.properties

```properties
# canal server port
canal.port = 11111

# zk配置
canal.zkServers = 127.0.0.1:2181

# 可配置的destination列表，逗号分隔
canal.destinations = example

# 配置MQ模式
canal.serverMode = kafka

# Kafka配置
canal.mq.servers = 127.0.0.1:9092
canal.mq.topic = canal_topic
canal.mq.partition = 0
canal.mq.partitionsNum = 1
```

### 3. 修改 conf/example/instance.properties

```properties
# MySQL地址
canal.instance.master.address = 127.0.0.1:3306

# MySQL 用户名/密码
canal.instance.dbUsername = root
canal.instance.dbPassword = your_password

# 需要监听的数据库
canal.instance.defaultDatabaseName = schill

# 监听的表正则表达式 (这里监听所有表)
canal.instance.filter.regex = schill\\..*

# 编码
canal.instance.connectionCharset = UTF-8

# MySQL binlog解析器
canal.instance.tsdb.spring.xml = classpath:spring/tsdb/h2-tsdb.xml
```

### 4. MySQL Binlog 配置

确保MySQL开启了binlog，在my.cnf中配置：

```ini
[mysqld]
log-bin = mysql-bin
binlog-format = ROW
server-id = 1
```

## Elasticsearch 索引创建

在启动服务前，先创建ES索引：

```bash
# 创建 user 索引
curl -X PUT "http://localhost:9200/user" -H "Content-Type: application/json" -d @es_index/user_index.json

# 创建 post 索引
curl -X PUT "http://localhost:9200/post" -H "Content-Type: application/json" -d @es_index/post_index.json

# 创建 topic 索引
curl -X PUT "http://localhost:9200/topic" -H "Content-Type: application/json" -d @es_index/topic_index.json

# 创建 follow 索引 (可选)
curl -X PUT "http://localhost:9200/follow" -H "Content-Type: application/json" -d @es_index/follow_index.json
```

## 启动服务

### 1. 启动 Canal Server

```bash
cd canal.deployer-1.1.7
sh bin/startup.sh
```

### 2. 启动 Kafka

```bash
# 启动Zookeeper
bin/zookeeper-server-start.sh config/zookeeper.properties

# 启动Kafka
bin/kafka-server-start.sh config/server.properties
```

### 3. 启动 Canal ES 同步服务

```bash
cd service/canal
go run canal.go -f etc/canal.yaml
```

或者编译后运行：

```bash
go build -o canal canal.go
./canal -f etc/canal.yaml
```

## 配置说明

### canal.yaml

```yaml
Name: canal.rpc
ListenOn: 0.0.0.0:8088

Kafka:
  Brokers:
    - 127.0.0.1:9092
  Topic: canal_topic
  Group: canal_es_consumer

Elasticsearch:
  Hosts:
    - http://127.0.0.1:9200
  Username: ""
  Password: ""

Log:
  Mode: file
  Path: ./logs
  Level: info
```

## 已支持同步的表

| 表名 | 说明 | ES索引 |
|------|------|--------|
| user | 用户表 | user |
| user_stat | 用户统计表 | user (更新) |
| post | 动态表 | post |
| post_content | 动态内容表 | post (更新) |
| topic | 话题表 | topic |

## 消息处理流程

1. Canal Server监听MySQL binlog
2. 将变更消息发送到Kafka的`canal_topic`主题
3. Canal Consumer从Kafka消费消息
4. 根据表名分发到对应的同步处理器
5. 将数据同步到Elasticsearch

## 数据类型转换

Canal消息中的数据类型转换：
- 数字类型: float64 → uint64/int64/int32/int8
- 日期类型: RFC3339 或 "2006-01-02 15:04:05" → UnixMilli
- 布尔类型: 1/true → true, 0/false → false

## 注意事项

1. **软删除处理**: 当检测到`deleted_at`字段有值时，会从ES中删除对应文档
2. **数据一致性**: 使用ES的`refresh=true`确保立即可见，生产环境可调整
3. **幂等性**: 使用ID作为ES文档ID，确保重复消息不产生问题
4. **错误处理**: 消费失败会记录日志并继续消费下一条消息
5. **性能优化**: 
   - 可批量提交ES请求
   - 调整refresh_interval
   - 使用ES批量API
