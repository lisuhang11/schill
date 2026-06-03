# SChill on Docker Desktop Kubernetes

这套清单面向 Docker Desktop 自带的单节点 Kubernetes，本地构建镜像，不推送远端仓库�?
## 设计约定

- 统一 namespace: `schill`
- 对外入口: `gateway`，容器端�?`8086`，NodePort `30086`
- MinIO 额外暴露 `30090/30091`，便于本地浏览器访问对象和控制台
- 所有业�?Pod 都使�?`imagePullPolicy: IfNotPresent`
- 业务配置通过 `Secret` 挂载文件，不依赖应用直接读取环境变量
- MySQL/Redis/Etcd/Kafka/MinIO/Elasticsearch 使用 `hostPath` 做本地持久化
- RPC 服务打开 `AutoMigrate: true`，便于本地环境直接建表；如果你要导入仓库�?`db.sql`，可以在服务稳定后手工导�?
## 前提

1. �?Docker Desktop 中启�?Kubernetes�?2. 确认当前上下文为 `docker-desktop`:

```powershell
kubectl config use-context docker-desktop
kubectl config current-context
```

## 构建镜像

默认标签�?`latest`:

```powershell
pwsh .\deploy\k8s\docker-desktop\build-images.ps1
```

如果你要�?Git 提交打标签，例如当前仓库�?SHA `eddf8f9`:

```powershell
pwsh .\deploy\k8s\docker-desktop\build-images.ps1 -Tag eddf8f9
```

这套 YAML 默认引用 `latest`。如果你改用提交标签，请同步修改�?Deployment �?`image` 字段，或�?`kustomization.yaml` 增加 `images:` 覆盖�?
## 部署

```powershell
kubectl apply -k .\deploy\k8s\docker-desktop
kubectl get pods -n schill -w
```

如果你二次修改了 `job-es-init.yaml` 后重复应用，先删掉旧 Job 再重�?`apply`:

```powershell
kubectl delete job schill-es-init -n schill --ignore-not-found
```

建议等待以下核心依赖先变�?Ready:

- `mysql`
- `redis`
- `etcd`
- `kafka`
- `elasticsearch`
- `minio`
- `canal-server`

随后观察业务组件:

- `content-rpc` / `content-api`
- `comment-rpc` / `comment-api`
- `search-rpc`
- `canal-sync`
- `gateway`

ES 初始�?Job 会自动创�?`post`、`user`、`topic` 三个索引:

```powershell
kubectl logs -n schill job/schill-es-init
```

## 验证

查看资源:

```powershell
kubectl get all -n schill
kubectl get pvc -n schill
```

查看日志:

```powershell
kubectl logs -n schill deploy/gateway
kubectl logs -n schill deploy/user-rpc
kubectl logs -n schill deploy/search-rpc
kubectl logs -n schill deploy/canal-sync
```

访问网关:

```powershell
```

查看 MinIO:

- API: `http://localhost:30090`
- Console: `http://localhost:30091`

## 清理

```powershell
kubectl delete -k .\deploy\k8s\docker-desktop
```

`hostPath` 数据不会自动删除；如需重置，请删除节点内对应目录或重建 Docker Desktop Kubernetes 集群�?
