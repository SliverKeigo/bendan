# 管理员命令

Bendan 使用 `ADMINISTRATOR_QQ` 环境变量或 `.config` 中的 `administrator_qq` 配置唯一管理员。管理员验证按收到的 OneBot 发送者 QQ 号进行，不依赖群角色或昵称。

未配置管理员时，词表管理与代码执行会保持禁用状态。配置示例见 [配置与部署](/guide/configuration)。

## 权限边界

| 能力 | 普通成员 | 配置的管理员 |
| --- | --- | --- |
| 动作、`/me`、`//whoami` | 可用 | 可用 |
| 自动回应与静默控制 | 可用 | 可用 |
| `//actions` 词表运维 | 不执行 | 可用 |
| `//go`、`//js` 代码执行 | 不执行 | 可用 |

普通成员发送受限命令时，Bendan 会静默拦截，不返回执行结果。

## 动作词表管理

完整命令见 [动作词表](/guide/actions)。常见工作流：

```text
//actions add latin wave 挥了挥
//actions list
//actions remove latin wave
```

词表写入完成后会立即加载，也会被后台热重载检查检测到。

## 代码执行

Bendan 支持管理员通过消息运行 Go 或 JavaScript：

````text
//go
package main

import "fmt"

func main() {
  fmt.Println("hello")
}
````

```text
//js console.log('hello')
```

代码执行属于高风险能力。管理员 QQ 必须是你完全控制的帐号，并且不要在公开群中执行会暴露 Token、文件路径或隐私信息的代码。

## 修改管理员帐号

在部署环境的 `.env` 中设置：

```dotenv
ADMINISTRATOR_QQ=replace-with-your-qq-number
```

也可在 `.config` 中设置：

```json
{ "administrator_qq": "replace-with-your-qq-number" }
```

修改环境变量后重启 Bendan。不要将真实 QQ 号写入公开文档、示例文件或提交记录。
