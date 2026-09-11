# Bendan Bot

基于 Go、QQNT、NapCatQQ 和 OneBot v11 的 QQ 机器人。

## 功能

### 动作

动作词来自内置白名单，可用中文或常见英文动作名；英文动作会转换为中文表达。动作可带 `/`，也可直接发送“动作 + 目标”。

| 输入方式 | 示例 | 效果 |
| --- | --- | --- |
| 回复消息后发送动作 | 回复某人的消息后发送 `摸` | `发送者 摸了 对方！` |
| 回复消息后发送动作和部位 | 回复 Bendan 后发送 `摸 头` | `发送者 摸了 Bendan的头！` |
| 提及目标后发送动作 | `摸 @Bendan` | `发送者 摸了 Bendan！` |
| 直接指定目标 | `抱 对方` | `发送者 抱了 对方！` |
| 使用英文别名 | `rua 对方` | `发送者 揉了揉 对方！` |

常用动作包括：摸、摸摸、抱、抱抱、拍、拍拍、戳、戳戳、亲、亲亲、揉、揉揉、捏、捏捏、蹭、蹭蹭、贴贴、啵、啵啵、抓、挠、挠挠、打、踢、咬、咬咬、舔、舔舔、夸、夸夸、举、举高高；英文别名包括 `rua`、`hug`、`kiss`、`pat`、`poke`、`boop`、`bonk`、`slap`、`bite`、`lick`、`fistbump`、`highfive`。

动作词表由启动目录下的 `actions.json` 驱动，未提供该文件时使用内置默认词表。可复制 `actions.json.example` 为 `actions.json` 后编辑，无需重新编译；修改后最多一秒自动重载，语法错误时会保留上一份有效词表并记录日志。`zh` 是中文动作到输出动词的映射，`latin` 是不区分大小写的英文别名到输出动词的映射。

```json
{
  "zh": { "拍": "拍了", "拍拍": "拍了拍" },
  "latin": { "pat": "拍了拍", "wave": "挥了挥" }
}
```

### 其他指令与自动回复

- 发送 `/me 内容`。如 `/me 喝醉了`，机器人发送 `A 喝醉了！`
- 发送 `//whoami`，查询当前 QQ 号和会话 ID
- 发送 `/没关系` 或 `/没事的`，获得一条鼓励回复
- 发送 `？`，机器人回复一个问号
- 发送 `看看…`、`是…吗`、`有没有…`、`能不能…` 等句式，机器人半随机回应
- 发送 `//go` 或 `//js` 后跟代码，执行 Go 或 JavaScript（仅管理员可用）
- 管理员可发送 `//actions` 查看词表状态、`//actions list` 查看动作、`//actions reload` 立即重载配置、`//actions add <zh|latin> <动作> <输出>` 添加动作，以及 `//actions remove <zh|latin> <动作>` 删除动作
- 回复机器人说“别说话”“闭嘴”或“安静”会使它在该会话中静默 30 分钟；回复“说话”恢复
- 链接净化实现仍保留，但已禁用自动回复，避免在群聊中打断对话

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
  "onebot_access_token": "replace-with-a-long-random-token",
  "administrator_qq": "replace-with-your-qq-number",
  "action_lexicon_path": "actions.json"
}
```

| 配置项 | 必填 | 说明 |
| --- | --- | --- |
| `ONEBOT_WS_URL` / `onebot_ws_url` | 是 | NapCat OneBot v11 正向 WebSocket 地址，例如 `ws://127.0.0.1:3001` |
| `ONEBOT_ACCESS_TOKEN` / `onebot_access_token` | 建议 | 与 NapCat WebSocket Access Token 一致；生产环境必须设置 |
| `ADMINISTRATOR_QQ` / `administrator_qq` | 建议 | 唯一管理员的 QQ 号；未配置时管理员与代码执行功能均禁用 |
| `ACTION_LEXICON_PATH` / `action_lexicon_path` | 否 | 动作词表 JSON 路径，默认 `actions.json` |

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
