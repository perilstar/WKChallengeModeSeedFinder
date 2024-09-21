package main

import (
	"fmt"
	"math"
	"math/bits"
	"time"

	"github.com/crazy3lf/colorconv"
	"github.com/dim13/djb2"
)

// var seed uint64 = 23485210398122134

// var rng = pcg.NewPCG32()
var p_inc uint64 = 1442695040888963407
var p_seed uint64 = 12047754176567800795 // required for godot, it's the default seed it uses for when no seed is set
var current_inc uint64 = p_inc
var current_seed uint64 = 0

// windowkill

var charList = [6]string{"basic", "mage", "laser", "melee", "pointer", "swarm"}

func main() {
	set_seed(uint64(djb2.SumString("eaea")))
	itemData := Array{
		"speed": {
			Cost: 1.0,
			Icon: "res://src/ui/shop/speed.svg",
		},
		"fireRate": {
			Cost: 2.8,
			Icon: "res://src/ui/shop/fireRate.svg",
		},
		"multiShot": {
			Cost: 3.3,
			Icon: "res://src/ui/shop/multiShot.svg",
		},
		"wallPunch": {
			Cost: 1.25,
			Icon: "res://src/ui/shop/wallPunch.svg",
		},
		"splashDamage": {
			Cost: 2.0,
			Icon: "res://src/ui/shop/splashDamage.svg",
		},
		"piercing": {
			Cost: 2.4,
			Icon: "res://src/ui/shop/piercing.svg",
		},
		"freezing": {
			Cost: 1.5,
			Icon: "res://src/ui/shop/freezing.svg",
		},
		"infection": {
			Cost: 2.15,
			Icon: "res://src/ui/shop/infect.svg",
		},
	}

	fmt.Println("Hello, world!")
	fmt.Println("state:", get_state())
	fmt.Println("seed:", get_seed())
	var intensity = randf_range(0.20, 1.0)
	var char = charList[int(randi())%len(charList)]
	var abilityChar = charList[int(randi())%len(charList)]
	var abilityLevel = 1.0 + math.Round(run(randf(), 1.5/(1.0+intensity), 1.0, 0.0)*6)

	itemCategories := []string{"speed", "fireRate", "multiShot", "wallPunch", "splashDamage", "piercing", "freezing", "infection"}
	fmt.Println("itemCategories:", itemCategories)

	var itemCount float64 = float64(len(itemCategories))

	var points = 0.66 * itemCount * randf_range(0.5, 1.5) * (1.0 + 4.0*math.Pow(intensity, 1.5))

	var itemDistSteepness = randf_range(-0.5, 2.0)
	var itemDistArea = 1.0 / (1.0 + math.Pow(2, 0.98*itemDistSteepness))

	var oldstate = pcg.state
	var oldInc = pcg.inc
	seed(get_seed())
	shuffle(itemCategories)
	pcg.inc = oldInc
	pcg.state = oldstate
	if randf() < intensity {
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
		insertIdx := int32(itemCount) - 1 - randi_range(0, 2)
		itemCategories = append(itemCategories[:insertIdx], append([]string{"multiShot"}, itemCategories[insertIdx:]...)...)
	}
	if randf() < intensity {
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
		insertIdx := int32(itemCount) - 1 - randi_range(0, 2)
		itemCategories = append(itemCategories[:insertIdx], append([]string{"fireRate"}, itemCategories[insertIdx:]...)...)
	}
	itemCounts := make(map[string]int)
	catMax := float64(itemCount - 1)
	total := 0
	for i := 0; i < int(itemCount); i++ {
		item := itemCategories[i]
		catT := float64(i) / catMax
		cost := itemData[item].Cost
		cost = 1.0 + ((cost - 1.0) / 2.5)
		baseAmount := 0.0

		special := 0.0
		if i == int(itemCount)-1 {
			special += 4.0 * randf_range(0.0, float32(math.Pow(intensity, 2.0)))
		}
		amount := math.Max(0.0, 3.0*run(catT, itemDistSteepness, 1.0, 0.0)+3.0*clamp(randfn(0.0, 0.15), -0.5, 0.5))
		itemCounts[item] = int(clamp(math.Round(baseAmount+amount*((points/cost)/(1.0+5.0*itemDistArea))+special), 0.0, 26.0))
		total += itemCounts[item]
	}
	intensity = -0.05 + intensity*lerp(0.33, 1.2, smoothCorner((float64(itemCounts["multiShot"])*1.8+float64(itemCounts["fireRate"]))/12.0, 1.0, 1.0, 4.0))
	var finalT = randfn(float32(math.Pow(intensity, 1.2)), 0.05)
	var startTime = clamp(lerp(60.0*2.0, 60.0*20.0, finalT), 60.0*2.0, 60.0*25.0)
	var r, g, b, _ = colorconv.HSVToRGB(randf(), randf(), float64(1.0))
	var colorState = randi_range(0, 2)
	fmt.Println("char:", char)
	fmt.Println("abilityChar:", abilityChar)
	fmt.Println("abilityLevel:", abilityLevel)
	fmt.Println("itemCategories:", itemCategories)
	fmt.Println("itemCounts:", itemCounts)
	fmt.Println("startTime:", startTime)
	fmt.Println("colorState:", colorState)
	fmt.Println(float32(r)/255, float32(g)/255, float32(b)/255, 1.0)
}

