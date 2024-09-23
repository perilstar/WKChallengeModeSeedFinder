package main

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/crazy3lf/colorconv"
)

const ( // constants, defined in different
	Math_TAU           = 6.2831853071795864769252867666
	PCG_DEFAULT_INC_64 = 1442695040888963407
)

var itemCategories = []string{"speed", "fireRate", "multiShot", "wallPunch", "splashDamage", "piercing", "freezing", "infection"}

var itemCosts = map[string]float64{
	"speed":        1.0,
	"fireRate":     2.8,
	"multiShot":    3.3,
	"wallPunch":    1.25,
	"splashDamage": 2.0,
	"piercing":     2.4,
	"freezing":     1.5,
	"infection":    2.15,
}

var charList = [6]string{"basic", "mage", "laser", "melee", "pointer", "swarm"}

var pcg pcg32_random_t // rename to rng maybe?

var p_inc uint64 = 1442695040888963407
var p_seed uint64 = 12047754176567800795 // required for godot, it's the default seed it uses for when no seed is set
var current_inc uint64 = p_inc
var current_seed uint64 = 0

// windowkill

func main() {
	Set_seed(uint64(3823837572363))

	/* 	test seed, this seed should print:

	   	seed: 3823837572363
	   	char: laser
	   	abilityChar: mage
	   	abilityLevel: 1
	   	itemCategories: [wallPunch speed infection splashDamage multiShot fireRate piercing freezing]
	   	itemCounts: map[fireRate:6 freezing:26 infection:7 multiShot:3 piercing:9 speed:0 splashDamage:0 wallPunch:0]
	   	startTime: 641.089106798172
	   	colorState: 1
	   	color: 1 0.75686276 0.75686276 1
	*/

	// fmt.Println("state:", get_state())
	fmt.Println("seed:", Get_seed())

	// intensity determines basis for other rolls
	var intensity = Randf_range(0.20, 1.0)

	var char = charList[int(Randi())%len(charList)]
	var abilityChar = charList[int(Randi())%len(charList)]
	var abilityLevel = 1.0 + math.Round(run(Randf(), 1.5/(1.0+intensity), 1.0, 0.0)*6)

	var itemCount float64 = float64(len(itemCategories))
	// points determine item layout
	var points = 0.66 * itemCount * Randf_range(0.5, 1.5) * (1.0 + 4.0*math.Pow(intensity, 1.5))

	var itemDistSteepness = Randf_range(-0.5, 2.0)
	var itemDistArea = 1.0 / (1.0 + math.Pow(2, 0.98*itemDistSteepness))

	// windowkill uses the ""global"" randomness and shuffle()
	// instead of that this saves the state (and inc)
	// then sets the seed to the current seed (because setting the current seed advances the state 2 times im pretty sure, although doesnt really matter)
	// and after calling shuffle() set the state (and inc) back to what it was before directly

	var oldstate = pcg.state
	var oldInc = pcg.inc

	Set_seed(Get_seed())
	shuffle(itemCategories)

	pcg.inc = oldInc
	pcg.state = oldstate

	// chance to move offensive upgrades closer to end if not already

	if Randf() < intensity {
		multishotIdx := -1
		for i, category := range itemCategories {
			if category == "multiShot" {
				multishotIdx = i
				break
			}
		}

		if multishotIdx != -1 {
			itemCategories = append(itemCategories[:multishotIdx], itemCategories[multishotIdx+1:]...)
		}
		insertIdx := int32(itemCount) - 1 - Randi_range(0, 2)
		itemCategories = append(itemCategories[:insertIdx], append([]string{"multiShot"}, itemCategories[insertIdx:]...)...)
	}

	if Randf() < intensity {
		fireRateIdx := -1
		for i, category := range itemCategories {
			if category == "fireRate" {
				fireRateIdx = i
				break
			}
		}

		if fireRateIdx != -1 {
			itemCategories = append(itemCategories[:fireRateIdx], itemCategories[fireRateIdx+1:]...)
		}
		insertIdx := int32(itemCount) - 1 - Randi_range(0, 2)
		itemCategories = append(itemCategories[:insertIdx], append([]string{"fireRate"}, itemCategories[insertIdx:]...)...)
	}

	var itemCounts = make(map[string]int)
	var catMax = float64(itemCount - 1)
	var total = 0
	for i := 0; i < int(itemCount); i++ {
		var item = itemCategories[i]
		var catT = float64(i) / catMax
		var cost = itemCosts[item]
		cost = 1.0 + ((cost - 1.0) / 2.5)
		baseAmount := 0.0

		var special = 0.0
		if i == int(itemCount)-1 {
			special += 4.0 * Randf_range(0.0, float32(math.Pow(intensity, 2.0)))
		}
		amount := math.Max(0.0, 3.0*run(catT, itemDistSteepness, 1.0, 0.0)+3.0*clamp(Randfn(0.0, 0.15), -0.5, 0.5))
		itemCounts[item] = int(clamp(math.Round(baseAmount+amount*((points/cost)/(1.0+5.0*itemDistArea))+special), 0.0, 26.0))
		total += itemCounts[item]
	}

	// balance for offensive upgrades
	intensity = -0.05 + intensity*lerp(0.33, 1.2, smoothCorner((float64(itemCounts["multiShot"])*1.8+float64(itemCounts["fireRate"]))/12.0, 1.0, 1.0, 4.0))

	var finalT = Randfn(float32(math.Pow(intensity, 1.2)), 0.05)
	var startTime = clamp(lerp(60.0*2.0, 60.0*20.0, finalT), 60.0*2.0, 60.0*25.0)

	var r, g, b, _ = colorconv.HSVToRGB(Randf(), Randf(), float64(1.0)) // color
	var colorState = Randi_range(0, 2)

	fmt.Println("char:", char)
	fmt.Println("abilityChar:", abilityChar)
	fmt.Println("abilityLevel:", abilityLevel)
	fmt.Println("itemCategories:", itemCategories)
	fmt.Println("itemCounts:", itemCounts)
	fmt.Println("startTime:", startTime)
	fmt.Println("colorState:", colorState)
	fmt.Println("color:", float32(r)/255, float32(g)/255, float32(b)/255, 1.0) // figure out how to pack r g and b into a single `color` variable
}

