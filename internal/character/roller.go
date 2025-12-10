package character

import (
	"crypto/rand"
	"math/big"
	"sort"
)

// RollMethod represents the method used to generate ability scores
type RollMethod int

const (
	Roll4d6DropLowest RollMethod = iota
	StandardArray
	PointBuy
)

// StandardArrayValues are the values for the standard array
var StandardArrayValues = []int{15, 14, 13, 12, 10, 8}

// PointBuyCosts maps ability score to point cost
var PointBuyCosts = map[int]int{
	8:  0,
	9:  1,
	10: 2,
	11: 3,
	12: 4,
	13: 5,
	14: 7,
	15: 9,
}

// PointBuyTotal is the number of points available
const PointBuyTotal = 27

// PointBuyMin is the minimum score in point buy
const PointBuyMin = 8

// PointBuyMax is the maximum score in point buy
const PointBuyMax = 15

// Roll represents a single die roll result
type Roll struct {
	Values  []int
	Dropped int
	Total   int
}

// AbilityRolls represents all 6 ability score rolls
type AbilityRolls struct {
	Rolls  []Roll
	Totals []int
}

// rollDie rolls a single die with n sides using crypto/rand
func rollDie(sides int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		// Fallback to a simple value if crypto/rand fails
		return 1
	}
	return int(n.Int64()) + 1
}

// Roll4d6 rolls 4d6 and drops the lowest
func Roll4d6() Roll {
	dice := make([]int, 4)
	for i := 0; i < 4; i++ {
		dice[i] = rollDie(6)
	}

	// Sort to find the lowest
	sorted := make([]int, 4)
	copy(sorted, dice)
	sort.Ints(sorted)

	dropped := sorted[0]
	total := sorted[1] + sorted[2] + sorted[3]

	return Roll{
		Values:  dice,
		Dropped: dropped,
		Total:   total,
	}
}

// RollAbilityScores rolls 6 sets of 4d6 drop lowest
func RollAbilityScores() AbilityRolls {
	rolls := make([]Roll, 6)
	totals := make([]int, 6)

	for i := 0; i < 6; i++ {
		rolls[i] = Roll4d6()
		totals[i] = rolls[i].Total
	}

	return AbilityRolls{
		Rolls:  rolls,
		Totals: totals,
	}
}

// GetStandardArray returns a copy of the standard array
func GetStandardArray() []int {
	arr := make([]int, len(StandardArrayValues))
	copy(arr, StandardArrayValues)
	return arr
}

// PointBuyState tracks the current state of point buy allocation
type PointBuyState struct {
	Scores         map[string]int
	PointsRemaining int
}

// NewPointBuyState creates a new point buy state with all scores at 8
func NewPointBuyState() *PointBuyState {
	scores := make(map[string]int)
	for _, ability := range Abilities {
		scores[ability] = PointBuyMin
	}
	return &PointBuyState{
		Scores:         scores,
		PointsRemaining: PointBuyTotal,
	}
}

// CanIncrease checks if an ability can be increased
func (p *PointBuyState) CanIncrease(ability string) bool {
	current := p.Scores[ability]
	if current >= PointBuyMax {
		return false
	}
	cost := PointBuyCosts[current+1] - PointBuyCosts[current]
	return p.PointsRemaining >= cost
}

// CanDecrease checks if an ability can be decreased
func (p *PointBuyState) CanDecrease(ability string) bool {
	return p.Scores[ability] > PointBuyMin
}

// Increase increases an ability score by 1
func (p *PointBuyState) Increase(ability string) bool {
	if !p.CanIncrease(ability) {
		return false
	}
	current := p.Scores[ability]
	cost := PointBuyCosts[current+1] - PointBuyCosts[current]
	p.Scores[ability]++
	p.PointsRemaining -= cost
	return true
}

// Decrease decreases an ability score by 1
func (p *PointBuyState) Decrease(ability string) bool {
	if !p.CanDecrease(ability) {
		return false
	}
	current := p.Scores[ability]
	refund := PointBuyCosts[current] - PointBuyCosts[current-1]
	p.Scores[ability]--
	p.PointsRemaining += refund
	return true
}

// GetScores returns the current scores in order
func (p *PointBuyState) GetScores() []int {
	scores := make([]int, len(Abilities))
	for i, ability := range Abilities {
		scores[i] = p.Scores[ability]
	}
	return scores
}

// RollDice rolls a specified number of dice with given sides
func RollDice(count, sides int) []int {
	results := make([]int, count)
	for i := 0; i < count; i++ {
		results[i] = rollDie(sides)
	}
	return results
}

// RollDiceTotal rolls dice and returns the total
func RollDiceTotal(count, sides int) int {
	total := 0
	for i := 0; i < count; i++ {
		total += rollDie(sides)
	}
	return total
}

// RollD20 rolls a d20
func RollD20() int {
	return rollDie(20)
}

// RollWithAdvantage rolls 2d20 and takes the higher
func RollWithAdvantage() (int, int, int) {
	r1 := RollD20()
	r2 := RollD20()
	result := r1
	if r2 > r1 {
		result = r2
	}
	return result, r1, r2
}

// RollWithDisadvantage rolls 2d20 and takes the lower
func RollWithDisadvantage() (int, int, int) {
	r1 := RollD20()
	r2 := RollD20()
	result := r1
	if r2 < r1 {
		result = r2
	}
	return result, r1, r2
}
