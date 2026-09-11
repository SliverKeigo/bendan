# Bendan Bot

基于 Go、QQNT、NapCatQQ 和 OneBot v11 的 QQ 机器人。

## 功能

- 对消息发送 `/动作`。如 `/吃`，机器人发送 `A 吃了 B！`
- 对消息发送 `/动作 结果`。如 `/吃 豆腐`，机器人发送 `A 吃 B 豆腐！`
- 发送 `/me 内容`。如 `/me 喝醉了`，机器人发送 `A 喝醉了！`
- 发送 `//whoami`，查询当前 QQ 号和会话 ID
- 发送 `/没关系` 或 `/没事的`，获得一条鼓励回复
- 支持链接中的 Bilibili、YouTube、Twitter 等平台净化工具（当前未启用自动回复）
- 发送 `？`，机器人回复一个问号
- 发送 `看看…`、`是…吗`、`有没有…`、`能不能…` 等句式，机器人半随机回应
- 发送 `//go` 或 `//js` 后跟代码，执行 Go 或 JavaScript
- 回复机器人说“别说话”“闭嘴”或“安静”会使它在该会话中静默 30 分钟；回复“说话”恢复

以下 Telegram 专属能力已经移除：Inline Query、频道自动转发、Telegram 消息编辑、Telegram DC 查询和 Vercel Webhook 部署。

## 架构

```text
QQNT + NapCatQQ
      | OneBot v11 正向 WebSocket
      v
Bendan Bot (Go)
```

本服务使用 OneBot v11 **正向 WebSocket**：Bendan 主动连接 NapCat 提供的 WebSocket 服务端。NapCat 的 OneBot 配置需要启用正向 WebSocket，并允许本服务连接。

## 配置

服务会优先读取环境变量；也可在项目根目录创建未纳入 Git 的 `.config`：

```json
{
  "onebot_ws_url": "ws://127.0.0.1:3001",
  "onebot_access_token": "replace-with-a-long-random-token"
}
```

| 配置项 | 必填 | 说明 |
| --- | --- | --- |
| `ONEBOT_WS_URL` / `onebot_ws_url` | 是 | NapCat OneBot v11 反向 WebSocket 地址，例如 `ws://127.0.0.1:3001` |
| `ONEBOT_ACCESS_TOKEN` / `onebot_access_token` | 建议 | 与 NapCat WebSocket Access Token 一致；生产环境必须设置 |

## 本地运行

前提：QQNT 和 NapCatQQ 已在同一台主机上运行并登录目标 QQ 账号，且 OneBot v11 的反向 WebSocket 服务已开启。

```sh
go run .
```

首次部署先在 NapCat WebUI 确认：

1. OneBot v11 正向 WebSocket 服务已启用，地址与 `ONEBOT_WS_URL` 一致。
2. 配置了 Access Token，并与 Bendan 一致。
3. Bot QQ 账号在目标群中具备所需的发言和撤回权限。
4. 用另一个账号向群发送 `//whoami`，确认机器人返回 QQ 与会话 ID。

## 开发验证

```sh
go test ./...
```

`commands/eval` 的两项旧测试锁定了历史 Go 伪随机数输出；在 Go 1.25 上会因运行时随机数实现变化而失败，不影响 OneBot 接入与其余测试包。

## 风险

NapCatQQ/QQNT 是个人 QQ 自动化方案，不是 QQ 开放平台机器人。QQ 客户端升级、登录校验和风控均可能影响可用性；请使用专门的机器人账号，避免将主账号用于自动化。

## License

Bendan Bot is licensed under the [MIT License](LICENSE).
