# woker-BI   
### 数据分析处理中心

### 安装
> 1. golang 1.17 +   
> 2. mysql

### 依赖工具
>1. gocron任务调度工具 [教程](https://github.com/ouqiang/gocron/wiki) 

### TEST 
```sh
go test cxe/model -v -count=1  // count 禁用编译缓存
go test ./... 
```

### linux构建
```sh
#sh
set GOARCH=amd64
set GOOS=linux
go mod tidy
go build
```