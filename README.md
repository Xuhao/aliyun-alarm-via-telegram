# aliyun-alarm-via-telegram

使用电报(Telegram)接收阿里云报警信息

## 功能

* 阿里云 k8s 事件中心报警
* ECS、负载均衡器健康检查等其它通用报警

## 使用

```bash
docker run -d \
           -p 80:8080 \
           --restart=always \
           --name aliyun-alarm-via-telegram \
           xuhao/aliyun-alarm-via-telegram:latest
```

### 1. 阿里云 k8s 事件中心报警

webhook 地址设置为: `http://<host>/bot<bot-token>/sendMessage`

发送信息示例：

```
parse_mode=<HTML或MarkdownV2>&disable_web_page_preview=true&chat_id=<群ID>&text=<自定义消息内容>
```

### 2. 通用报警

将回调地址设置为: `http://<host>/bot<bot-token>/sendMessage?chat_id=<群ID>&text=assembly`

text固定为`assembly`，程序会自动将阿里云回调时附带的body信息格式化并覆盖text推送。