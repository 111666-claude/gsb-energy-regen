package energy

import (
	"math/rand"
	"testing"
	"time"
)

func TestRegenClockBackRejected(t *testing.T) {
	pool := &Pool{Energy: 10, Max: 20, RatePerMin: 1, UpdatedAt: 1000}
	if got := pool.Regen(700); got != 10 {
		t.Fatalf("回拨不应倒扣体力：%d", got)
	}
	if pool.UpdatedAt != 1000 {
		t.Fatalf("回拨不应回退 UpdatedAt：%d", pool.UpdatedAt)
	}
}

func TestRegenCappedAtMax(t *testing.T) {
	pool := &Pool{Energy: 20, Max: 20, RatePerMin: 1, UpdatedAt: 0}
	if got := pool.Regen(600); got != 20 {
		t.Fatalf("满体力回复应截断到 Max：%d", got)
	}
}

func TestRegenDayReset(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 23, 59, 0, 0, time.UTC).Unix()
	day2 := day1 + 120 // 次日 00:01
	pool := &Pool{Energy: 3, Max: 20, RatePerMin: 1, UpdatedAt: day1}
	if got := pool.Regen(day2); got != 20 {
		t.Fatalf("跨天应重置到 Max：%d", got)
	}
	if pool.UpdatedAt != day2 {
		t.Fatalf("跨天后 UpdatedAt 应推进：%d", pool.UpdatedAt)
	}
}

func TestRegenSameDayNoReset(t *testing.T) {
	base := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC).Unix()
	pool := &Pool{Energy: 3, Max: 20, RatePerMin: 1, UpdatedAt: base}
	if got := pool.Regen(base + 300); got != 8 {
		t.Fatalf("同一天内不应触发重置：%d", got)
	}
}

func TestRegenIdempotent(t *testing.T) {
	pool := &Pool{Energy: 5, Max: 20, RatePerMin: 2, UpdatedAt: 0}
	first := pool.Regen(600)
	second := pool.Regen(600)
	if first != second {
		t.Fatalf("同一时间重复结算结果应一致：%d != %d", first, second)
	}
	if pool.UpdatedAt != 600 {
		t.Fatalf("UpdatedAt 应停在 600：%d", pool.UpdatedAt)
	}
}

func TestRegenUpdatedAtMonotone(t *testing.T) {
	pool := &Pool{Energy: 0, Max: 20, RatePerMin: 1, UpdatedAt: 500}
	for _, now := range []int64{400, 300, 500, 499, 1000, 999} {
		before := pool.UpdatedAt
		pool.Regen(now)
		if pool.UpdatedAt < before {
			t.Fatalf("UpdatedAt 不应回退：%d -> %d", before, pool.UpdatedAt)
		}
	}
}

// referenceRegen 是独立的参考实现：服务端时间（回拨按 UpdatedAt）+ 上限截断 + 跨天重置，
// 用秒数 floor 除法算日期，与 pool.go 的 time 包实现走不同路径。
func referenceRegen(energy, max, rate int, updatedAt, now int64) (int, int64) {
	effective := now
	if effective < updatedAt {
		effective = updatedAt
	}
	if floorDiv(effective, 86400) != floorDiv(updatedAt, 86400) {
		energy = max
	} else {
		energy += int((effective-updatedAt)/60) * rate
	}
	if energy > max {
		energy = max
	}
	if energy < 0 {
		energy = 0
	}
	return energy, effective
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func TestRegenAgainstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	const day = int64(86400)
	for trial := 0; trial < 200; trial++ {
		max := 1 + rng.Intn(50)
		rate := rng.Intn(5)
		pool := &Pool{
			Energy:     rng.Intn(max + 1),
			Max:        max,
			RatePerMin: rate,
			UpdatedAt:  rng.Int63n(30 * day),
		}
		refEnergy, refUpdatedAt := pool.Energy, pool.UpdatedAt
		now := pool.UpdatedAt
		for step := 0; step < 20; step++ {
			switch rng.Intn(4) {
			case 0: // 前进
				now += rng.Int63n(7200)
			case 1: // 回拨
				now -= rng.Int63n(7200)
			case 2: // 跨天
				now = (now/day+1)*day + rng.Int63n(3600)
			case 3: // 满体力再前进
				pool.Energy = pool.Max
				refEnergy = max
				now += rng.Int63n(3600)
			}
			got := pool.Regen(now)
			refEnergy, refUpdatedAt = referenceRegen(refEnergy, max, rate, refUpdatedAt, now)
			if got != refEnergy {
				t.Fatalf("trial=%d step=%d：实现=%d 参考=%d", trial, step, got, refEnergy)
			}
			if pool.UpdatedAt != refUpdatedAt {
				t.Fatalf("trial=%d step=%d：UpdatedAt=%d 参考=%d", trial, step, pool.UpdatedAt, refUpdatedAt)
			}
			if pool.Energy < 0 || pool.Energy > pool.Max {
				t.Fatalf("trial=%d step=%d：体力越界 %d (Max=%d)", trial, step, pool.Energy, pool.Max)
			}
		}
	}
}

func TestRegenScale(t *testing.T) {
	const pools = 100000
	const settlements = 100000
	slice := make([]Pool, pools)
	for i := range slice {
		slice[i] = Pool{Energy: 5, Max: 20, RatePerMin: 1, UpdatedAt: int64(i)}
	}
	start := time.Now()
	for i := 0; i < settlements; i++ {
		slice[i%pools].Regen(int64(i)*60 + 60)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("10 万池 10 万次结算应在 1 秒内完成，实际 %v", elapsed)
	}
}
