package main

import (
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const (
	maxNameLen = 42

	offsetX          = 0
	offsetY          = 4
	offsetZ          = 8
	offsetGold       = 12
	offsetRespect    = 16
	offsetStrength   = 16
	offsetExperience = 17
	offsetLevel      = 17
	offsetMana       = 18
	offsetHouse      = 19
	offsetGun        = 19
	offsetFamily     = 19
	offsetType       = 19
	offsetHealth     = 20
	offsetNameLength = 21
	offsetName       = 22

	maskStrength   = 0b00001111
	maskLevel      = 0b00001111
	maskMana       = 0b00000011
	maskHouse      = 0b00010000
	maskGun        = 0b00001000
	maskFamily     = 0b00000100
	maskHealth     = 0b00000011
	maskType       = 0b00000011
	maskNameLength = 0b00111111

	shiftManaHigh   = 2
	shiftManaLow    = 6
	shiftHealthHigh = 2
	shiftHealthLow  = 6
	shiftRespect    = 4
	shiftExperience = 4
)

func save[T any](ptr unsafe.Pointer, val T) {
	*(*T)(ptr) = val
}

func load[T any](ptr unsafe.Pointer) T {
	return *(*T)(ptr)
}

type Option func(*GamePerson)

func WithName(name string) func(*GamePerson) {
	return func(person *GamePerson) {
		namelen := len(name)
		if namelen > maxNameLen {
			name, namelen = name[:maxNameLen], maxNameLen
		}
		person.data[offsetNameLength] |= byte(namelen)
		copy(person.data[offsetName:offsetName+maxNameLen], name[:])
	}
}

func WithCoordinates(x, y, z int) func(*GamePerson) {
	return func(person *GamePerson) {
		save[int32](unsafe.Pointer(&person.data[offsetX]), int32(x))
		save[int32](unsafe.Pointer(&person.data[offsetY]), int32(y))
		save[int32](unsafe.Pointer(&person.data[offsetZ]), int32(z))
	}
}

func WithGold(gold int) func(*GamePerson) {
	return func(person *GamePerson) {
		if gold < 0 {
			gold = 0 // fix overflow
		}
		save[uint32](unsafe.Pointer(&person.data[offsetGold]), uint32(gold))
	}
}

func WithMana(mana int) func(*GamePerson) {
	return func(person *GamePerson) {
		if mana < 0 {
			mana = 0 // fix overflow
		}
		person.data[offsetMana] |= byte(mana >> shiftManaHigh)
		person.data[offsetMana+1] |= byte(mana&maskMana) << shiftManaLow
	}
}

func WithHealth(health int) func(*GamePerson) {
	return func(person *GamePerson) {
		if health < 0 {
			health = 0 // fix overflow
		}
		person.data[offsetHealth] |= byte(health >> shiftHealthHigh)
		person.data[offsetHealth+1] |= byte(health&maskHealth) << shiftHealthLow
	}
}

func WithRespect(respect int) func(*GamePerson) {
	return func(person *GamePerson) {
		if respect < 0 {
			respect = 0 // fix overflow
		}
		person.data[offsetRespect] |= byte(respect) << shiftRespect
	}
}

func WithStrength(strength int) func(*GamePerson) {
	return func(person *GamePerson) {
		if strength < 0 {
			strength = 0 // fix overflow
		}
		person.data[offsetStrength] |= byte(strength) & maskStrength
	}
}

func WithExperience(experience int) func(*GamePerson) {
	return func(person *GamePerson) {
		if experience < 0 {
			experience = 0 // fix overflow
		}
		person.data[offsetExperience] |= byte(experience) << shiftExperience
	}
}

func WithLevel(level int) func(*GamePerson) {
	return func(person *GamePerson) {
		if level < 0 {
			level = 0 // fix overflow
		}
		person.data[offsetLevel] |= byte(level) & maskLevel
	}
}

func WithHouse() func(*GamePerson) {
	return func(person *GamePerson) {
		person.data[offsetHouse] |= maskHouse
	}
}

func WithGun() func(*GamePerson) {
	return func(person *GamePerson) {
		person.data[offsetGun] |= maskGun
	}
}

func WithFamily() func(*GamePerson) {
	return func(person *GamePerson) {
		person.data[offsetFamily] |= maskFamily
	}
}

func WithType(personType int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.data[offsetType] |= byte(personType & maskType)
	}
}

