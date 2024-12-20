package core

import (
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// Plan represents a parameter update plan that will take effect at specified height
type Plan struct {
	ID     uint64 // Unique identifier for the plan
	Height uint64 // Block height when this plan takes effect

	// Parameters that can be updated
	Difficulty     *big.Int
	TargetPowRatio *common.Rational
	MinPowGas      uint64
	MaxPowGas      uint64
	InitialGas     uint64
	MinPrice       *big.Int
	MaxPrice       *big.Int
	Alpha          *common.Rational
	Fmin           *common.Rational
	Fmax           *common.Rational
	Kp             *common.Rational
	Ki             *common.Rational
}

// PlanPool manages all parameter update plans
type PlanPool struct {
	mu        sync.RWMutex
	plans     map[uint64][]*Plan // Height -> Plans mapping
	nextID    uint64             // For generating unique plan IDs
	minHeight uint64             // Track minimum height
}

// NewPlanPool creates a new plan pool
func NewPlanPool() *PlanPool {
	return &PlanPool{
		plans:     make(map[uint64][]*Plan),
		minHeight: ^uint64(0), // Initialize to max uint64
	}
}

// AddPlan adds a new plan to the pool
func (pp *PlanPool) AddPlan(plan *Plan) uint64 {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Generate and set plan ID
	pp.nextID++
	plan.ID = pp.nextID

	// Add plan to the pool
	pp.plans[plan.Height] = append(pp.plans[plan.Height], plan)

	// Update minimum height if necessary
	if plan.Height < pp.minHeight {
		pp.minHeight = plan.Height
	}

	return plan.ID
}

// RemovePlan removes a plan by its ID
func (pp *PlanPool) RemovePlan(id uint64) bool {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for height, plans := range pp.plans {
		for i, plan := range plans {
			if plan.ID == id {
				// Remove the plan
				pp.plans[height] = append(plans[:i], plans[i+1:]...)

				// If this height has no more plans, delete it
				if len(pp.plans[height]) == 0 {
					delete(pp.plans, height)

					// Update minimum height if necessary
					if height == pp.minHeight {
						pp.updateMinHeight()
					}
				}
				return true
			}
		}
	}
	return false
}

// updateMinHeight updates the minimum height by scanning all heights
func (pp *PlanPool) updateMinHeight() {
	pp.minHeight = ^uint64(0) // Reset to max uint64
	for height := range pp.plans {
		if height < pp.minHeight {
			pp.minHeight = height
		}
	}
}

// GetMinHeight returns the minimum height of all plans
func (pp *PlanPool) GetMinHeight() uint64 {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	if pp.minHeight == ^uint64(0) {
		return 0 // Return 0 if no plans exist
	}
	return pp.minHeight
}

// HasPendingPlans checks if there are any plans with height <= current
func (pp *PlanPool) HasPendingPlans(currentHeight uint64) bool {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	return pp.minHeight <= currentHeight
}

// GetPlansByHeight returns all plans scheduled for the given height
func (pp *PlanPool) GetPlansByHeight(height uint64) []*Plan {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	if plans, exists := pp.plans[height]; exists {
		// Return a copy to prevent external modifications
		result := make([]*Plan, len(plans))
		copy(result, plans)
		return result
	}
	return nil
}

// GetPlanByID returns a specific plan by its ID
func (pp *PlanPool) GetPlanByID(id uint64) *Plan {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	for _, plans := range pp.plans {
		for _, plan := range plans {
			if plan.ID == id {
				return plan
			}
		}
	}
	return nil
}

type ParamUpdater struct {
	update func(interface{})
	value  func(p *Plan) interface{}
}

// MergePlans merges multiple plans and returns a new plan with the latest values
func (pp *PlanPool) MergePlans(height uint64) *Plan {
	plans := pp.GetPlansByHeight(height)
	if len(plans) == 0 {
		return nil
	}

	// Sort plans by ID in ascending order
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].ID < plans[j].ID
	})

	// Create merged plan with the same height
	mergedPlan := &Plan{Height: height}
	lastPlan := plans[len(plans)-1]

	// Take the latest non-nil/non-zero value for each parameter
	if lastPlan.Difficulty != nil {
		mergedPlan.Difficulty = lastPlan.Difficulty
	}
	if lastPlan.TargetPowRatio != nil {
		mergedPlan.TargetPowRatio = lastPlan.TargetPowRatio
	}
	if lastPlan.MinPowGas != 0 {
		mergedPlan.MinPowGas = lastPlan.MinPowGas
	}
	if lastPlan.MaxPowGas != 0 {
		mergedPlan.MaxPowGas = lastPlan.MaxPowGas
	}
	if lastPlan.InitialGas != 0 {
		mergedPlan.InitialGas = lastPlan.InitialGas
	}
	if lastPlan.MinPrice != nil {
		mergedPlan.MinPrice = lastPlan.MinPrice
	}
	if lastPlan.MaxPrice != nil {
		mergedPlan.MaxPrice = lastPlan.MaxPrice
	}
	if lastPlan.Alpha != nil {
		mergedPlan.Alpha = lastPlan.Alpha
	}
	if lastPlan.Fmin != nil {
		mergedPlan.Fmin = lastPlan.Fmin
	}
	if lastPlan.Fmax != nil {
		mergedPlan.Fmax = lastPlan.Fmax
	}
	if lastPlan.Kp != nil {
		mergedPlan.Kp = lastPlan.Kp
	}
	if lastPlan.Ki != nil {
		mergedPlan.Ki = lastPlan.Ki
	}

	return mergedPlan
}

