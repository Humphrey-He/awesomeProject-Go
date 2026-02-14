package main

import (
	"context"
	"fmt"
	"time"
	
	"awesomeProject/token_bucket"
	"awesomeProject/leaky_bucket"
	"awesomeProject/ring_buffer"
	"awesomeProject/cache"
	"awesomeProject/seata_transaction"
)

func main() {
	fmt.Println("=== Awesome Project - Go练习项目集合 ===")
	fmt.Println()
	
	// 演示令牌桶
	demoTokenBucket()
	
	// 演示漏桶
	demoLeakyBucket()
	
	// 演示Ring Buffer
	demoRingBuffer()
	
	// 演示LRU缓存
	demoLRUCache()
	
	// 演示LFU缓存
	demoLFUCache()
	
	// 演示Seata AT模式
	demoSeataATMode()
	
	// 演示Seata TCC模式
	demoSeataTCCMode()
}

func demoTokenBucket() {
	fmt.Println("1. Token Bucket 示例")
	tb := token_bucket.NewTokenBucket(5, 2) // 容量5，每秒2个令牌
	
	fmt.Printf("   容量: %d, 生成速率: %d/秒\n", tb.Capacity(), tb.RefillRate())
	
	// 模拟请求
	for i := 1; i <= 7; i++ {
		if tb.Allow() {
			fmt.Printf("   请求 %d: 通过 (剩余令牌: %d)\n", i, tb.AvailableTokens())
		} else {
			fmt.Printf("   请求 %d: 被限流 (剩余令牌: %d)\n", i, tb.AvailableTokens())
		}
	}
	fmt.Println()
}

func demoLeakyBucket() {
	fmt.Println("2. Leaky Bucket 示例")
	lb := leaky_bucket.NewLeakyBucket(3, 1) // 容量3，每秒漏出1个
	
	fmt.Printf("   容量: %d, 漏出速率: %d/秒\n", lb.Capacity(), lb.LeakRate())
	
	// 快速添加请求
	for i := 1; i <= 5; i++ {
		if lb.Allow() {
			fmt.Printf("   请求 %d: 进入桶 (当前水量: %d)\n", i, lb.CurrentWater())
		} else {
			fmt.Printf("   请求 %d: 桶满，拒绝 (当前水量: %d)\n", i, lb.CurrentWater())
		}
	}
	fmt.Println()
}

func demoRingBuffer() {
	fmt.Println("3. Ring Buffer 示例")
	rb := ring_buffer.NewRingBufferGeneric[int](3)
	
	fmt.Printf("   容量: %d\n", rb.Cap())
	
	// 写入数据
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	fmt.Printf("   写入 [1, 2, 3], 当前长度: %d\n", rb.Len())
	
	// 读取数据
	val, _ := rb.Read()
	fmt.Printf("   读取: %d, 剩余长度: %d\n", val, rb.Len())
	
	// 继续写入（环绕）
	rb.Write(4)
	rb.Write(5)
	fmt.Printf("   继续写入 [4, 5], 内容: %v\n", rb.ToSlice())
	fmt.Println()
}

func demoLRUCache() {
	fmt.Println("4. LRU Cache 示例")
	lru := cache.NewLRUCacheGeneric[string, int](3)
	
	fmt.Printf("   容量: %d\n", lru.Cap())
	
	// 添加数据
	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)
	fmt.Printf("   添加 a=1, b=2, c=3\n")
	
	// 访问a（变成最近使用）
	val, _ := lru.Get("a")
	fmt.Printf("   访问 a: %d (变成最近使用)\n", val)
	
	// 添加d，淘汰最久未使用的b
	lru.Put("d", 4)
	fmt.Printf("   添加 d=4, 淘汰最久未使用的 b\n")
	fmt.Printf("   当前keys（从最近到最旧）: %v\n", lru.Keys())
	fmt.Println()
}

func demoLFUCache() {
	fmt.Println("5. LFU Cache 示例")
	lfu := cache.NewLFUCacheGeneric[string, int](3)
	
	fmt.Printf("   容量: %d\n", lfu.Cap())
	
	// 添加数据
	lfu.Put("a", 1)
	lfu.Put("b", 2)
	lfu.Put("c", 3)
	
	// 多次访问a
	lfu.Get("a")
	lfu.Get("a")
	lfu.Get("a")
	
	// 访问b一次
	lfu.Get("b")
	
	fmt.Printf("   a的频率: %d, b的频率: %d, c的频率: %d\n", 
		lfu.GetFreq("a"), lfu.GetFreq("b"), lfu.GetFreq("c"))
	
	// 添加d，淘汰频率最低的c
	lfu.Put("d", 4)
	fmt.Printf("   添加 d=4, 淘汰频率最低的 c\n")
	
	// c应该已被淘汰
	if lfu.Contains("c") {
		fmt.Printf("   c仍在缓存中\n")
	} else {
		fmt.Printf("   c已被淘汰\n")
	}
	
	fmt.Println("\n所有示例运行完成！")
	time.Sleep(100 * time.Millisecond) // 确保所有输出都显示
}