// pcg.h

type pcg32_random_t struct {
	state uint64
	inc   uint64
}

// pcg.cpp

func pcg32_random_r(rng *pcg32_random_t) uint32 {
	var oldstate uint64 = rng.state
	rng.state = (oldstate * 6364136223846793005) + (rng.inc | 1)
	var xorshifted uint32 = uint32(((oldstate >> uint64(18)) ^ oldstate) >> uint64(27))
	var rot uint32 = uint32(oldstate >> uint64(59))
	return (xorshifted >> rot) | (xorshifted << ((-rot) & 31))
}

func pcg32_srandom_r(rng *pcg32_random_t, initstate uint64, initseq uint64) {
	rng.state = uint64(0)
	rng.inc = (initseq << 1) | 1
	pcg32_random_r(rng)
	rng.state += initstate
	pcg32_random_r(rng)
}

func pcg32_boundedrand_r(rng *pcg32_random_t, bound uint32) uint32 {
	threshold := -bound % bound
	for {
		r := pcg32_random_r(rng)
		if r >= threshold {
			return r % bound
		}
	}
}

// random_pcg.cpp

func randomf32(p_from float32, p_to float32) float32 { // `random()` float version
	return randf32()*(p_to-p_from) + p_from
}

func randomi(p_from int, p_to int) int { // `random()` int version
	if p_from == p_to {
		return p_from
	}
	bounds := uint32(int(math.Abs(float64(p_from-p_to))) + 1)
	randomValue := int(randbound(bounds))
	if p_from < p_to {
		return p_from + randomValue
	}
	return p_to + randomValue
}

// random_pcg.h

func randbound(bounds uint32) uint32 { // rand() with bounds
	return pcg32_boundedrand_r(&pcg, bounds)
}

func rand() uint32 { // normal rand
	return pcg32_random_r(&pcg)
}

func randf32() float32 {
	var proto_exp_offset uint32 = rand()
	if proto_exp_offset == 0 {
		return 0
	}
	return float32(math.Ldexp(float64(rand()|0x80000001), -32-bits.LeadingZeros32(proto_exp_offset)))
}

// random_number_generator.h

func Set_seed(p_seed uint64) {
	current_seed = p_seed
	pcg32_srandom_r(&pcg, current_seed, current_inc)
}
func Get_seed() uint64 { return current_seed }

func Randi() uint32 {
	return rand()
}

func Randf() float64 {
	var proto_exp_offset uint32 = rand()
	if proto_exp_offset == 0 {
		return 0
	}
	return float64(float32(math.Ldexp(float64(rand()|0x80000001), -32-bits.LeadingZeros32(proto_exp_offset))))
}