// UpdateParams updates all parameters based on the given plan
func (pp *PlanPool) UpdateParams(plan *Plan) {
	if plan == nil {
		return
	}

	paramUpdaters := map[string]ParamUpdater{
		"Difficulty": {
			update: func(v interface{}) { params.InitialDifficulty = v.(*big.Int) },
			value:  func(p *Plan) interface{} { return p.Difficulty },
		},
		"TargetPowRatio": {
			update: func(v interface{}) { params.TargetPowRatio = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.TargetPowRatio },
		},
		"MinPowGas": {
			update: func(v interface{}) { params.MinPowGas = v.(uint64) },
			value:  func(p *Plan) interface{} { return p.MinPowGas },
		},
		"MaxPowGas": {
			update: func(v interface{}) { params.MaxPowGas = v.(uint64) },
			value:  func(p *Plan) interface{} { return p.MaxPowGas },
		},
		"InitialGas": {
			update: func(v interface{}) { params.InitialGas = v.(uint64) },
			value:  func(p *Plan) interface{} { return p.InitialGas },
		},
		"MinPrice": {
			update: func(v interface{}) { params.MinPrice = v.(*big.Int) },
			value:  func(p *Plan) interface{} { return p.MinPrice },
		},
		"MaxPrice": {
			update: func(v interface{}) { params.MaxPrice = v.(*big.Int) },
			value:  func(p *Plan) interface{} { return p.MaxPrice },
		},
		"Alpha": {
			update: func(v interface{}) { params.Alpha = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.Alpha },
		},
		"Fmin": {
			update: func(v interface{}) { params.Fmin = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.Fmin },
		},
		"Fmax": {
			update: func(v interface{}) { params.Fmax = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.Fmax },
		},
		"Kp": {
			update: func(v interface{}) { params.Kp = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.Kp },
		},
		"Ki": {
			update: func(v interface{}) { params.Ki = v.(*common.Rational) },
			value:  func(p *Plan) interface{} { return p.Ki },
		},
	}

	for name, updater := range paramUpdaters {
		if value := updater.value(plan); value != nil && value != 0 {
			updater.update(value)
			log.Info("Parameter updated", "name", name, "value", value)
		}
	}
}
