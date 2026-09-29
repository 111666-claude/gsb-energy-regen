package energy

import "testing"

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
