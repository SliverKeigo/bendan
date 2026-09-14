# Bendan Bot

基于 Go、QQNT、NapCatQQ 和 OneBot v11 的 QQ 机器人。

## 功能

### 动作

动作词来自内置白名单，可用中文或常见英文动作名；英文动作会转换为中文表达。动作可带 `/`，也可直接发送“动作 + 目标”。

| 输入方式 | 示例 | 效果 |
| --- | --- | --- |
| 回复消息后发送动作 | 回复某人的消息后发送 `摸` | `发送者 摸了摸对方！` |
| 回复消息后发送动作和部位 | 回复 Bendan 后发送 `摸 头` | `发送者 摸了摸Bendan的头！` |
| 提及目标后发送动作 | `摸 @Bendan` | `发送者 摸了摸Bendan！` |
| 直接指定目标 | `抱 对方` | `发送者 抱了抱对方！` |
| 使用自然语序动作 | `比心 对方` | `发送者 给对方比了个心！` |
| 使用英文别名 | `highfive 对方` | `发送者 和对方击了个掌！` |

常用中文动作包括摸、抱、拍、戳、亲、揉、捏、蹭、贴、比心、击掌、握手、举高高等；英文别名包括 `rua`、`hug`、`pet`、`headpat`、`cuddle`、`wave`、`feed`、`fistbump`、`highfive`、`handshake` 等。完整清单以 `actions.json.example` 为准。

动作词表由启动目录下的 `actions.json` 驱动，未提供该文件时使用内置默认词表。可复制 `actions.json.example` 为 `actions.json` 后编辑，无需重新编译；修改后最多一秒自动重载，语法错误时会保留上一份有效词表并记录日志。`zh` 是中文动作到输出文本的映射，`latin` 是不区分大小写的英文别名到输出文本的映射。内置动作都使用 `{target}` 明确目标位置；不含该占位符的旧版自定义动作仍按“发送者 动作 目标”格式输出。

```json
{
  "zh": {
    "拍": "拍了拍",
    "比心": "给{target}比了个心"
  },
  "latin": {
    "pat": "拍了拍",
    "highfive": "和{target}击了个掌"
  }
}
```

### 其他指令与自动回复

- 发送 `/me 内容`。如 `/me 喝醉了`，机器人发送 `A 喝醉了！`
- 发送 `//whoami`，查询当前 QQ 号和会话 ID
- 发送 `/没关系` 或 `/没事的`，获得一条鼓励回复
- 发送 `？`，机器人直接发送一个问号（自动回应默认不引用原消息）
- 发送 `看看…`、`是…吗`、`是不是…`、`有没有…`、`能不能…`、`会不会…`、`可不可以…`、`行不行…`、`好不好…`、`要不要…`、`该不该…`、`值不值得…` 等句式，机器人半随机回应
- `A 还是 B` 选择问句目前不会触发自动回应
- 发送 `//go` 或 `//js` 后跟代码，执行 Go 或 JavaScript（仅管理员可用）
- 管理员可发送 `//actions` 查看词表状态、`//actions list` 查看动作、`//actions reload` 立即重载配置、`//actions add <zh|latin> <动作> <输出>` 添加动作，以及 `//actions remove <zh|latin> <动作>` 删除动作
- 回复机器人，或直接 `@Bendan`，并在消息中包含“闭嘴”“别说话”“不要说话”“别讲话”“不要讲话”“住嘴”“安静”“别吵”或“消停”等关键词，会使它在该会话中静默 30 分钟；同样回复或 `@Bendan` 并包含“说话”可提前恢复
- 只有明确回复或提及机器人时才会触发静默，群成员之间正常使用这些词不会影响机器人
- 链接净化实现仍保留，但已禁用自动回复，避免在群聊中打断对话

## 架构

```text
QQNT + NapCatQQ
      | OneBot v11 正向 WebSocket
      v
Bendan Bot (Go)
```

本服务使用 OneBot v11 **正向 WebSocket**：Bendan 主动连接 NapCat 提供的 WebSocket 服务端。NapCat 的 OneBot 配置需要启用正向 WebSocket，并允许本服务连接。引用消息中的 `reply` 消息段会通过 OneBot `get_msg` 补全，用于判断静默指令是否确实指向机器人。

成功发送的自动回应可以异步写入 PostgreSQL，用于分析处理器触发频率和优化自然语言规则。数据库写入采用有界非阻塞队列；数据库不可用或队列已满时只记录日志，不影响机器人继续回复。事件会保存输入、输出、处理器和投递方式，QQ 用户与群标识仅保存带盐 SHA-256 哈希。

## 配置

服务会优先读取环境变量；也可在项目根目录创建未纳入 Git 的 `.config`：

```json
{
  "onebot_ws_url": "ws://127.0.0.1:3001",
  "onebot_access_token": "replace-with-a-long-random-token",
  "administrator_qq": "replace-with-your-qq-number",
  "automatic_reply_delay": "false",
  "action_lexicon_path": "actions.json"
}
```

| 配置项 | 必填 | 说明 |
| --- | --- | --- |
| `ONEBOT_WS_URL` / `onebot_ws_url` | 是 | NapCat OneBot v11 正向 WebSocket 地址，例如 `ws://127.0.0.1:3001` |
| `ONEBOT_ACCESS_TOKEN` / `onebot_access_token` | 建议 | 与 NapCat WebSocket Access Token 一致；生产环境必须设置 |
| `ADMINISTRATOR_QQ` / `administrator_qq` | 建议 | 唯一管理员的 QQ 号；未配置时管理员与代码执行功能均禁用 |
| `AUTOMATIC_REPLY_DELAY` / `automatic_reply_delay` | 否 | 是否为自动回应增加 500ms～1200ms 随机延迟，默认 `false` |
| `ACTION_LEXICON_PATH` / `action_lexicon_path` | 否 | 动作词表 JSON 路径，默认 `actions.json` |
| `BENDAN_DATABASE_URL` | 否 | PostgreSQL 连接地址；配置后异步记录成功发送的自动回应事件 |
| `BENDAN_EVENT_HASH_SALT` | 建议随数据库配置 | 用户及会话标识的不可逆哈希盐；应使用独立的长随机值 |

## 本地运行

前提：QQNT 和 NapCatQQ 已在同一台主机上运行并登录目标 QQ 账号，且 OneBot v11 的正向 WebSocket 服务已开启。

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
npm run docs:build
```

以上命令分别运行 Go 测试套件并构建文档站。

## 风险

NapCatQQ/QQNT 是个人 QQ 自动化方案，不是 QQ 开放平台机器人。QQ 客户端升级、登录校验和风控均可能影响可用性；请使用专门的机器人账号，避免将主账号用于自动化。

## License

Bendan Bot is licensed under the [MIT License](LICENSE).
