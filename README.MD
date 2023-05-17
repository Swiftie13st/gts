# 一个伟大的史诗级项目

GTS(Go Tcp Server)使用golang编写的一个高性能TCP服务器框架

1. 通过TLV(Type-Len-Value)封包格式解决TCP粘包问题
2. 通过协程池复用对象以及限制最大协程数量，减少内存分配的频率以及协程创建开销，提高执行效率。