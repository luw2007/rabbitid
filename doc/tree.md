# store 目录维护
## 问题
为了接入mysql中所有的主键，这里需要将原有`/rabbitid/[dc]/[db]`增加层级`/rabbitid/[dc]/[db]/[table]`。
`/rabbitid/[dc]`是提前创建好，进程启动也会检查。
## 解决方案
1. 增加1个`create`的方法用来提前创建`db`目录
2. redis 部分方法是支持单个参数，这里使用"|"分隔db和table，用在一些查询方法中。


