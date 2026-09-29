// Package energy 是体力池：按时间回复、消耗与上限。
package energy

import (
	"errors"
	"time"
)

// ErrEnergy 表示体力不足。
var ErrEnergy = errors.New("energy: 体力不足")

// Pool 是体力池。
type Pool struct {
	Energy     int
	Max        int
	RatePerMin int
	UpdatedAt  int64
}

// Regen 按服务器时间 now 结算体力回复。
// 回拨不受理：now 早于 UpdatedAt 时按 UpdatedAt 结算，体力不变也不倒扣；
// 服务器日期跨天时体力重置到 Max（客户端时间不参与判断）；
// 结算后体力截断到 [0, Max]，UpdatedAt 单调不减，重复结算幂等。
func (p *Pool) Regen(now int64) int {
	effective := now
	if effective < p.UpdatedAt {
		effective = p.UpdatedAt
	}
	if dayChanged(p.UpdatedAt, effective) {
		p.Energy = p.Max
	} else {
		minutes := (effective - p.UpdatedAt) / 60
		p.Energy += int(minutes) * p.RatePerMin
	}
	if p.Energy > p.Max {
		p.Energy = p.Max
	}
	if p.Energy < 0 {
		p.Energy = 0
	}
	p.UpdatedAt = effective
	return p.Energy
}

// dayChanged 判断两个服务器时间戳（UTC）是否跨天。
func dayChanged(a, b int64) bool {
	ta := time.Unix(a, 0).UTC()
	tb := time.Unix(b, 0).UTC()
	ya, ma, da := ta.Date()
	yb, mb, db := tb.Date()
	return ya != yb || ma != mb || da != db
}

// Spend 消耗体力。
func (p *Pool) Spend(cost int) error {
	if cost <= 0 || cost > p.Energy {
		return ErrEnergy
	}
	p.Energy -= cost
	return nil
}
