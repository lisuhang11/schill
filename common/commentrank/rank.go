package commentrank

import (
	"math"
	"time"
)

const (
	// zScore is the 95% confidence z-value for Wilson score interval.
	zScore = 1.96
	// gravity is the time decay exponent. Larger = faster decay.
	gravity = 1.8
	// replyWeight controls how much each reply boosts the score (log-scaled).
	replyWeight = 0.05
	// halfLifeHours is the "half-life" in hours — after this many hours a
	// comment with fixed votes loses half its score weight (approximately).
	halfLifeHours = 2.0
)

// Score computes a unified ranking score for a comment.
//
// The score combines three components:
//  1. Wilson confidence interval lower bound — penalizes items with few votes.
//  2. Gravity time decay — newer comments get a boost.
//  3. Reply count bonus — log-scaled to avoid domination by reply-heavy threads.
//
// Parameters:
//   - likeCount, dislikeCount: positive vote counts
//   - replyCount: number of direct replies
//   - createdAt: comment creation time
func Score(likeCount, dislikeCount, replyCount int64, createdAt time.Time) float64 {
	// Wilson score lower bound: (p + z²/(2n) - z*sqrt((p*(1-p) + z²/(4n))/n)) / (1 + z²/n)
	// where p = up/n, n = up + down
	n := float64(likeCount + dislikeCount)
	if n == 0 {
		// No votes yet — return a small default so new comments don't sink to bottom.
		// They still get time-decay boost.
		wilson := 0.0
		ageHours := math.Max(0, time.Since(createdAt).Hours())
		timeFactor := math.Pow(ageHours+halfLifeHours, -gravity)
		replyBonus := math.Log1p(float64(replyCount)) * replyWeight
		return wilson*timeFactor + replyBonus
	}

	p := float64(likeCount) / n
	z2 := zScore * zScore

	// Wilson lower bound
	nominator := p + z2/(2*n) - zScore*math.Sqrt((p*(1-p)+z2/(4*n))/n)
	denominator := 1 + z2/n
	wilson := nominator / denominator

	// Time decay
	ageHours := math.Max(0, time.Since(createdAt).Hours())
	timeFactor := math.Pow(ageHours+halfLifeHours, -gravity)

	// Reply bonus (log-scaled)
	replyBonus := math.Log1p(float64(replyCount)) * replyWeight

	return wilson*timeFactor + replyBonus
}
