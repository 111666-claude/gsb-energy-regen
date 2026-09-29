// Package energy 是体力池：按时间回复、消耗与上限。
package energy

import "errors"

// ErrEnergy 表示体力不足。
var ErrEnergy = errors.New("energy: 体力不足")

const secondsPerDay = 24 * 60 * 60

// Pool 是体力池。
type Pool struct {
	Energy     int
	Max        int
	RatePerMin int
	UpdatedAt  int64
}

// Regen 按服务端时间 now 结算体力：
// now 早于 UpdatedAt 时按 UpdatedAt 结算（不倒扣）；
// 服务器日期跨天时重置到 Max；回复后截断到 [0, Max]。
func (p *Pool) Regen(now int64) int {
	if now < p.UpdatedAt {
		now = p.UpdatedAt
	}
	if now/secondsPerDay != p.UpdatedAt/secondsPerDay {
		p.Energy = p.Max
	} else {
		p.Energy += int((now-p.UpdatedAt)/60) * p.RatePerMin
	}
	if p.Energy > p.Max {
		p.Energy = p.Max
	}
	if p.Energy < 0 {
		p.Energy = 0
	}
	p.UpdatedAt = now
	return p.Energy
}

// Spend 消耗体力。
func (p *Pool) Spend(cost int) error {
	if cost <= 0 || cost > p.Energy {
		return ErrEnergy
	}
	p.Energy -= cost
	return nil
}
