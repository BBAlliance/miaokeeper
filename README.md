# Miaokeeper  

> Miaokeeper 是一个群成员自主管理机器人，可以在 telegram 群组中实现：群成员自主管理、入群验证、积分统计、积分抽奖等功能。使用案例：[品云](https://t.me/PinYunYes)、项目群组：[喵屋](https://t.me/MiaoGroup)。   
>
> ## 前期准备  
>
> 1.会使用 `Screen、Pm2、Systemd、Supervisor` 其中任意一直进程守护方式。  
> 2.会搭建 `Go语言编译环境` 、或会使用 `Docker的基础用法` 。  
> 3.会使用 `MySQL` 或其他数据库。  
>
> ## 如何安装  
>
> 目前支持  `直接安装` 和 `docker` 安装两种模式。    
>

### 0.在线开发

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/BBAlliance/miaokeeper)

### 1.直接安装  

> 1.自行前往[release](https://github.com/BBAlliance/miaokeeper/releases)，下载对源码，自行编译并赋予权限，或下载服务器对应架构二进制文件。  
> 2.自行安装数据库，并设置好用户、密码、数据库名。  

```bash
./miaokeeper -token 机器人Token -upstream TG官方API或反代API网址 -database '数据库用户名:数据库密码@tcp(127.0.0.1:3306)/数据库名'
```

>例如：  

```bash  
./miaokeeper -token 123456:XXXXXXXXXXXXXXXX -upstream https://api.telegram.org -database 'miaokeeper:miaokeeper@tcp(127.0.0.1:3306)/miaokeeper'
```

> 3.若无报错说明启动成功。  
> 4.启动成功根据控制台 log 点击授权链接进行首个超管授权。如果链接失效或超管账号丢失，可以通过 `-setadmin 用户ID` 添加管理。使用 `-setadmin` 参数重置管理员成功后程序会自动退出。

### 2.Docker安装  

> 1.创建miaokeeper文件夹并进去。  

```bash  
mkdir miaokeeper && chmod +x miaokeeper && cd miaokeeper/
```

> 2.下载miaokeeper的docker启动文件 `docker-compose.yml`  

```bash
wget https://raw.githubusercontent.com/BBAlliance/miaokeeper/master/docker-compose.yml

```

> 3.修改 `docker-compose.yml` 中的 `<YOUR_TOKEN>` 为你自己的机器人Token。  

> 4.使用 `docker-compose` (或 `docker compose`) 命令启动Docker容器。  
> 5.启动成功后，根据控制台 log (`docker logs miaokeeper`) 点击授权链接进行首个超管授权。
> 6.如果链接失效或超管账号丢失，你也可以使用 `docker exec -it 容器名 or ID bash` 进入容器，进入容器后通过 `./miaokeeper -setadmin 123456 -token <YOUR_TOKEN> -database root:miaokeeper2022@tcp\(mariadb:3306\)/miaokeeper`  添加管理。其中 `-setadmin` 后加上TG用户ID。若无报错，说明添加成功，重启容器即可。  
> 
> 例如：  

```bash
./miaokeeper -setadmin 123456 -token 123456:XXXXXXXXXXXXXXXX -database root:miaokeeper2022@tcp\(mariadb:3306\)/miaokeeper 

```

```bash
docker命令：
	docker ps                           # 查看正在运行的容器
	docker ps -a                        # 查看所有容器，包括已运行和未运行的
	docker logs name_or_id              # 查看容器日志
	docker restart name_or_id           # 重启容器
	docker stop name_or_id              # 停止容器
	docker start name_or_id             # 启动容器
	docker rm name_or_id -f             # 强制删除容器

docker-compose命令：
	docker-compose up                   # 前台启动容器，主要观察日志使用
	docker-compose up -d                # 后台启动容器，长期运行
	docker-compose logs --tail=500      # 截取输出最后500行日志
	docker-compose down                 # 停止并删除容器
	docker-compose restart              # 重启
	docker-compose pull                 # 更新


```

## 如何后台运行  

> 本文以 `Systemd` 为例教你如何保持机器人后台执行，请自行学习 `screen / pm2 / supervisor` 等工具。  


> 1.各系统启动服务保存文件夹如下。如需创建请根据自己系统选择。  

```bash	
Centos:systemctlDIR="/usr/lib/systemd/system/"
Debian:systemctlDIR="/etc/systemd/system/"
Ubuntu:systemctlDIR="/etc/systemd/system/"
```

> 2.自行创建启动服务文件  

```bash	
[Unit]
Description=miaokeeper Tunnel Service          #进程名称miaokeeper
After=network.target
[Service]
Type=simple
Restart=always
 
WoringDirectory=/root                          #miaokeeper文件保存路径
ExecStart=/root/miaokeeper -token 123456:XXXXXXXXXXXXXXXX  -upstream https://api.telegram.org -database 'miaokeeper:miaokeeper@tcp(127.0.0.1:3306)/miaokeeper'
[Install]
WantedBy=multi-user.target
```

> 3.常用 `Systemd命令`  

```bash	
systemctl daemon-reload                             #首次添加服务需要执行
systemctl start miaokeeper.service                  #启动miaokeeper
systemctl stop miaokeeper.service                   #停止miaokeeper
systemctl enable miaokeeper.service                 #将服务设置为每次开机启动
systemctl enable --now miaokeeper.service           #立即启动且每次重启也启动
systemctl restart miaokeeper.service                #重启miaokeeper
systemctl disable miaokeeper.service                #关闭miaokeeper开机自启
systemctl status miaokeeper.service                 #查看miaokeeper状态

```

## 关于这个机器人的使用场景  

> 配合已有机器人（鲁小迅）做到群员自主管理，进行广告封杀。群积分内部抽奖等。  

## 启动核心参数  

> 如果想熟练启动机器人，请务必看这一部分。  

```bash
-database string       #mysql或其兼容的数据库连接URL
-redis string          #redis连接地址，通过redis提升后台计时器的可用性和容错率，例如 'password@your.ip.address:port'
-ping                  #测试bot和电报服务器之间的往返时间
-token string          #电报机器人令牌
-upstream string       #电报上游api url （可选）
-verbose               #显示所有日志
-version               #显示当前版本并退出

-bind                  #启用API服务器并绑定端口
-api-token             #给API服务器增加验证，建议前置增加 https 反向代理来确保安全性
```

## 机器人常用命令参数  

> 如果想熟练使用机器人，请务必看这一部分。  

### `Super Admin`  

```
/su_export_credit      #导出积分
/su_add_group          #开启积分统计
/su_del_group          #删除当前群组统计积分
/su_add_admin          #添加全局管理
/su_del_admin          #删除全局管理员

```

### `Group Admin`  

```
/add_admin            #提权群组管理
/del_admin            #删除群组管理
/ban_forward          #封禁频道转发（回复转发或频道iD）
/unban_forward        #解禁频道转发（回复转发或频道iD）
/export_policy        #导出规则
/import_policy        #导入规则
/export_token         #导出群API Token
/set_credit           #回复或id设置积分
/add_credit           #回复或id添加积分
/check_credit         #查看某群友积分
/set_antispoiler      #是否开启剧透
/set_channel          #绑定频道（回复空则解绑频道） 要把bot扔进频道给管理
/set_locale           #设置群组的默认语言，可加一个参数，如果是 - 则为清空设置
/create_lottery       #开启抽奖  create_lottery 奖品名称 :limit=所需积分:consume=（n|y）是否扣除积分 :num=奖品数量 :participant=参与人数
/creditrank           #开榜获取积分排行榜
/redpacket            #积分红包请输入 /redpacket <总分数> <红包个数> 来发红包哦～
/lottery              #抽奖（可加A B两个参数，从A总人数中抽B人数）

```

### `用户可用命令`  

```
/mycredit      #我的积分
/version       #版本查询
/transfer      #回复一个用户来完成积分转移
/ping          #检测bot和群组响应速度
```
### Policy jSON相关解释

`CreditMapping`     是自定义群里说话的积分表现，很多双倍积分就是修改这里

`UnderAttackMode`   开启会把验证时间缩短到30喵，并且一旦踢出是永久封

`RedPacketCaptcha`  是是否启用红包验证码

`RedPacketCaptchaFailCreditBehavior`    点击不正确的验证码处理方式（如果是正的，就是补偿分，负的就是扣分）

`CustomReply`   就是关键词回复

`NameBlackListReg`  自动放逐是 

`Match`     是什么关键词可以触发这个规则

`Name`  是规则ID，如果 Name 一样，计数器就会一样，（比如多个规则共享一个 InvokeOptions计数器，这个用法太高级，暂时不用详细写。就让他们别用重复的 Name 就好）

`Limit`     是这个规则累计能被触发多少次

`CreditBehavior`    是触发了这个规则会不会对用户积分发生变动（正就是加分，负就是扣分）

`NoForceCreditBehaviorError`    是指如果 CreditBehavior 是扣用户分的，但是用户没那么多分，展示什么错误消息（空就是不展示错误消息）

`CallbackURL`   规则被触发以后调用什么回调函数（例如 https://api.xxxx.com/someapi）调用的时候会附带上群组 token (可以 /export_token 查看，里面有回调token，这样可以防止别人非法掉用回调函数，以确保这个请求一定是 miaokeeper 发出来的，不是别人 curl 或者其他方式伪造的)

`ReplyMessage`  触发规则以后回复什么消息，（所有消息都是 markdown 语法，包括 NoForceCreditBehaviorError 也是，`{{xxx}}` 可以调用内部函数 ，比如 `{{UserName}}` 表示用户名 `{{UserLink}}` 表示用户链接，用截图里的那个方式就可以表现为 at 那个用户）

`ReplyTo`   怎么发这个消息 可以为 message group private，分别表示回复这则消息，不回复直接发送，私聊发送者。默认为 message

`ReplyButtons`  是是否添加按钮，为数组，一个一行，编写方式为：`按钮文字|按钮命令 ` 多个个一行，编写方式为：`按钮文字|按钮命令||按钮文字2|按钮命令2`如果按钮命令是链接，就是个链接按钮。否则可以是内部指令。比如 `close` 就是删除按钮消息（其他暂时是内部用的，可以不用披露）。

`ReplyMode`     是回复模式，可以为 `deleteself`, `deleteorigin`, `deleteboth` 表示为 10秒后只删除 `ReplyMessage` 发送的消息、10秒后只删除触发这个规则的消息，10秒后两个消息都删

`ReplyImage`     回复消息的头图，可以让 ReplyMessage 更好看

`InvokeOptions`     用户触发规则（之前的定义都是全局的，即便是 limit 也是只这个规则总的来说能触发多少次，但并没有限制单个用户是否能触发/触发多少次。1.14以后加了这个选项来对这个功能进行补足）

**下面都是 InvokeOptions 下的配置：**

`Rule`  指定用户触发规则，可以为 `unlimit`, `peruser`, `peruserinterval`, `peruserday` 表示为不限制。限制一个用户能触发多少次，限制一个用户多少时间（second）能触发一次，限制用户每天能触发多少次。默认为不限制

`Value`     只 Rule 里面的多少

`ShowError`     用户超出限制以后是否展示错误信息

`Reset`     导入规则以后是否重置计数器（如果false假使设定了peruserday value=1即便/import_policy更新规则，今天触发过的用户还是触发不了）

`UserOnly`  只有用户能触发，匿名不能触发

### Policy案例
```
{
  "ID": 群ID,
  "Admins": [
    AdminID
  ],
  "BannedForward": [],
  "MergeTo": 0,
  "Locale": "zh",
  "MustFollow": "",
  "MustFollowOnJoin": false,
  "MustFollowOnMsg": false,
  "GroupAPISignSeed": "",
  "CreditMapping": {
    "PerValidTextMessage": 1,
    "PerValidStickerMessage": 1,
    "Command": -1,
    "Duplicated": -2,
    "Warn": -10,
    "Ban": -50,
    "BanBouns": 15,
    "HourlyUpperBound": 30
  },
  "UnderAttackMode": false,
  "AntiSpoiler": false,
  "DisableWarn": false,
  "RedPacketCaptcha": true,
  "RedPacketCaptchaFailCreditBehavior": -5,
  "WarnKeywords": [
    "warn",
    "/warn",
    "/sb"
  ],
  "BanKeywords": [
    "ban",
    "/ban"
  ],
  "NameBlackListReg": [  
    "^关键词",
    "关键词2"
  ],
  "CustomReply": [
    {
      "Match": "^关键词$|^关键词2$|^关键词3$",
      "Name": "细则名称",
      "Limit": -1,
      "CreditBehavior": 0,
      "NoForceCreditBehaviorError": "",
      "CallbackURL": "",
      "ReplyMessage": "⬇️ [{{UserName}}]({{UserLink}})，请点击下面链接进入相关的页面哦 ～",
      "ReplyTo": "",
      "ReplyButtons": [
        "Link1|https://baidu.com||Link2|https://google.com",
        "Link3|https://baidu.com||Link4|https://google.com",
        "👌 知道啦|close"
      ],
      "ReplyMode": "",
      "ReplyImage": "",
      "InvokeOptions": {
        "Rule": "unlimit",
        "Value": 0,
        "ShowError": true,
        "Reset": false,
        "UserOnly": true
      }
    },
    {
      "Match": "://|\\.com|\\.me|\\.cn|\\.cc|\\.xyz|\\.ml|\\.pub|\\.la|\\.pro|\\.us|\\.biz|\\.name|\\.jp",
      "Name": "禁止链接",
      "Limit": -1,
      "CreditBehavior": -30,
      "NoForceCreditBehaviorError": "",
      "CallbackURL": "",
      "ReplyMessage": "😠 [{{UserName}}]({{UserLink}})，请不要发送链接哦 …",
      "ReplyTo": "",
      "ReplyButtons": [],
      "ReplyMode": "deleteboth",
      "ReplyImage": "",
      "InvokeOptions": {
        "Rule": "unlimit",
        "Value": 0,
        "ShowError": true,
        "Reset": false,
        "UserOnly": false
      }
    },
    {
      "Match": "关键词1|关键词2|关键词3|关键词4",
      "Name": "禁止广告",
      "Limit": -1,
      "CreditBehavior": -30,
      "NoForceCreditBehaviorError": "",
      "CallbackURL": "",
      "ReplyMessage": "😠 [{{UserName}}]({{UserLink}})，请不要发送广告哦 …",
      "ReplyTo": "",
      "ReplyButtons": [],
      "ReplyMode": "deleteboth",
      "ReplyImage": "",
      "InvokeOptions": {
        "Rule": "unlimit",
        "Value": 0,
        "ShowError": true,
        "Reset": false,
        "UserOnly": false
      }
    },
    {
      "Match": "^积分规则|积分怎么获得|积分查看|^积分$|^获取积分|^获得积分|^怎么获取积分|^怎么获得积分",
      "Name": "积分规则",
      "Limit": -1,
      "CreditBehavior": 0,
      "NoForceCreditBehaviorError": "",
      "CallbackURL": "",
      "ReplyMessage": "⬇️ [{{UserName}}]({{UserLink}})您好，积分查看命令为 `/mycredit`,本群积分规则为：水群聊天增加 **1** 点，使用指令扣除 **1** 点积分，刷屏扣除 **10** 点积分。另可以每日签到来获取积分，指令有： `/签到` 等 ～",
      "ReplyTo": "",
      "ReplyButtons": [],
      "ReplyMode": "",
      "ReplyImage": "",
      "InvokeOptions": {
        "Rule": "unlimit",
        "Value": 0,
        "ShowError": true,
        "Reset": false,
        "UserOnly": true
      }
    },
    {
      "Match": "^/che|^/chi|^/cke|^/kfc|^/mac|^/crazy|^/qian|^/turkey|^/shakeshark|^/签到|^/疯狂|^/打卡|^/hase|^签到$|^打卡",
      "Name": "每日签到",
      "Limit": -1,
      "CreditBehavior": 15,
      "NoForceCreditBehaviorError": "",
      "CallbackURL": "",
      "ReplyMessage": "🎉[{{UserName}}]({{UserLink}}),感谢您的签到,15分到账~~",
      "ReplyTo": "",
      "ReplyButtons": [],
      "ReplyMode": "deleteboth",
      "ReplyImage": "",
      "InvokeOptions": {
        "Rule": "peruserday",
        "Value": 1,
        "ShowError": true,
        "Reset": false,
        "UserOnly": true
      }
    }
  ]
}
```
