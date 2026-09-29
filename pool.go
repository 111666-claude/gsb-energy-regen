// Package energy 是体力池：按时间回复、消耗与上限。
// 缺陷：直接采信客户端时间（回拨会倒扣体力）、上限没有截断、跨天重置没做。
package energy

import "errors"

// ErrEnergy 表示体力不足。
var ErrEnergy = errors.New("energy: 体力不足")

// Pool 是体力池。
type Pool struct {
	Energy     int
	Max        int
	RatePerMin int
	UpdatedAt  int64
}

// Regen 按 now 回复体力。缺陷：now 比 UpdatedAt 小时按负数分钟倒扣，且不截断上限。
func (p *Pool) Regen(now int64) int {
	minutes := (now - p.UpdatedAt) / 60
	p.Energy += int(minutes) * p.RatePerMin
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
