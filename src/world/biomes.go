package world

import (
	"go-engine/src/pkg"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
)

// Function that calculates the height of a biome
type HeightFunc func(gx, gz int, p2, p3 *perlin.Perlin) float64

type BiomeProperties struct {
	Modifier         HeightFunc
	SurfaceBlock     string
	UndergroundBlock string
}

var BiomeTypes = map[string]BiomeProperties{
	"Meadow": {
		Modifier:         meadowModifier,
		SurfaceBlock:     "Grass",
		UndergroundBlock: "Dirt",
	},
	"Desert": {
		Modifier:         desertModifier,
		SurfaceBlock:     "Sand",
		UndergroundBlock: "Sand",
	},
}

type Cell struct {
	X, Z  int
	Biome BiomeProperties
}

type WorleyNoise struct {
	Seed     int64
	CellSize int
}

func NewWorleyNoise(seed int64, cellSize int) *WorleyNoise {
	return &WorleyNoise{Seed: seed, CellSize: cellSize}
}

func (w *WorleyNoise) featurePoint(cellX, cellZ int) (float64, float64) {
	h := int64(cellX*73856093^cellZ*19349663) ^ w.Seed
	r := rand.New(rand.NewSource(h))
	fx := float64(cellX*w.CellSize) + r.Float64()*float64(w.CellSize)
	fz := float64(cellZ*w.CellSize) + r.Float64()*float64(w.CellSize)
	return fx, fz
}

// Returns the nearest cell
func (w *WorleyNoise) Evaluate(x, z int) (float64, float64, int, int, int, int) {
	cellX := int(math.Floor(float64(x) / float64(w.CellSize)))
	cellZ := int(math.Floor(float64(z) / float64(w.CellSize)))

	d1, d2 := math.MaxFloat64, math.MaxFloat64
	nearestX, nearestZ := cellX, cellZ
	secondX, secondZ := cellX, cellZ

	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			fx, fz := w.featurePoint(cellX+dx, cellZ+dz)
			d := math.Hypot(float64(x)-fx, float64(z)-fz)
			if d < d1 {
				d2, secondX, secondZ = d1, nearestX, nearestZ
				d1, nearestX, nearestZ = d, cellX+dx, cellZ+dz
			} else if d < d2 {
				d2, secondX, secondZ = d, cellX+dx, cellZ+dz
			}
		}
	}
	return d1, d2, nearestX, nearestZ, secondX, secondZ
}

type BiomeSelector struct {
	Seed     int64
	CellSize int
}

func NewBiomeSelector(seed int64, cellSize int) *BiomeSelector {
	return &BiomeSelector{Seed: seed, CellSize: cellSize}
}

func (b *BiomeSelector) biomeForCell(cellX, cellZ int) BiomeProperties {
	h := int64(cellX*83492791^cellZ*1234567) ^ b.Seed
	r := rand.New(rand.NewSource(h))

	biomeKeys := []string{"Meadow", "Desert"}
	key := biomeKeys[r.Intn(len(biomeKeys))]
	return BiomeTypes[key]
}

func globalHeight(gx, gz int, p *perlin.Perlin) float64 {
	freq := 0.002
	n := p.Noise2D(float64(gx)*freq, float64(gz)*freq) // [-1,1]
	return (n + 1.0) / 2.0                             // normalizes to [0,1]
}

func meadowModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n2 := p2.Noise2D(float64(gx)*0.09, float64(gz)*0.09)
	n3 := p3.Noise2D(float64(gx)*0.01, float64(gz)*0.01)

	// Narrower mountains
	n2 = math.Pow(math.Abs(n2), 4)

	// Combines both (can be done with sum, average, or another function)
	return (n2*1.2 + n3*0.3) * float64(pkg.WorldHeight/2) // Normalizes the noise value to [0, worldHeight / 2]
}

// Flatter terrain
func desertModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n := p2.Noise2D(float64(gx)*0.003, float64(gz)*0.003)
	base := n * float64(pkg.WorldHeight/2)
	return base
}
