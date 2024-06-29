# 监控摄像头和麦克风使用状态

## 需手动加载配置文件

```json
{
  "sendEmail": false,
  "emailConfig": {
        "from": "your-email@example.com",
        "password": "your-email-password",
        "to": "recipient@example.com",
        "smtpHost": "smtp.example.com",
        "smtpPort": "587"
    },
    "checkInterval": 5
}
```
---
## 使用命令
```
go run main.go -config=config.json
```