const (
	BuilderGamePersonType = iota
	BlacksmithGamePersonType
	WarriorGamePersonType
)

var _ uintptr = unsafe.Sizeof(GamePerson{}) - 64
var _ uintptr = 64 - unsafe.Sizeof(GamePerson{})

type GamePerson struct {
	data [64]byte
}

func NewGamePerson(opts ...Option) GamePerson {
	p := GamePerson{}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

func (p *GamePerson) Name() string {
	namelen := p.data[offsetNameLength] & maskNameLength
	return unsafe.String(&p.data[offsetName], namelen)
}

func (p *GamePerson) X() int {
	return int(load[int32](unsafe.Pointer(&p.data[offsetX])))
}

func (p *GamePerson) Y() int {
	return int(load[int32](unsafe.Pointer(&p.data[offsetY])))
}

func (p *GamePerson) Z() int {
	return int(load[int32](unsafe.Pointer(&p.data[offsetZ])))
}

func (p *GamePerson) Gold() int {
	return int(load[uint32](unsafe.Pointer(&p.data[offsetGold])))
}

func (p *GamePerson) Mana() int {
	return int(p.data[offsetMana])<<shiftManaHigh | int(p.data[offsetMana+1]>>shiftManaLow)
}

func (p *GamePerson) Health() int {
	return int(p.data[offsetHealth])<<shiftHealthHigh | int(p.data[offsetHealth+1]>>shiftHealthLow)
}

func (p *GamePerson) Respect() int {
	return int(p.data[offsetRespect] >> shiftRespect)
}

func (p *GamePerson) Strength() int {
	return int(p.data[offsetStrength] & maskStrength)
}

func (p *GamePerson) Experience() int {
	return int(p.data[offsetExperience] >> shiftExperience)
}

func (p *GamePerson) Level() int {
	return int(p.data[offsetLevel] & maskLevel)
}

func (p *GamePerson) HasHouse() bool {
	return p.data[offsetHouse]&maskHouse != 0
}

func (p *GamePerson) HasGun() bool {
	return p.data[offsetGun]&maskGun != 0
}

func (p *GamePerson) HasFamily() bool {
	return p.data[offsetFamily]&maskFamily != 0
}

func (p *GamePerson) Type() int {
	return int(p.data[offsetType] & maskType)
}

func TestGamePerson(t *testing.T) {
	assert.LessOrEqual(t, unsafe.Sizeof(GamePerson{}), uintptr(64))

	const x, y, z = math.MinInt32, math.MaxInt32, 0
	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const personType = BuilderGamePersonType
	const gold = math.MaxInt32
	const mana = 1000
	const health = 1000
	const respect = 10
	const strength = 10
	const experience = 10
	const level = 10

	options := []Option{
		WithName(name),
		WithCoordinates(x, y, z),
		WithGold(gold),
		WithMana(mana),
		WithHealth(health),
		WithRespect(respect),
		WithStrength(strength),
		WithExperience(experience),
		WithLevel(level),
		WithHouse(),
		WithFamily(),
		WithType(personType),
	}

	person := NewGamePerson(options...)
	assert.Equal(t, name, person.Name())
	assert.Equal(t, x, person.X())
	assert.Equal(t, y, person.Y())
	assert.Equal(t, z, person.Z())
	assert.Equal(t, gold, person.Gold())
	assert.Equal(t, mana, person.Mana())
	assert.Equal(t, health, person.Health())
	assert.Equal(t, respect, person.Respect())
	assert.Equal(t, strength, person.Strength())
	assert.Equal(t, experience, person.Experience())
	assert.Equal(t, level, person.Level())
	assert.True(t, person.HasHouse())
	assert.True(t, person.HasFamily())
	assert.False(t, person.HasGun())
	assert.Equal(t, personType, person.Type())
}

func TestCoordinates(t *testing.T) {
	const maxVal = 2_000_000_000
	const minVal = -2_000_000_000

	p := NewGamePerson(WithCoordinates(maxVal, maxVal, maxVal))
	assert.Equal(t, maxVal, p.X())
	assert.Equal(t, maxVal, p.Y())
	assert.Equal(t, maxVal, p.Z())

	p = NewGamePerson(WithCoordinates(minVal, minVal, minVal))
	assert.Equal(t, minVal, p.X())
	assert.Equal(t, minVal, p.Y())
	assert.Equal(t, minVal, p.Z())
}

func TestGold(t *testing.T) {
	const maxVal = 2_000_000_000
	const minVal = 0

	p := NewGamePerson(WithGold(maxVal))
	assert.Equal(t, maxVal, p.Gold())

	p = NewGamePerson(WithGold(minVal))
	assert.Equal(t, minVal, p.Gold())

	p = NewGamePerson(WithGold(-1))
	assert.Equal(t, 0, p.Gold())
}

func TestRespect(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithRespect(maxVal))
	assert.Equal(t, maxVal, p.Respect())

	p = NewGamePerson(WithRespect(minVal))
	assert.Equal(t, minVal, p.Respect())

	p = NewGamePerson(WithRespect(-1))
	assert.Equal(t, 0, p.Respect())
}