func Randf_range(p_from float32, p_to float32) float64 {
	return float64(float32(float64(randomf32(p_from, p_to))))
}

func Randfn(p_mean float32, p_deviation float32) float64 {
	var temp float32 = randf32()
	if temp < 0.00001 {
		temp += 0.00001 // this is what CMP_EPSILON is defined as
	}
	return float64(p_mean + p_deviation*(float32(math.Cos(Math_TAU*float64(randf32()))*math.Sqrt(-2.0*math.Log(float64(temp))))))
}

func Randi_range(p_from int, p_to int) int32 {
	return int32(randomi(p_from, p_to))
}

// helper functions

func pinch(v float64) float64 { // function run() uses
	if v < 0.5 {
		return -v * v
	}
	return v * v
}

func run(x, a, b, c float64) float64 { // TorCurve.run() in godot
	c = pinch(c)
	x = math.Max(0, math.Min(1, x))

	const eps = 0.00001
	s := math.Exp(a)
	s2 := 1.0 / (s + eps)
	t := math.Max(0, math.Min(1, b))
	u := c

	var res, c1, c2, c3 float64

	if x < t {
		c1 = (t * x) / (x + s*(t-x) + eps)
		c2 = t - math.Pow(1/(t+eps), s2-1)*math.Pow(math.Abs(x-t), s2)
		c3 = math.Pow(1/(t+eps), s-1) * math.Pow(x, s)
	} else {
		c1 = (1-t)*(x-1)/(1-x-s*(t-x)+eps) + 1
		c2 = math.Pow(1/((1-t)+eps), s2-1)*math.Pow(math.Abs(x-t), s2) + t
		c3 = 1 - math.Pow(1/((1-t)+eps), s-1)*math.Pow(1-x, s)
	}

	if u <= 0 {
		res = (-u)*c2 + (1+u)*c1
	} else {
		res = (u)*c3 + (1-u)*c1
	}

	return res
}

func smoothCorner(x, m, l, s float64) float64 { // TorCurve.smoothCorner in godot
	s1 := math.Pow(s/10.0, 2.0)
	return 0.5 * ((l*x + m*(1.0+s1)) - math.Sqrt(math.Pow(math.Abs(l*x-m*(1.0-s1)), 2.0)+4.0*m*m*s1))
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func clamp(m_a, m_min, m_max float64) float64 {
	if m_a < m_min {
		return m_min
	} else if m_a > m_max {
		return m_max
	}
	return m_a
}

func shuffle(arr []string) {
	n := len(arr)
	if n <= 1 {
		return
	}
	for i := n - 1; i > 0; i-- {
		j := randbound(uint32(i + 1))
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// end of helper functions

// unused functions

// random_pcg.h
//
// func randd() float64 {
// 	proto_exp_offset := rand()
// 	if proto_exp_offset == 0 {
// 		return 0
// 	}
// 	var significand uint64 = (uint64(rand()) << 32) | uint64(rand()) | 0x8000000000000001
// 	return math.Ldexp(float64(significand), -64-bits.LeadingZeros32(proto_exp_offset))
// }
//
// func seed(p_seed uint64) {
// 	current_seed = p_seed
// 	pcg32_srandom_r(&pcg, current_seed, current_inc)
// }
// func randfn64(p_mean float64, p_deviation float64) float64 {
// 	var temp float64 = randd()
// 	if temp < 0.00001 {
// 		temp += 0.00001 // this is what CMP_EPSILON is defined as
// 	}
// 	// Math_TAU is defined as 6.2831853071795864769252867666
// 	Math_TAU := 6.2831853071795864769252867666
// 	return float64(float32(p_mean + p_deviation*(math.Cos(Math_TAU*randd())*math.Sqrt(-2.0*math.Log(temp)))))
// }

// random_number_generator.h
// func set_state(p_state uint64) { pcg.state = p_state } // this one is just unused in my code, but is something the RandomNumberGenerator object uses
// func get_state() uint64        { return pcg.state }
//
// func Randomize() { // required for godot, but techincally will never be used since it just randomises, can only really be used for seeing which random numbers are more likely than others
// 	seed((uint64(time.Now().Unix()+time.Now().UnixNano()/1000)*pcg.state + PCG_DEFAULT_INC_64))
// }

// random_pcg.cpp
// func randomf64(p_from float64, p_to float64) float64 {
// 	return randd()*(p_to-p_from) + p_from
// }
