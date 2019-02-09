### 介绍
一个用于下载bagoum.com上的卡牌<del>老婆</del>数据的一个小工具。
### 安装
非Windows用户：  
运行`go get -v github.com/coderbaka/bagoum-card-spider`  
Windows用户：  
请下载exe版本 https://github.com/coderbaka/bagoum-card-spider/releases/download/v0.0.1/bagoum-card-spider.exe  
依赖的第三方库: github.com/jessevdk/go-flags  github.com/gocolly/colly

### 帮助
代开终端/CMD，输入`bagoum-card-spider -h`
### TODO
由于用Go写爬虫极其难受，加上要开学了，所以大概是不会更新了。
1. 增加卡牌信息下载，增加多语言支持
2. <del>用Python重写一遍</del>
### 目前的bug
1. 有些不是随从卡没有进化后的版本，但还是会下载。这些卡下的evo-*文件通常都是打不开的（为纯文本格式）。但不影响使用。
### 截图
![screenshot-1.png](https://i.loli.net/2019/02/09/5c5eb2698e5de.png)
![screensho-3.png](https://i.loli.net/2019/02/09/5c5eb2699053e.png)
![screenshot-2.png](https://i.loli.net/2019/02/09/5c5eb26997348.png)