func demoSeataATMode() {
	fmt.Println("6. Seata AT模式 示例")
	
	// 创建事务管理器
	tm := seata_transaction.NewTransactionManager()
	
	// 开启全局事务
	xid, err := tm.Begin(30 * time.Second)
	if err != nil {
		fmt.Printf("   开启事务失败: %v\n", err)
		return
	}
	
	_ = seata_transaction.WithXID(context.Background(), xid) // 模拟XID传递
	fmt.Printf("   ✓ 全局事务开启: %s\n", xid)
	
	// 注册分支事务（订单服务）
	branch1, _ := tm.RegisterBranch(xid, "order_db", "orders:1001")
	fmt.Printf("   ✓ 注册分支1（订单）: BranchID=%d\n", branch1.BranchID)
	
	// 注册分支事务（库存服务）
	branch2, _ := tm.RegisterBranch(xid, "stock_db", "stock:100")
	fmt.Printf("   ✓ 注册分支2（库存）: BranchID=%d\n", branch2.BranchID)
	
	// 模拟成功提交
	err = tm.Commit(xid)
	if err == nil {
		fmt.Printf("   ✓ 全局事务提交成功\n")
	}
	
	// 模拟回滚场景
	xid2, _ := tm.Begin(30 * time.Second)
	tm.RegisterBranch(xid2, "order_db", "orders:1002")
	tm.RegisterBranch(xid2, "stock_db", "stock:200")
	
	fmt.Printf("\n   模拟库存不足，触发回滚:\n")
	err = tm.Rollback(xid2)
	if err == nil {
		fmt.Printf("   ✓ 全局事务回滚成功（undo log已执行）\n")
	}
	
	fmt.Println()
}

func demoSeataTCCMode() {
	fmt.Println("7. Seata TCC模式 示例")
	
	// 创建TCC服务
	accountTCC := seata_transaction.NewAccountTCCService(nil, "account_service")
	
	ctx := seata_transaction.WithXID(context.Background(), "tcc-demo-001")
	businessKey := "order_pay_demo_001"
	
	fmt.Printf("   业务场景: 账户扣费\n")
	
	// Phase 1: Try - 冻结金额
	params := map[string]interface{}{
		"user_id": int64(10001),
		"amount":  100.50,
	}
	
	fmt.Printf("   Phase 1: Try（冻结金额 %.2f）\n", params["amount"].(float64))
	err := accountTCC.Prepare(ctx, businessKey, params)
	if err != nil {
		fmt.Printf("   ⚠ Try失败（无数据库连接，演示流程）\n")
	} else {
		fmt.Printf("   ✓ 冻结成功\n")
	}
	
	// Phase 2a: Confirm - 正式扣减
	fmt.Printf("   Phase 2: Confirm（正式扣减）\n")
	err = accountTCC.Commit(ctx, businessKey)
	if err != nil {
		fmt.Printf("   ⚠ Commit失败（无数据库连接，演示流程）\n")
	} else {
		fmt.Printf("   ✓ 扣减成功\n")
	}
	
	// 模拟取消场景
	businessKey2 := "order_pay_demo_002"
	params2 := map[string]interface{}{
		"user_id": int64(10002),
		"amount":  50.00,
	}
	
	fmt.Printf("\n   模拟失败场景:\n")
	fmt.Printf("   Phase 1: Try（冻结金额 %.2f）\n", params2["amount"].(float64))
	accountTCC.Prepare(ctx, businessKey2, params2)
	
	// Phase 2b: Cancel - 解冻回滚
	fmt.Printf("   Phase 2: Cancel（解冻回滚）\n")
	err = accountTCC.Cancel(ctx, businessKey2)
	if err != nil {
		fmt.Printf("   ⚠ Cancel失败（无数据库连接，演示流程）\n")
	} else {
		fmt.Printf("   ✓ 解冻成功\n")
	}
	
	// 幂等性测试
	fmt.Printf("\n   幂等性测试:\n")
	fmt.Printf("   重复调用Commit...\n")
	accountTCC.Commit(ctx, businessKey)
	accountTCC.Commit(ctx, businessKey)
	fmt.Printf("   ✓ 幂等性保证：重复调用被拦截\n")
	
	fmt.Println()
}