# 需求调研
## 现状
目前业务中的使用生成器，主要用于生成唯一ID。
主要场景有生成用户I号，生成帖子ID。
目前各个业务使用redis的INCR方式生成自增的唯一ID。

## 总结
基本上常见的分布式情况下的UniqueID的生成方法就分为了两大类,
一类是基于分配的发号器,另外一类是基于规则计算的唯一序列。
而后者的常用算法通常就是UUID以及OjbectID和SnowFlake
基于发号器的优势在于可以按照准连续的增加,可以用于Int32等的存储。但是性能和系统复杂度上存在一定的缺陷。
而基于计算规则的优势主要是分布式情况下,各机器独立运算,性能上有保证。但是通常都需要使用64位以上的空间来进行存储。
目前业务中很少业务使用规则计算的唯一序列，都是基于分配的发号器。
因此优先需要实现分配的发号器。

# 已有实现
## 其他厂商实现
- [有赞](https://tech。youzan。com/id_generator)
- [阅文](http://geek。csdn。net/news/detail/82281)
- [twitter](https://github。com/twitter/finagle)
- [微信](http://www。infoq。com/cn/articles/wechat-serial-number-generator-architecture)
- [flickr](http://code。flickr。net/2010/02/08/ticket-servers-distributed-unique-primary-keys-on-the-cheap)
- [58同城](http://www。ita1024。com/eventlist/view/id/67)
- [mongo](https://docs。mongodb。com/manual/reference/method/ObjectId/)
- [微博](http://upyun-open-talk.b0.upaiyun.com/sns.pdf)

## 个人实现
- [dhetis](github.com/lsytj0413/dhetis)
- [idGenerator](https://github.com/cclehui/idGenerator)
- [vesta](https://github.com/robertleepeak/vesta-id-generator)

