# 运行与排查

## 启动日志

正常启动时会输出：

```text
starting Bendan onebot_ws_url="ws://127.0.0.1:3001" action_lexicon_path="actions.json" zh=46 latin=21
onebot connected endpoint="ws://127.0.0.1:3001"
automatic reply event storage enabled
```

第一行说明 Bendan 已读取配置和动作词表；第二行表示已连接 NapCat 的 OneBot WebSocket。配置 PostgreSQL 后还会看到事件存储启用日志（各行实际顺序可能略有不同）。

Docker 环境中查看日志：

```sh
docker compose logs -f bendan
```

## OneBot 连接失败

日志若持续出现：

```text
onebot connect failed endpoint="..." retry_in=... error=...
```

检查：

1. NapCat 中 OneBot v11 正向 WebSocket 是否已启用。
2. `ONEBOT_WS_URL` 的主机、端口和协议是否正确。
3. Access Token 是否与 NapCat 中的设置一致。
4. Docker 场景下，容器是否能访问宿主机的 OneBot 服务。

Bendan 会以递增退避间隔持续重连，NapCat 恢复后会自动重新建立连接。

## 动作词表没有生效

先从管理员 QQ 发送：

```text
//actions
//actions list
```

然后检查：

- `ACTION_LEXICON_PATH` 指向的文件是否正确；
- JSON 是否有效；
- `zh` 与 `latin` 至少有一组包含动作；
- 动作名和输出文本是否都非空；
- Docker 挂载的 `actions.json` 是否指向实际文件。

保存正确内容后最多等待一秒。日志会出现：

```text
action lexicon reloaded path="..." zh=46 latin=21
```

无效内容会产生 `action lexicon reload failed`，但旧词表仍会保留。

## 指令没有反应

- `//whoami` 可验证机器人是否实际收到了消息。
- `//status` 可快速确认 OneBot 连接、词表数量、运行时间和当前会话静默状态。
- 代码执行、`//actions`、`//status` 与 `//hush status` 只允许配置的管理员帐号。
- 回复动作时，确认消息确实是“回复”而不是只复制了文本。
- 机器人自己发出的消息会被忽略，避免自我循环。

## 回复 Bot 的静默指令没有生效

静默控制要求消息回复 Bendan，或明确 `@Bendan`，并在正文中包含静默关键词。NapCat 的引用关系来自 OneBot `reply` 消息段，Bendan 会调用 `get_msg` 查询被引用消息的发送者。

如果没有生效，请检查：

1. NapCat 是否允许并正确响应 OneBot `get_msg`；
2. Bendan 日志是否出现 `onebot resolve reply failed`；
3. 消息是否真正回复了 Bendan，而不是只复制文字；
4. `@` 的 QQ 是否为当前 Bot 自身账号。

为避免误伤，回复其他群成员后说“闭嘴”，以及没有回复或 `@Bendan` 的普通聊天，都不会触发整个会话静默。

## 自动回应事件没有写入数据库

配置 `BENDAN_DATABASE_URL` 后，启动日志应包含：

```text
automatic reply event storage enabled
```

成功写入时会出现 `automatic reply events stored count=...`。如果数据库不可用，Bot 仍会正常回复，只在日志中记录批次写入失败；队列已满时会记录 `reason=queue_full`。使用 `network_mode: host` 时，同机 PostgreSQL 通常应连接 `127.0.0.1`。

## 重复回复或异常文本

Bendan 会按照会话和消息 ID 去重，并对自动回应设置短暂冷却。出现异常时，请保留以下内容以便定位：

- 触发消息的原文；
- NapCat 的 OneBot 事件日志；
- Bendan 容器的对应时间段日志；
- 是否包含回复、`@all`、多重提及、图片或转发消息段。
