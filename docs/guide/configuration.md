# 配置与部署

Bendan 的配置优先级为：**环境变量 > 项目根目录 `.config` > 默认值**。

## 环境变量

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `ONEBOT_WS_URL` | 是 | NapCat OneBot v11 正向 WebSocket 地址，例如 `ws://127.0.0.1:3001`。 |
| `ONEBOT_ACCESS_TOKEN` | 建议 | 与 NapCat WebSocket Access Token 一致；生产环境应设置。 |
| `ADMINISTRATOR_QQ` | 建议 | 唯一管理员的 QQ 号；未设置时受限管理能力保持禁用。 |
| `AUTOMATIC_REPLY_DELAY` | 否 | 是否为自动回应增加 500ms～1200ms 随机延迟；默认 `false`。 |
| `ACTION_LEXICON_PATH` | 否 | 动作词表 JSON 文件路径；默认为 `actions.json`。 |
| `BENDAN_DATABASE_URL` | 否 | PostgreSQL 连接地址；配置后异步记录成功发送的自动回应事件。 |
| `BENDAN_EVENT_HASH_SALT` | 建议随数据库配置 | 用户及会话标识的不可逆哈希盐；使用独立的长随机值。 |

`.env.example`：

```dotenv
ONEBOT_WS_URL=ws://host.docker.internal:3001
ONEBOT_ACCESS_TOKEN=replace-with-a-long-random-token
ADMINISTRATOR_QQ=replace-with-your-qq-number
# 可选：为自动回应增加 500ms～1200ms 随机延迟，默认关闭
AUTOMATIC_REPLY_DELAY=false
# 可选：将成功发送的自动回应异步写入 PostgreSQL
BENDAN_DATABASE_URL=postgres://bot:password@host.docker.internal:5432/bendan?sslmode=disable
# 启用事件存储时建议设置独立的长随机哈希盐
BENDAN_EVENT_HASH_SALT=replace-with-a-long-random-value
ACTION_LEXICON_PATH=actions.json
```

`.env` 和 `.config` 已在 `.gitignore` 中，避免将 Token 提交到仓库。

## `.config` 文件

不使用环境变量时，可以在工作目录创建 `.config`：

```json
{
  "onebot_ws_url": "ws://127.0.0.1:3001",
  "onebot_access_token": "replace-with-a-long-random-token",
  "administrator_qq": "replace-with-your-qq-number",
  "automatic_reply_delay": "false",
  "action_lexicon_path": "actions.json"
}
```

键名与环境变量对应，但使用小写形式。

## 自动回应事件存储

配置 `BENDAN_DATABASE_URL` 后，Bendan 会自动创建事件表及时间、处理器索引，并异步记录成功发送的自动回应。记录内容包括处理器、会话类型、输入、输出、投递方式和时间。

用户和会话标识不会以真实 QQ 号或群号保存，而是使用 `BENDAN_EVENT_HASH_SALT` 进行带盐 SHA-256 哈希。数据库写入采用有界非阻塞队列；连接或写入失败只记录日志，不会阻塞或禁用 Bot 回复。

Docker 使用默认桥接网络时，数据库位于宿主机可使用 `host.docker.internal`。如果生产 Compose 使用 `network_mode: host`，应使用宿主机实际监听地址（同机 PostgreSQL 通常为 `127.0.0.1`），不要依赖 `host.docker.internal`。

## Docker Compose

项目内的 `docker-compose.yml` 会：

- 注入 OneBot 连接信息；
- 挂载运行时 `actions.json` 到容器；
- 让管理员词表命令能够更新该文件；
- 让热重载检测到文件变化。

启动：

```sh
docker compose up -d --build
```

查看日志：

```sh
docker compose logs -f bendan
```

## 生产注意事项

- Access Token 不要出现在 README、截图或公开日志中。
- 为机器人账号单独准备 QQ，避免使用日常主账号承载自动化。
- 升级 QQNT 或 NapCat 前先保留现有配置与运行日志。
- `//go`、`//js` 会执行代码，管理员 QQ 必须保持为可信帐号。

继续阅读：[动作词表](/guide/actions) 与 [运行排查](/guide/operations)。