func TestStrength(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithStrength(maxVal))
	assert.Equal(t, maxVal, p.Strength())

	p = NewGamePerson(WithStrength(minVal))
	assert.Equal(t, minVal, p.Strength())

	p = NewGamePerson(WithStrength(-1))
	assert.Equal(t, 0, p.Strength())
}

func TestRespectWithStrength(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithRespect(maxVal), WithStrength(maxVal))
	assert.Equal(t, maxVal, p.Respect())
	assert.Equal(t, maxVal, p.Strength())

	p = NewGamePerson(WithRespect(maxVal), WithStrength(minVal))
	assert.Equal(t, maxVal, p.Respect())
	assert.Equal(t, minVal, p.Strength())

	p = NewGamePerson(WithRespect(minVal), WithStrength(maxVal))
	assert.Equal(t, minVal, p.Respect())
	assert.Equal(t, maxVal, p.Strength())

	p = NewGamePerson(WithRespect(minVal), WithStrength(minVal))
	assert.Equal(t, minVal, p.Respect())
	assert.Equal(t, minVal, p.Strength())
}

func TestExperience(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithExperience(maxVal))
	assert.Equal(t, maxVal, p.Experience())

	p = NewGamePerson(WithExperience(minVal))
	assert.Equal(t, minVal, p.Experience())

	p = NewGamePerson(WithExperience(-1))
	assert.Equal(t, 0, p.Experience())
}

func TestLevel(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithLevel(maxVal))
	assert.Equal(t, maxVal, p.Level())

	p = NewGamePerson(WithLevel(minVal))
	assert.Equal(t, minVal, p.Level())

	p = NewGamePerson(WithLevel(-1))
	assert.Equal(t, 0, p.Level())
}

