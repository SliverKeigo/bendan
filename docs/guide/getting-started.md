# 快速开始

Bendan 是一个运行在 QQNT、NapCatQQ 与 OneBot v11 之上的 QQ 机器人。它通过 **正向 WebSocket** 主动连接 NapCat，不需要公网回调地址。

## 运行前准备

需要准备：

1. 一台已登录目标机器人 QQ 的 QQNT 主机。
2. 已安装并启用的 NapCatQQ。
3. 已安装 Go 1.24 或更高版本，或 Docker Compose。
4. 机器人 QQ 已进入目标群，并有发言权限；若要使用撤回能力，也需要撤回权限。

## 配置 NapCat

在 NapCat WebUI 中创建并启用 OneBot v11 **正向 WebSocket** 服务：

- 服务地址应与 Bendan 的 `ONEBOT_WS_URL` 一致。
- 生产环境为 WebSocket 配置 Access Token。
- Bendan 与 NapCat 运行在同一主机时，可使用 `ws://127.0.0.1:3001`。

## 本地启动

复制环境变量示例并填写 Token：

```sh
cp .env.example .env
```

```dotenv
ONEBOT_WS_URL=ws://127.0.0.1:3001
ONEBOT_ACCESS_TOKEN=replace-with-a-long-random-token
```

然后启动：

```sh
go run .
```

启动日志中出现 `onebot connected` 即表示已连接 NapCat。

## Docker Compose 启动

```sh
docker compose up -d --build
```

默认 Compose 使用 `host.docker.internal:3001` 访问宿主机上的 NapCat。Linux Docker 环境已通过 `extra_hosts` 映射该主机名。

## 验证连接

从另一个 QQ 帐号在目标群发送：

```text
//whoami
```

Bendan 会回复发送者 QQ 号与会话 ID。收到回应后，可继续测试：

```text
回复某人的消息后发送：摸
```

预期输出类似：

```text
发送者 摸了 对方！
```

下一步：[浏览全部功能](/guide/features) 或 [配置动作词表](/guide/actions)。
