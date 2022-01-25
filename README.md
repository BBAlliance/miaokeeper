# Miaokeeper  

> Miaokeeper 是一个群成员自主管理机器人，可以在 telegram 群组中实现：群成员自主管理、入群验证、积分统计、积分抽奖等功能。  
>
> ## 前期准备  
>
> 1.会使用 `Screen、Pm2、Systemd、Supervisor` 其中任意一直进程守护方式。  
> 2.会搭建 `Go语言编译环境` 、或会使用 `Docker的基础用法` 。  
> 3.会使用 `MySQL` 或其他数据库。  
>
> ## 如何安装  
>
> 目前支持  `直接安装` 和 `docker` 安装两种模式    
>

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
> 4.启动成功后通过 `-setadmin 用户ID` 添加管理。或者在启动时填入参数，若在启动时添加将会自动退出程序。若无报错，请去掉 `-setadmin 用户ID` 后再次启动。

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

> 4.使用 `Docker-compose`  命令启动Docker容器。  
> 5.使用 `docker exec -it 容器名 or ID bash` 进入容器，进入容器后通过 `./miaokeeper  -setadmin 123456 -token <YOUR_TOKEN> -database root:miaokeeper2022@tcp\(mariadb:3306\)/miaokeeper`  添加管理。其中 `-setadmin   `  后加上TG用户ID。若无报错，说明添加成功，重启容器即可。  
> 
> 例如：  

```bash
./miaokeeper  -setadmin 123456 -token 123456:XXXXXXXXXXXXXXXX -database root:miaokeeper2022@tcp\(mariadb:3306\)/miaokeeper 

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
-ping                  #测试bot和电报服务器之间的往返时间
-token string          #电报机器人令牌
-upstream string       #电报上游api url
-verbose               #显示所有日志
-version               #显示当前版本并退出
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
/set_credit           #回复或id设置积分
/add_credit           #回复或id添加积分
/check_credit         #查看某群友积分
/set_antispoiler      #是否开启剧透
/set_channel          #绑定频道（回复空则解绑频道） 要把bot扔进频道给管理
/set_locale           #设置群组的默认语言，可加一个参数，如果是 - 则为清空设置
/send_redpacket       #发运气红包
/create_lottery       #开启抽奖  create_lottery 奖品名称 :limit=所需积分:consume=（n|y）是否扣除积分 :num=奖品数量 :participant=参与人数
/creditrank           #开榜获取积分排行榜
/redpacket            #积分红包请输入 /redpacket <总分数> <红包个数> 来发红包哦～
/lottery              #抽奖（可加A B两个参数，从A总人数中抽B人数）

```

### `用户可用命令`  

```
/mycredit      #我的积分
/version       #版本查询
/ping          #检测bot和群组响应速度
```
