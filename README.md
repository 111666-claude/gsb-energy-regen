# energy-regen

体力池：按时间回复、消耗与上限，只用 Go 标准库。

```
go test ./...
go run ./cmd/energy-regen --sample clockback
go run ./cmd/energy-regen --sample overflow
```

## 口径（README 为准）

- **回拨时间不受理**：`now < UpdatedAt` 时必须按 UpdatedAt 结算（体力不变，也不倒扣），
  `--sample clockback`（体力 10、UpdatedAt=1000，客户端给 now=700）按口径是 `energy=10`，
  现在按负数分钟倒扣，给 `energy=5`。
- **上限要截断**：回复后不能超过 `Max`。`--sample overflow`（体力 20/20、10 分钟）按口径是
  `energy=20`，现在给 `energy=30`。
- **跨天重置**：服务器日期变化时体力重置到 `Max`（客户端时间不可信，用服务端时间戳判断）。
- 不变量：体力永远在 `[0, Max]`；同一段时间重复结算不会重复回复（UpdatedAt 单调不减）；
  同一批请求重复执行结果相同。
- 规模：10 万个体力池、每秒 10 万次结算 1 秒内完成，内存 O(池数)。

## 现在的行为

- `Regen` 直接用 `(now-UpdatedAt)/60` 算分钟数，负数就倒扣。
- 没有 `Max` 截断。
- 没有跨天重置逻辑。

## 输出

```
energy=10
energy=20 max=20
```

## 目录

```
pool.go             体力池
cmd/energy-regen    命令行入口
pool_test.go        go test 用例
```