func pinch(v float64) float64 {
	if v < 0.5 {
		return -v * v
	}
	return v * v
}

func run(x, a, b, c float64) float64 {
	c = pinch(c)
	x = math.Max(0, math.Min(1, x)) // Clamp input to [0-1]

	const eps = 0.00001 // Protect against div/0
	s := math.Exp(a)    // Could be any exponential like 2^a or 3^a, or just linear
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

	return res // Prevent NaN from Infinity values
}

func smoothCorner(x, m, l, s float64) float64 {
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

type Array map[string]item

type item struct {
	Cost float64
	Icon string
}

func shuffle(arr []string) {
	n := len(arr)
	if n <= 1 {
		return
	}
	for i := n - 1; i > 0; i-- {
		// Generate a random index between 0 and i (inclusive) using your randbound function
		j := randbound(uint32(i + 1))
		// Swap the elements at i and j
		arr[i], arr[j] = arr[j], arr[i]
	}
}

var pcg pcg32_random_t

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

// func randomf64(p_from float64, p_to float64) float64 {
// 	return randd()*(p_to-p_from) + p_from
// }
// this ones in the source code from godot, and might be used for the global random number generator

func randomf32(p_from float32, p_to float32) float32 {
	return randf32()*(p_to-p_from) + p_from
}

func randomi(p_from int, p_to int) int {
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

func randbound(bounds uint32) uint32 {
	return pcg32_boundedrand_r(&pcg, bounds)
}

func rand() uint32 {
	return pcg32_random_r(&pcg)
}

// func randd() float64 {
// 	proto_exp_offset := rand()
// 	if proto_exp_offset == 0 {
// 		return 0
// 	}
// 	var significand uint64 = (uint64(rand()) << 32) | uint64(rand()) | 0x8000000000000001
// 	return math.Ldexp(float64(significand), -64-bits.LeadingZeros32(proto_exp_offset))
// }
//

func randf32() float32 {
	var proto_exp_offset uint32 = rand()
	if proto_exp_offset == 0 {
		return 0
	}
	return float32(math.Ldexp(float64(rand()|0x80000001), -32-bits.LeadingZeros32(proto_exp_offset)))
}

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

func seed(p_seed uint64) {
	current_seed = p_seed
	pcg32_srandom_r(&pcg, current_seed, current_inc)
}

func set_seed(p_seed uint64)   { seed(p_seed) }
func get_seed() uint64         { return current_seed }
func set_state(p_state uint64) { pcg.state = p_state } // this one is just unused in my code, but is something the RandomNumberGenerator object uses
func get_state() uint64        { return pcg.state }

func randomize() { // required for godot, but techincally will never be used since it just randomises, can only really be used for seeing which random numbers are more likely than others
	// PCG_DEFAULT_INC_64 is defined as 1442695040888963407 in pcg.h
	seed((uint64(time.Now().Unix()+time.Now().UnixNano()/1000)*pcg.state + 1442695040888963407))
}

func randi() uint32 {
	return rand()
}
func randf() float64 {
	var proto_exp_offset uint32 = rand()
	if proto_exp_offset == 0 {
		return 0
	}
	return float64(float32(math.Ldexp(float64(rand()|0x80000001), -32-bits.LeadingZeros32(proto_exp_offset))))
}
func randf_range(p_from float32, p_to float32) float64 {
	return float64(float32(float64(randomf32(p_from, p_to))))
}
func randfn(p_mean float32, p_deviation float32) float64 {
	var temp float32 = randf32()
	if temp < 0.00001 {
		temp += 0.00001 // this is what CMP_EPSILON is defined as
	}
	// Math_TAU is defined as 6.2831853071795864769252867666
	Math_TAU := 6.2831853071795864769252867666
	return float64(p_mean + p_deviation*(float32(math.Cos(Math_TAU*float64(randf32()))*math.Sqrt(-2.0*math.Log(float64(temp))))))
}
func randi_range(p_from int, p_to int) int32 {
	return int32(randomi(p_from, p_to))
}
