package world

import (
	"go-engine/src/pkg"
	"math/rand"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const perlinFrequency = 0.04

func chooseRandomTree() string {
	model := rand.Intn(10)

	switch {
	case model > 4:
		return "F=F[FA(3)L][FA(3)L][FA(3)L]A(3)"
	case model < 2:
		return "F=F[A(3)L]F[-A(2)L]F[/A(2)L]F[+A(1)L]"
	default:
		return "F=FFF[FA(3)L][FA(3)L][FA(3)L]"
	}
}

func GenerateAbovegroundChunk(position rl.Vector3, p *perlin.Perlin, reusePlants bool) *pkg.Chunk {
	chunk := &pkg.Chunk{}

	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) - 1

	for x := range pkg.ChunkSize {
		for z := range pkg.ChunkSize {
			// Use Perlin noise to generate the height of the terrain
			height := calculateHeight(position, p, x, z)

			for y := range pkg.ChunkSize {
				isSolid := y <= height

				if isSolid {
					//	Grass shouldn't generate under water nor bellow other blocks
					if y == height && y > waterLevel {
						chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Grass"}
					} else if y < waterLevel {
						chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Dirt"}
					} else if y <= height-5 {
						chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Stone"}
					}
				} else {
					//	Air blocks need to be placed because water is only generated over Air blocks!!
					//	Otherwise water wold be placed on the margins...
					chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Air"}
				}
			}
		}
	}
	// Add water to specific layer
	genWaterFormations(chunk)

	//  Generate the plants after the terrain generation
	generatePlants(chunk, position, reusePlants)
	generateTrees(chunk, chooseRandomTree())

	// Marks the chunk as outdated so that the mesh can be generated
	chunk.IsOutdated = true

	return chunk
}

func calculateHeight(position rl.Vector3, p *perlin.Perlin, x, z int) int {
	noiseValue := p.Noise2D(float64(position.X+float32(x))*perlinFrequency, float64(position.Z+float32(z))*perlinFrequency)
	return int((noiseValue + 1.0) / 2.0 * float64(pkg.ChunkSize)) // Normalizes the noise value to [0, chunkSize]
}

/*
func calculateHeight(position rl.Vector3, p *perlin.Perlin, x, z int) int {
	globalX := float64(position.X + float32(x))
	globalZ := float64(position.Z + float32(z))

	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) - 2

	// Continental map — define onde há terra vs oceano
	continentFreq := 0.008
	continentVal := p.Noise2D(globalX*continentFreq, globalZ*continentFreq)
	continentVal = (continentVal + 1) / 2 // normaliza de [-1,1] pra [0,1]

	// Transição costeira suave
	if continentVal < 0.5 {
		coastBlend := (continentVal - 0.4) / 0.1 // de 0 a 1
		landBase := float64(waterLevel) + 2.0
		oceanBase := 6 + continentVal*8.8
		return int(oceanBase*(1-coastBlend) + landBase*coastBlend)
	}

	// Base do continente: levemente variada
	baseFreq := 0.008
	baseNoise := p.Noise2D(globalX*baseFreq, globalZ*baseFreq)
	baseNoise = (baseNoise + 1) / 2.0 // [0,1]
	baseHeight := float64(waterLevel) + baseNoise*10.0

	// Cadeia de montanhas
	// Ex: transformar vales profundos em picos estreitos
	mountainNoise := p.Noise2D(globalX*0.02, globalZ*0.02)
	mountainEffect := -math.Abs(mountainNoise) + 1.0 // pico em 1.0, vales em 0.0
	mountainEffect = math.Pow(mountainEffect, 4)     // só picos agudos se elevam
	mountainHeight := mountainEffect * 20.0          // até 20 blocos extras

	// Elevação base dos continentes
	finalHeight := baseHeight + mountainHeight*0.55

	mountainMaskFreq := 0.015
	mountainMask := p.Noise2D(globalX*mountainMaskFreq, globalZ*mountainMaskFreq)

	//mountainHeight := 0.0
	if mountainMask > 0.3 && baseNoise > 0.6 { // só algumas montanhas
		mountainNoise := p.Noise2D(globalX*0.03, globalZ*0.03)
		mountainEffect := math.Pow(-math.Abs(mountainNoise)+1.0, 4) // picos agudos
		mountainHeight = mountainEffect * 18.0
	}

	// Altura final com base na chunk size
	if finalHeight > float64(pkg.ChunkHeight-1) {
		finalHeight = float64(pkg.ChunkHeight - 1)
	}
	return int(finalHeight)
}
*/
