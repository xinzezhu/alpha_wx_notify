# alpha_wx_notify
币安alpha互动微信通知

# 功能
爬取空投日历，把今日的空投信息通过调用方糖的接口推送到微信。
只要今日空投有任何变化，就会发通知给你：
![image](https://github.com/user-attachments/assets/a4c55f0d-633b-438a-a709-c25a82599e36)
![image](https://github.com/user-attachments/assets/6641ab15-eeb6-4f79-a7eb-1f6e713bc60c)


# 使用方法
在config.json中填入你的sendKey，有多个就填入多个。（sendKey获取方法：用微信打开https://sct.ftqq.com/sendkey）

填完之后，直接启动
linux: nohup ./alpha_wx_notify &

window：双击打开

config.json的配置：
{
    "sendkeys": [""], #sendkey
    "interval": 5, # 间隔多少分钟检测一次
    "fiterTge": true # 是否过滤tge活动
}


# 编译
go build
