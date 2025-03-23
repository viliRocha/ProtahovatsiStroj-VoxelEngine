package main

import (
	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const perlinFrequency = 0.03

func generateAbovegroundChunk(position rl.Vector3, p *perlin.Perlin, reusePlants bool) *Chunk {
	chunk := &Chunk{}

	for x := 0; x < chunkSize; x++ {
		for z := 0; z < chunkSize; z++ {
			// Use Perlin noise to generate the height of the terrain
			height := calculateHeight(position, p, x, z)

			for y := 0; y < chunkSize; y++ {
				isSolid := y <= height

				if isSolid {
					chunk.Voxels[x][y][z] = VoxelData{Type: "Dirt"}

					//	Grass shoulden't generate under water
					if y == height && y > 12 {
						chunk.Voxels[x][y][z] = VoxelData{Type: "Grass"}
					} else if y <= height-5 {
						chunk.Voxels[x][y][z] = VoxelData{Type: "Stone"}
					}
				} else {
					chunk.Voxels[x][y][z] = VoxelData{Type: "Air"}
				}
			}
		}
	}
	//  Generate the plants after the terrain generation
	generatePlants(chunk, position, reusePlants)

	generateTrees(chunk, "F=F[+F[+F]F[-F]]F[-F[+F]F[-F]]")

	// Add water to specific layer
	genWaterFormations(chunk)

	return chunk
}

func calculateHeight(position rl.Vector3, p *perlin.Perlin, x, z int) int {
	noiseValue := p.Noise2D(float64(position.X+float32(x))*perlinFrequency, float64(position.Z+float32(z))*perlinFrequency)
	return int((noiseValue + 1.0) / 2.0 * float64(chunkSize)) // Normalizes the noise value to [0, chunkSize]
}
