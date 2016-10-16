package sm

import "fmt"

type algorithm struct {
	efMapping                   map[int16]int16
	resetQuality, repeatQuality int16
}

var (
	SM2 = &algorithm{
		efMapping: map[int16]int16{
			0: -80,
			1: -54,
			2: -32,
			3: -14,
			4: 00,
			5: 10,
		},
		resetQuality:  3,
		repeatQuality: 4,
	}

	SM2Mod = &algorithm{
		efMapping: map[int16]int16{
			0: -80,
			1: -30,
			2: 00,
			3: 10,
		},
		resetQuality:  2,
		repeatQuality: 2,
	}
)

func (a *algorithm) nextEF(q, ef int16) int16 {
	efDelta, ok := a.efMapping[q]
	if !ok {
		panic(fmt.Sprintf("Invalid quality supplied %d", q))
	}
	ef += efDelta
	if ef < 130 {
		ef = 130
	}
	return ef
}

func (a *algorithm) nextRepetition(q, repetition int16) int16 {
	if q < a.resetQuality {
		return 1
	} else {
		return repetition + 1
	}
}

func (a *algorithm) nextInterval(q, ef, repetition, interval int16) int16 {
	// If q < repeatQuality we need to repeat it today
	if q < a.repeatQuality {
		return 0
	}

	switch repetition {
	case 1:
		return 1
	case 2:
		return 6
	default:
		nextInterval := int64(interval) * int64(ef)
		interval = int16(nextInterval / 100)
		if nextInterval%100 >= 50 {
			interval++
		}
		return interval
	}
}

// returns (nextRepetition, nextEF, nextInterval)
func (a *algorithm) Calc(q, repetition, ef, interval int16) (int16, int16, int16) {
	return a.nextRepetition(q, repetition), a.nextEF(q, ef), a.nextInterval(q, ef, repetition, interval)
}
