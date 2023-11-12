# Carrota 2.0 Plugin Center API 文档

- ligen131 \<[i@ligen131.com](mailto:i@ligen131.com)\>

## 总览

### 目录

  + 1 [总览](#总览)
    + 1.1 [目录](#目录)
    + 1.2 [约定](#约定)
  + 2 [Health](#health)
    + 2.1 [[GET] `/health`](#get-health)
  + 3 [插件 Plugin](#插件-plugin)
    + 3.1 [[POST] `/plugin/register`](#post-pluginregister)
    + 3.2 [[POST] 插件端接口](#post-插件端接口)
    + 3.3 [[GET] `/plugin/list`](#get-pluginlist)
  + 4 [消息 Message](#消息-message)
    + 4.1 [[POST] `/message/send`](#post-messagesend)

### 约定

- **API 请求链接：<https://plugin-center.carrot.cool/api/v1>**
- **所有需要传递参数的 GET 请求都使用 QueryString 格式或 URL 而非 JSON Body。**

## Health

### [GET] `/health`

获取服务状态。

#### Request

无。

#### Response

```json
{
  "code": 200,
  "msg": null,
  "data": "ok",
}
```

## 插件 Plugin

### [POST] `/plugin/register`

提交/更新插件注册信息。推荐插件定时调用该接口以及时更新插件信息。

#### Request

```json
{
  "id": "homework_notify",
  "name": "作业提醒",
  "author": "ligen131",
  "description": "作业提醒系统，同学们可以通过机器人查询指定时间范围内的作业，老师或学习委员可以通过机器人添加作业内容和截止时间。该系统还会定时提醒当天截止的作业。",
  "prompt": "需要与查询作业相关的所有消息，不一定是疑问句。",
  "param": [
    {
      "key": "date",
      "type": "integer",
      "description": "提取日期或时间，格式为时间戳整数形式，以秒为单位，今天是 2023 年 11 月 13 日。"
    },
    {
      "key": "subject",
      "type": "string",
      "description": "提取科目名称"
    }
  ],
  "format": [
    "${date}的${subject}作业是什么？",
    "${date}有什么作业？",
    "${subject}作业什么时候截止？"
  ],
  "example": [
    "3 月 2 日的语文作业是什么？",
    "今天有什么作业要截止？"
  ],
  "url": "https://homework.carrot.cool/api/v1/message"
}
```

| 字段                  | 类型       | 可选                  | 描述                                                                                                             |
| --------------------- | ---------- | --------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `id`                  | `string`   | 必需                  | 插件唯一标识符，用于与其他插件区分开，推荐使用随机字符串填充。若后来者 `id` 与之前的冲突，会更新该插件所有信息。 |
| `name`                | `string`   | 必需                  | 插件名称。                                                                                                       |
| `author`              | `string`   | 必需                  | 插件作者。                                                                                                       |
| `description`         | `string`   | 必需                  | 插件功能描述。                                                                                                   |
| `prompt`              | `string`   | 必需                  | 用于告诉大模型 Parser 何时需要触发该插件并解析用户消息的描述文字。                                               |
| `param`               | `Object[]` | 可选                  | 参数数组，用于告诉大模型需要将消息解析返回哪些参数。                                                             |
| `param[].key`         | `string`   | 每一个 `param` 中必需 | 参数标识符，用于 `format` 字段中和大模型的返回参数。                                                             |
| `param[].type`        | `string`   | 每一个 `param` 中必需 | 参数类型，常见类型有 `interger, string, boolean`                                                                 |
| `param[].description` | `string`   | 每一个 `param` 中必需 | 参数描述，用于告诉大模型如何提取这部分参数。                                                                     |
| `format`              | `string`   | 可选                  | 可能出现的语句格式。                                                                                             |
| `example`             | `string`   | 可选                  | 触发该插件的语句举例。                                                                                           |
| `url`                 | `string`   | 必需                  | 从 Parser 接收到消息并触发该插件时，将信息上报给插件的 API 链接。                                                |

#### Response

```json
{
  "code": 200,
  "msg": null,
  "data": "ok"
}
```

### [POST] 插件端接口

每当接收到 Parser 上报的信息时，会 `POST` 字段 `url` 中的链接。

**若 Plugin Center 上报失败连续 3 次，则默认该插件已停止，以后不再上报消息。若插件重启，请调用 `/plugin/register` 接口再次注册。**

#### Request

Plugin Center 会以如下格式提交消息信息。

```json
{
  "agent": "feishu",
  "group_id": "926170830",
  "group_name": "软工交流群",
  "user_id": "1353055672",
  "user_name": "ligen131",
  "time": 1699806329,
  "message": "3 月 2 日的语文作业是什么？",
  "param": {
    "date": 1677686400,
    "subject": "语文"
  }
}
```

| 字段        | 类型      | 描述                                                                                    |
| ----------- | --------- | --------------------------------------------------------------------------------------- |
| `agent`     | `string`  | 即时通讯软件名称，通过 Carrota Agent 获取，如 `"feishu", "qq", "wechat", "telegram"` 等 |
| `group_id`  | `string`  | 群聊唯一标识符。若对话为私信，则该值为空字符串。                                        |
| `user_id`   | `string`  | 用户唯一标识符。                                                                        |
| `user_name` | `string`  | 用户名。                                                                                |
| `time`      | `integer` | 消息原始发送时间。                                                                      |
| `message`   | `string`  | 原始消息内容。                                                                          |
| `param`     | `object`  | 大模型解析出的参数结构体。                                                              |

#### Response

Plugin Center 需要得到以下格式的回复，无论是否需要发送消息。

```json
{
  "is_reply": true,
  "message": "语文作文 - 3 月 2 日 18:00 截止提交 - 学习通"
}
```

| 字段       | 类型      | 可选 | 描述                   |
| ---------- | --------- | ---- | ---------------------- |
| `is_reply` | `boolean` | 必需 | 是否直接原路回复消息。 |
| `message`  | `string`  | 可选 | 回复的消息。           |

### [GET] `/plugin/list`

获取已注册插件列表，可用于 Parser 获取大模型 prompt，也可用于插件监测是否注册成功。

#### Request

无

#### Response

```json
{
  "code": 200,
  "msg": null,
  "data": [
    {
      "id": "homework_notify",
      "name": "作业提醒",
      "author": "ligen131",
      "description": "作业提醒系统，同学们可以通过机器人查询指定时间范围内的作业，老师或学习委员可以通过机器人添加作业内容和截止时间。该系统还会定时提醒当天截止的作业。",
      "prompt": "需要与查询作业相关的所有消息，不一定是疑问句。",
      "param": [
        {
          "key": "date",
          "type": "integer",
          "description": "提取日期或时间，格式为时间戳整数形式，以秒为单位，今天是 2023 年 11 月 13 日。"
        },
        {
          "key": "subject",
          "type": "string",
          "description": "提取科目名称"
        }
      ],
      "format": [
        "${date}的${subject}作业是什么？",
        "${date}有什么作业？",
        "${subject}作业什么时候截止？"
      ],
      "example": [
        "3 月 2 日的语文作业是什么？",
        "今天有什么作业要截止？"
      ],
      "url": "https://homework.carrot.cool/api/v1/message"
    }
  ]
}
```

字段含义与 `/plugin/register` 中的请求参数相同。

## 消息 Message

### [POST] `/message`

Agent 将接收到的消息上报给 Plugin Center，并得到最终的回复信息。

#### Request

```json
{
  "agent": "feishu",
  "group_id": "926170830",
  "group_name": "软工交流群",
  "user_id": "1353055672",
  "user_name": "ligen131",
  "time": 1699806329,
  "message": "3 月 2 日的语文作业是什么？"
}
```

#### Response

```json
{
  "is_reply": true,
  "message": [
    "今天 18:00 需要在学习通上提交语文作业哦！别忘了！"
  ]
}
```

| 字段       | 类型       | 描述                                                        |
| ---------- | ---------- | ----------------------------------------------------------- |
| `is_reply` | `boolean`  | 是否直接原路回复消息，若为 `false`，请忽略 `message` 字段。 |
| `message`  | `string[]` | 回复的消息数组，由于可能触发多个插件，故该值可能不止一个。  |

### [POST] Carrota Parser 端接口

调用该接口将原始消息解析为触发哪些插件和插件参数信息，处理前 Parser 可能需要调用 `/plugin/list` 接口获取已注册插件信息。

#### Request

Plugin Center 会以如下格式发送请求。

```json
{
  "agent": "feishu",
  "group_id": "926170830",
  "group_name": "软工交流群",
  "user_id": "1353055672",
  "user_name": "ligen131",
  "time": 1699806329,
  "message": "3 月 2 日的语文作业是什么？"
}
```

#### Response

Plugin Center 需要得到以下格式回复。

```json
{
  "plugin": [
    {
      "id": "homework",
      "param": {
        "date": 1677686400,
        "subject": "语文"
      }
    }
  ]
}
```

### [POST] `/message/send`

插件直接发送消息到对应群聊/私信，用于无需引用消息的回复或定时消息。

#### Request

```json
{
  "agent": "feishu",
  "is_private": false,
  "to": "926170830",
  "message": "3 月 2 日记得在学习通提交语文作文哦。"
}
```

#### Response

```json
{
  "code": 200,
  "msg": null,
  "data": "ok"
}
```
