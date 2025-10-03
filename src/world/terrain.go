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
	waterLevel := int(float64(pkg.ChunkSize) * pkg.WaterLevelFraction)

	for x := range pkg.ChunkSize {
        axis:
        for z := range pkg.ChunkSize {
			// Use Perlin noise to generate the height of the terrain
			height := calculateHeight(position, p, x, z)

			for y := range pkg.ChunkSize {
                //	Grass shouldn't generate under water nor bellow other blocks
                if y == height && y > waterLevel {
                    chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Grass"}
                }

                if y < waterLevel {
                    chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Dirt"}
                }
                
                if y == waterLevel {
                    chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Sand"}
                }

                if y >= height {
                    continue axis
                }

                chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Stone"}
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
