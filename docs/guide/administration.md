# 管理员命令

Bendan 当前只有一个管理员：QQ `1226355793`。管理员验证按收到的 OneBot 发送者 QQ 号进行，不依赖群角色或昵称。

## 权限边界

| 能力 | 普通成员 | 管理员 |
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

代码执行属于高风险能力。只应将管理员 QQ 设为你完全控制的帐号，并且不要在公开群中执行会暴露 Token、文件路径或隐私信息的代码。

## 修改管理员帐号

当前管理员 QQ 在源码 `commands/admin.go` 的 `administratorQQ` 常量中定义。修改该值后重新构建并部署 Bendan。

> 这项修改故意不开放为群内命令，避免管理员权限被聊天消息改变。
