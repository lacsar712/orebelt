# orebelt

露天矿 **皮带机振动采样** 服务：接收边缘采集器推送的加速度样本，在环形缓冲上运行尖峰检测状态机，经皮带级节流聚合后向维护系统上报事件，并提供嵌入式 Web 操作台。

## 功能

- **环形缓冲** (`internal/ring`)：固定容量 per-belt 采样环，写满覆盖最旧；`Snapshot()` 返回时间升序独立拷贝
- **尖峰 FSM** (`internal/spike`)：`Idle → InSpike → Idle`，连续命中进入、连续未命中或超时退出
- **节流** (`internal/throttle`)：同一皮带最短上报间隔（默认 30s），拒绝期间缓存 pending 并可刷新更高 peak
- **接入/外发** (`internal/ingest`, `internal/emit`)：`POST /v1/samples` 接入，本地存储 + 可选 HTTP webhook 外发
- **Web UI**：嵌入静态控制台，查看最近尖峰与缓冲状态

## 快速开始

```bash
cd orebelt
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go run ./cmd/orebelt -listen :8080
```

浏览器打开 http://localhost:8080/

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/samples` | 推送单条或批量样本 |
| GET | `/v1/belts/{id}/spikes` | 查询皮带最近尖峰 |
| GET | `/v1/stats` | 缓冲/节流/尖峰汇总 |
| GET | `/v1/health` | 健康检查 |
| GET | `/` | Web 操作台 |

### 样本 JSON

```json
{
  "belt_id": "3",
  "accel": 3.2,
  "ts": "2026-08-20T09:00:00Z"
}
```

批量：

```json
{
  "samples": [
    {"belt_id": "3", "accel": 3.0, "ts": "2026-08-20T09:00:00.000Z"}
  ]
}
```

## 配置

| 环境变量 | 说明 | 默认 |
|----------|------|------|
| `OREBELT_LISTEN` | 监听地址 | `:8080` |
| `OREBELT_RING_CAPACITY` | 环缓冲容量 | `4096` |
| `OREBELT_SPIKE_THRESHOLD` | 尖峰阈值 (g) | `2.5` |
| `OREBELT_THROTTLE` | 节流间隔 | `30s` |
| `OREBELT_EMIT_URL` | 维护系统 webhook | 空 |

命令行 `-help` 查看全部 flag。

## 架构

```
边缘采集器 ──POST /v1/samples──▶ ingest ──▶ ring + spike FSM
                                              │
                                              ▼
                                    throttle ──▶ emit ──▶ 维护系统
                                              │
                                              ▼
                                         Web 操作台
```

## 测试

```bash
GOTOOLCHAIN=local go test ./... -count=1
```

## 许可证

MIT
