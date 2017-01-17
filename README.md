# store


* 实现 Storer接口 支持 文件存储和七牛云存储

## 文件存储支持

### 初始化

``` go
  import "github.com/antlinker/store/file"
  // 在这里指定根目录
  err:=file.InitStore("root")
  // 自定义store使用该方法
  selfstore,err:=  file.CreateStore(rootDir )
```
## 七牛云存储支持

### 初始化

``` go
  import "github.com/antlinker/store/qiniu"
  // 在这里指定AK SK 和容器
  err:=store.InitQiniuStore("AK", "SK", "mytest")
  // 使用mgodb数据库存储配置信息
  CreDefaultStoreByMGO(cfg MgoCfg)
```
