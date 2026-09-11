# 配置与部署

Bendan 的配置优先级为：**环境变量 > 项目根目录 `.config` > 默认值**。

## 环境变量

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `ONEBOT_WS_URL` | 是 | NapCat OneBot v11 正向 WebSocket 地址，例如 `ws://127.0.0.1:3001`。 |
| `ONEBOT_ACCESS_TOKEN` | 建议 | 与 NapCat WebSocket Access Token 一致；生产环境应设置。 |
| `ACTION_LEXICON_PATH` | 否 | 动作词表 JSON 文件路径；默认为 `actions.json`。 |

`.env.example`：

```dotenv
ONEBOT_WS_URL=ws://host.docker.internal:3001
ONEBOT_ACCESS_TOKEN=replace-with-a-long-random-token
ACTION_LEXICON_PATH=actions.json
```

`.env` 和 `.config` 已在 `.gitignore` 中，避免将 Token 提交到仓库。

## `.config` 文件

不使用环境变量时，可以在工作目录创建 `.config`：

```json
{
  "onebot_ws_url": "ws://127.0.0.1:3001",
  "onebot_access_token": "replace-with-a-long-random-token",
  "action_lexicon_path": "actions.json"
}
```

键名与环境变量对应，但使用小写形式。

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
