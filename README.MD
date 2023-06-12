# 一个伟大的史诗级项目

GTS(Go Tcp Server)使用golang编写的一个轻量级、高性能TCP服务器框架。

1. 提供抽象层接口，可进行二次开发作为后端模块的消息中转、长链接游戏服务器等功能。
2. 通过TLV(Type-Len-Value)封包格式解决TCP粘包问题
3. 仿造[ants](https://github.com/panjf2000/ants)实现协程池进行资源复用，减少内存分配的频率，提高执行效率。
4. 通过心跳检测机制监测双方连接状态。
5. 通过雪花算法分配连接id。
6. 支持TCP、WebSocket协议
7. 

## 架构设计

![架构图](https://raw.githubusercontent.com/Swiftie13st/Figurebed/main/img/202305242242762.png)

## TODO

1. 使用高性能的网络库，如 [evio](https://github.com/tidwall/evio) 或 [gnet](https://github.com/panjf2000/gnet)，可以提高服务器的性能。这些库经过优化，以减少内存分配和垃圾回收开销。
2. 增加UDP、~~Websocket~~等网络协议
3. 优化数据处理，使用protobuf等
4. 添加身份认证模块