func TestExperienceWithLevel(t *testing.T) {
	const maxVal = 10
	const minVal = 0

	p := NewGamePerson(WithExperience(maxVal), WithLevel(maxVal))
	assert.Equal(t, maxVal, p.Experience())
	assert.Equal(t, maxVal, p.Level())

	p = NewGamePerson(WithExperience(maxVal), WithLevel(minVal))
	assert.Equal(t, maxVal, p.Experience())
	assert.Equal(t, minVal, p.Level())

	p = NewGamePerson(WithExperience(minVal), WithLevel(maxVal))
	assert.Equal(t, minVal, p.Experience())
	assert.Equal(t, maxVal, p.Level())

	p = NewGamePerson(WithExperience(minVal), WithLevel(minVal))
	assert.Equal(t, minVal, p.Experience())
	assert.Equal(t, minVal, p.Level())
}

func TestMana(t *testing.T) {
	const maxVal = 1000
	const minVal = 0

	p := NewGamePerson(WithMana(maxVal))
	assert.Equal(t, maxVal, p.Mana())

	p = NewGamePerson(WithMana(minVal))
	assert.Equal(t, minVal, p.Mana())

	p = NewGamePerson(WithMana(-1))
	assert.Equal(t, 0, p.Mana())
}

func TestHouse(t *testing.T) {
	p := NewGamePerson(WithHouse())
	assert.True(t, p.HasHouse())

	p = NewGamePerson()
	assert.False(t, p.HasHouse())
}

func TestGun(t *testing.T) {
	p := NewGamePerson(WithGun())
	assert.True(t, p.HasGun())

	p = NewGamePerson()
	assert.False(t, p.HasGun())
}

func TestFamily(t *testing.T) {
	p := NewGamePerson(WithFamily())
	assert.True(t, p.HasFamily())

	p = NewGamePerson()
	assert.False(t, p.HasFamily())
}

func TestType(t *testing.T) {
	p := NewGamePerson(WithType(BuilderGamePersonType))
	assert.Equal(t, BuilderGamePersonType, p.Type())

	p = NewGamePerson(WithType(BlacksmithGamePersonType))
	assert.Equal(t, BlacksmithGamePersonType, p.Type())

	p = NewGamePerson(WithType(WarriorGamePersonType))
	assert.Equal(t, WarriorGamePersonType, p.Type())
}

func TestManaWithHouseWithGunWithFamilyWithType(t *testing.T) {
	const maxVal = 1000
	const minVal = 0

	p := NewGamePerson(
		WithMana(maxVal),
		WithHouse(),
		WithGun(),
		WithFamily(),
		WithType(BuilderGamePersonType),
	)
	assert.Equal(t, maxVal, p.Mana())
	assert.True(t, p.HasHouse())
	assert.True(t, p.HasGun())
	assert.True(t, p.HasFamily())
	assert.Equal(t, BuilderGamePersonType, p.Type())
}

func TestHealth(t *testing.T) {
	const maxVal = 1000
	const minVal = 0

	p := NewGamePerson(WithHealth(maxVal))
	assert.Equal(t, maxVal, p.Health())

	p = NewGamePerson(WithHealth(minVal))
	assert.Equal(t, minVal, p.Health())

	p = NewGamePerson(WithHealth(-1))
	assert.Equal(t, 0, p.Health())
}

func TestName(t *testing.T) {
	p := NewGamePerson(WithName("John"))
	assert.Equal(t, "John", p.Name())
	assert.Equal(t, 4, len(p.Name()))

	p = NewGamePerson(WithName("aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc#"))
	assert.Equal(t, "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc", p.Name())
	assert.Equal(t, 42, len(p.Name()))
}

func TestHealthWithNameLength(t *testing.T) {
	const maxVal = 1000
	const minVal = 0

	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const namelen = 42

	p := NewGamePerson(WithName(name), WithHealth(maxVal))
	assert.Equal(t, name, p.Name())
	assert.Equal(t, namelen, len(p.Name()))
	assert.Equal(t, maxVal, p.Health())

	p = NewGamePerson(WithName(name), WithHealth(minVal))
	assert.Equal(t, name, p.Name())
	assert.Equal(t, namelen, len(p.Name()))
	assert.Equal(t, minVal, p.Health())
}
