# UidGenerator

1. 基于[Snowflake](https://github.com/twitter/snowflake)算法的唯一 ID 生成器: `1 + 41 + 10 + 13`
2. UidGenerator 通过借用未来时间来解决 sequence 天然存在的并发限制
3. 采用 RingBuffer 来缓存已生成的 UID, 并行化 UID 的生产和消费
4. 同时对 CacheLine 补齐, 避免了由 RingBuffer 带来的硬件级「伪共享」问题.
5. 启动阶段通过 DB 进行分配; 如自定义实现, 则 DB 非必选依赖

## 实现-UidGenerator

![Snowflake](doc/snowflake.png)

1. Snowflake 算法描述: 指定机器 & 同一时刻 & 某一并发序列, 是唯一的, 据此可生成一个 64 bits 的唯一 ID（long）
2. 默认采用上图字节分配方式(可自定义):

   - sign(1bit): 固定 1bit 符号标识, 即生成的 UID 为正数
   - delta seconds (28 bits) : 当前时间, 相对于时间基点"2016-05-20"的增量值, 单位: 秒, 最多可支持约 8.7 年
   - worker id (22 bits): 机器 id, 最多可支持约 420w 次机器启动. 内置实现为在启动时由数据库分配, 默认分配策略为用后即弃, 后续可提供复用策略
   - sequence (13 bits): 每秒下的并发序列, 13 bits 可支持每秒 8192 个并发.

## 实现-CachedUidGenerator

![RingBuffer](doc/ringbuffer.png)

1. CachedUidGenerator 采用了双 RingBuffer[slot/ 2^sequence], Uid-RingBuffer 用于存储 Uid、Flag-RingBuffer 用于存储 Uid 状态(是否可填充、是否可消费)
2. Tail 指针: 生产

   - 表示 Producer 生产的最大序号(此序号从 0 开始, 持续递增).
   - Tail 不能超过 Cursor, 即生产者不能覆盖未消费的 slot
   - 当 Tail 已赶上 curosr, 此时可通过`rejectedPutBufferHandler`指定 PutRejectPolicy

3. Cursor 指针: 消费

   - 表示 Consumer 消费到的最小序号(序号序列与 Producer 序列相同)
   - Cursor 不能超过 Tail, 即不能消费未生产的 slot
   - 当 Cursor 已赶上 tail, 此时可通过`rejectedTakeBufferHandler`指定 TakeRejectPolicy

4. 由于数组元素在内存中是连续分配的, 可最大程度利用 CPU cache 以提升性能. 但同时会带来「伪共享」FalseSharing 问题, 为此在 Tail、Cursor 指针、Flag-RingBuffer 中采用了 CacheLine
   补齐方式.

   ![FalseSharing](doc/cacheline_padding.png)

5. RingBuffer 填充

   - 初始化预填充 : RingBuffer 初始化时, 预先填充满整个 RingBuffer.
   - 即时填充 : Take 消费时, 即时检查剩余可用 slot 量(`tail-cursor`), 如小于设定阈值, 则补全空闲 slots. 阈值可通过`paddingFactor`来进行配置
   - 周期填充 : 通过 Schedule 线程, 定时补全空闲 slots. 可通过`scheduleInterval`配置, 以应用定时填充功能, 并指定 Schedule 时间间隔

## 运行示例单测

## 关于 UID 比特分配的建议

1. 对于并发数要求不高、期望长期使用的应用, 可增加`timeBits`位数, 减少`seqBits`位数

   - 例如节点采取用完即弃的 WorkerIdAssigner 策略, 重启频率为 12 次/天,
   - `{"workerBits":23,"timeBits":31,"seqBits":9}`时, 可支持 28 个节点以整体并发量 14400 UID/s 的速度持续运行 68 年.

2. 对于节点重启频率频繁、期望长期使用的应用, 可增加`workerBits`和`timeBits`位数, 减少`seqBits`位数

   - 例如节点采取用完即弃的 WorkerIdAssigner 策略, 重启频率为 24\*12 次/天,
   - `{"workerBits":27,"timeBits":30,"seqBits":6}`时, 可支持 37 个节点以整体并发量 2400 UID/s 的速度持续运行 34 年.

---

## feature

1. 适用于 docker 等虚拟化环境下实例自动重启、漂移等场景
2. 支持自定义 workerId 位数和初始化策略
3. 最终单机 QPS 可达 600 万
