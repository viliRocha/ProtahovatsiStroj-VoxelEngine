package world

import (
	"go-engine/src/pkg"
	"math/rand"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var treeModels = []string{
	"F=F[FA(3)L][FA(3)L][FA(3)L]A(3)",
	"F=F[F+A(5)L][−A(5)L][/A(4)L][\\A(4)L]",
	"F=F[A(3)L]F[-A(2)L]F[/A(2)L]F[+A(1)L]",
	"F=FFF[FA(2)L][FA(3)L][FA(4)L]",
}

func chooseRandomTree() string {
	return treeModels[rand.Intn(len(treeModels))]

	//	Intersting model that looks like kelp
	//return "F=F[FA(2)L]F[+A(3)L]F[−A(3)L]F[/A(2)L]F[\\A(2)L]A(3)"
}

func GenerateChunk(position rl.Vector3, p1, p2 *perlin.Perlin, chunkCache *ChunkCache, oldPlants []pkg.PlantData, reusePlants bool, oldTrees []pkg.TreeData, reuseTrees bool) *pkg.Chunk {
	chunk := &pkg.Chunk{
		Plants: []pkg.PlantData{},
		Trees:  []pkg.TreeData{},
	}

	waterLevel := int(float64(pkg.WorldHeight)*pkg.WaterLevelFraction) - 1

	for x := range pkg.ChunkSize {
		for z := range pkg.ChunkSize {

			// Use Perlin noise to generate the height of the terrain
			height := calculateHeight(position, x, z, p1, p2)

			for y := range pkg.WorldHeight {
				isSolid := y <= height

				if isSolid {
					chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Dirt"}

					//	Grass shouldn't generate under water
					if y == height && y > waterLevel {
						chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Grass"}
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
	if int(position.Y) == 0 {
		genLakeFormations(chunk)
	}

	//  Generate the plants after the terrain generation
	generatePlants(chunk, position, oldPlants, reusePlants, p1, p2)
	generateTrees(chunk, chunkCache, position, chooseRandomTree(), oldTrees, reuseTrees, p1, p2)

	genClouds(chunk, position, p1)

	// Marks the chunk as outdated so that the mesh can be generated
	chunk.IsOutdated = true

	return chunk
}

func calculateHeight(position rl.Vector3, x, z int, p1, p2 *perlin.Perlin) int {
	mountainFreq := 0.008
	detailsFreq := 0.05

	amp1 := 1.5
	amp2 := 0.2

	n1 := p1.Noise2D(float64(position.X+float32(x))*mountainFreq, float64(position.Z+float32(z))*mountainFreq)
	n2 := p2.Noise2D(float64(position.X+float32(x))*detailsFreq, float64(position.Z+float32(z))*detailsFreq)

	// Combines both (can be done with sum, average, or another function)
	combined := (n1*amp1 + n2*amp2) / (amp1 + amp2)

	return int((combined + 1.0) / 2.0 * float64(pkg.WorldHeight)) // Normalizes the noise value to [0, chunkSize]
}

// 3D perlin noise generation for cave systems
func GenerateCaves(position rl.Vector3, p *perlin.Perlin) *pkg.Chunk {
	chunk := &pkg.Chunk{}

	frequency := 0.04

	threshold := 0.1 // Minimum density to be solid

	for x := 0; x < pkg.ChunkSize; x++ {
		for z := 0; z < pkg.ChunkSize; z++ {
			for y := 0; y < pkg.ChunkSize; y++ {
				// Global coordinates
				globalX := int(position.X) + x
				globalY := int(position.Y) + y
				globalZ := int(position.Z) + z

				// 3D noise for density
				noise := p.Noise3D(float64(globalX)*frequency, float64(globalY)*frequency, float64(globalZ)*frequency)

				//	Threshold defines whether the voxel is solid
				if noise > threshold {
					chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Stone"}

				} else {
					chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Air"}
				}
			}
		}
	}

	// Marks the chunk as outdated so that the mesh can be generated
	chunk.IsOutdated = true

	return chunk
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
