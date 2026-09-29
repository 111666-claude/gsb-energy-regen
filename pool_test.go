package energy

import (
	"math/rand"
	"testing"
	"time"
)

func TestRegenForward(t *testing.T) {
	pool := &Pool{Energy: 0, Max: 20, RatePerMin: 1, UpdatedAt: 0}
	if got := pool.Regen(300); got != 5 {
		t.Fatalf("5 分钟应该回 5 点：%d", got)
	}
}

func TestSpend(t *testing.T) {
	pool := &Pool{Energy: 10, Max: 20, RatePerMin: 1}
	if err := pool.Spend(4); err != nil {
		t.Fatalf("消耗出错：%v", err)
	}
	if pool.Energy != 6 {
		t.Fatalf("消耗后体力不对：%d", pool.Energy)
	}
	if err := pool.Spend(100); err != ErrEnergy {
		t.Fatalf("透支应该报错：%v", err)
	}
}

func TestUpdatedAtAdvanced(t *testing.T) {
	pool := &Pool{Energy: 1, Max: 20, RatePerMin: 2, UpdatedAt: 0}
	pool.Regen(600)
	if pool.UpdatedAt != 600 {
		t.Fatalf("时间戳应该推进：%d", pool.UpdatedAt)
	}
}

func TestSpendRejectsZero(t *testing.T) {
	if err := (&Pool{Energy: 10}).Spend(0); err != ErrEnergy {
		t.Fatalf("0 消耗应该报错：%v", err)
	}
}

func TestRegenClockBack(t *testing.T) {
	pool := &Pool{Energy: 10, Max: 20, RatePerMin: 1, UpdatedAt: 1000}
	if got := pool.Regen(700); got != 10 {
		t.Fatalf("回拨不应倒扣：%d", got)
	}
	if pool.UpdatedAt != 1000 {
		t.Fatalf("回拨不应回退时间戳：%d", pool.UpdatedAt)
	}
}

func TestRegenCapsAtMax(t *testing.T) {
	pool := &Pool{Energy: 20, Max: 20, RatePerMin: 1, UpdatedAt: 0}
	if got := pool.Regen(600); got != 20 {
		t.Fatalf("满体力不应超过上限：%d", got)
	}
}

func TestRegenCrossDayResets(t *testing.T) {
	pool := &Pool{Energy: 3, Max: 20, RatePerMin: 1, UpdatedAt: 86340}
	if got := pool.Regen(86400 + 120); got != 20 {
		t.Fatalf("跨天应重置到 Max：%d", got)
	}
	if pool.UpdatedAt != 86400+120 {
		t.Fatalf("跨天后时间戳应推进：%d", pool.UpdatedAt)
	}
}

func TestRegenSameDayNoReset(t *testing.T) {
	pool := &Pool{Energy: 3, Max: 20, RatePerMin: 1, UpdatedAt: 120}
	if got := pool.Regen(300); got != 6 {
		t.Fatalf("同一天应正常回复：%d", got)
	}
}

func TestRegenIdempotent(t *testing.T) {
	pool := &Pool{Energy: 5, Max: 20, RatePerMin: 2, UpdatedAt: 0}
	first := pool.Regen(600)
	second := pool.Regen(600)
	if first != second {
		t.Fatalf("重复结算结果应相同：%d != %d", first, second)
	}
}

// referenceRegen 是「服务端时间 + 截断」的参考实现，用于随机对照。
func referenceRegen(energy, max, rate int, updatedAt, now int64) (int, int64) {
	if now < updatedAt {
		now = updatedAt
	}
	if now/secondsPerDay != updatedAt/secondsPerDay {
		return max, now
	}
	energy += int((now-updatedAt)/60) * rate
	if energy > max {
		energy = max
	}
	if energy < 0 {
		energy = 0
	}
	return energy, now
}

func TestRegenRandomTimelines(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	const timelines = 200
	for tl := 0; tl < timelines; tl++ {
		max := 1 + rng.Intn(100)
		rate := 1 + rng.Intn(5)
		pool := &Pool{Energy: rng.Intn(max + 1), Max: max, RatePerMin: rate, UpdatedAt: rng.Int63n(10 * secondsPerDay)}
		refEnergy, refUpdatedAt := pool.Energy, pool.UpdatedAt
		now := pool.UpdatedAt
		for step := 0; step < 50; step++ {
			switch rng.Intn(4) {
			case 0: // 前进
				now += rng.Int63n(3600)
			case 1: // 回拨
				now -= rng.Int63n(3600)
			case 2: // 跨天
				now += secondsPerDay + rng.Int63n(secondsPerDay)
			case 3: // 满体力
				pool.Energy = pool.Max
				refEnergy = max
			}
			got := pool.Regen(now)
			refEnergy, refUpdatedAt = referenceRegen(refEnergy, max, rate, refUpdatedAt, now)
			if got != refEnergy {
				t.Fatalf("时间线 %d 第 %d 步：got=%d want=%d", tl, step, got, refEnergy)
			}
			if pool.UpdatedAt != refUpdatedAt {
				t.Fatalf("时间线 %d 第 %d 步时间戳：got=%d want=%d", tl, step, pool.UpdatedAt, refUpdatedAt)
			}
			if got < 0 || got > max {
				t.Fatalf("时间线 %d 第 %d 步越界：%d 不在 [0, %d]", tl, step, got, max)
			}
		}
	}
}

func TestRegenScale(t *testing.T) {
	const pools = 100000
	const settlements = 100000
	store := make([]Pool, pools)
	for i := range store {
		store[i] = Pool{Energy: 0, Max: 20, RatePerMin: 1, UpdatedAt: 0}
	}
	now := time.Now().Unix()
	start := time.Now()
	for i := 0; i < settlements; i++ {
		store[i%pools].Regen(now)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("%d 次结算超过 1 秒：%v", settlements, elapsed)
	}
}
