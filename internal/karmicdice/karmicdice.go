package karmicdice

import (
	"fmt"
	"math"
	"math/rand/v2"
)

// weight stores the karmic balance. This variable persists across calls to karmicDice.
// A positive value increases the chance of success on the next roll.
// A negative value decreases it.
var weight float64 = 0.0

// Int performs a d20 roll, adjusted by a persistent "karmic" weight.
// The outcome of the roll then modifies the weight for future rolls.
func Int(difficulty int) int {
	// A scaling factor for how much the weight changes.
	// A smaller value means the karma adjusts more slowly.
	const karmicFactor = 0.2

	// 1. Perform a standard 1-20 roll.
	// In modern Go (1.20+), the global rand is automatically seeded.
	baseRoll := rand.IntN(20) + 1

	// 2. Calculate the final roll by applying the current karmic weight.
	// We round the result to get a whole number.
	adjustedRoll := int(math.Round(float64(baseRoll) + weight))

	// 3. Compare the un-adjusted baseRoll to the difficulty to update the weight.
	// This feels more "pure": the weight affects the outcome, not the luck itself.
	if baseRoll < difficulty {
		// FAILURE: The roll was too low.
		// Add to the weight to help the next roll.
		// The amount added is proportional to how badly the roll failed.
		// Formula: weight += |baseRoll - difficulty| * karmicFactor
		difference := float64(difficulty - baseRoll)
		weightChange := difference * karmicFactor
		weight += weightChange

		fmt.Printf(
			"Roll: %2d (Fail) -> Adjusted: %2d. Weight changes by +%.2f to %.2f\n",
			baseRoll,
			adjustedRoll,
			weightChange,
			weight,
		)
	} else {
		// SUCCESS: The roll met or beat the difficulty.
		// Subtract from the weight to balance out the good luck.
		// The amount is proportional to how much the roll succeeded by.
		// Formula: weight -= (baseRoll - difficulty) * karmicFactor
		difference := float64(baseRoll - difficulty)
		weightChange := difference * karmicFactor
		weight -= weightChange

		fmt.Printf(
			"Roll: %2d (Pass) -> Adjusted: %2d. Weight changes by -%.2f to %.2f\n",
			baseRoll,
			adjustedRoll,
			weightChange,
			weight,
		)
	}

	// 4. Return the final, karmically-adjusted roll value.
	return adjustedRoll
}
