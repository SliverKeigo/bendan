---
layout: home

hero:
  name: Bendan
  text: QQ 群里的轻量互动机器人
  tagline: 由 Go、NapCatQQ 与 OneBot v11 驱动。动作、自动回应和受控运维功能集中在一处，配置清晰，运行可观察。
  actions:
    - theme: brand
      text: 开始使用
      link: /guide/getting-started
    - theme: alt
      text: 查看功能
      link: /guide/features

features:
  - icon: "✦"
    title: 自然的群聊动作
    details: 回复、提及或直接指定对象即可触发动作，中文动作与英文别名都可用。
  - icon: "↻"
    title: 无重启改词表
    details: 编辑 actions.json 后最多一秒自动加载；无效配置不会覆盖正在使用的词表。
  - icon: "⌁"
    title: 管理范围明确
    details: 词表管理与代码执行只向指定 QQ 管理员开放，普通成员不会获得执行入口。
---

## 从一条消息开始

Bendan 面向已经运行 QQNT 与 NapCatQQ 的环境。完成 OneBot 正向 WebSocket 配置后，机器人主动连接 NapCat，处理群聊与私聊消息。

```text
回复某人的消息后：摸

发送者 摸了摸对方！
```

[查看完整上手流程](/guide/getting-started)
