package main

import (
	"math/rand"
	"sync"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	chunkSize       int = 32
	chunkDistance   int = 1
	perlinFrequency     = 0.03
)

var plantModels [4]rl.Model

type VoxelData struct {
	Type  string
	Model rl.Model
}

type BlockProperties struct {
	Color     rl.Color
	IsSolid   bool
	IsVisible bool
}

var blockTypes = map[string]BlockProperties{
	"Air": {
		Color:     rl.NewColor(0, 0, 0, 0), // Transparent
		IsSolid:   false,
		IsVisible: false,
	},
	"Grass": {
		Color:     rl.NewColor(72, 174, 34, 255), // Green
		IsSolid:   true,
		IsVisible: true,
	},
	"Dirt": {
		Color:     rl.Brown,
		IsSolid:   true,
		IsVisible: true,
	},
	"Plant": {
		Color:     rl.Red,
		IsSolid:   false,
		IsVisible: true,
	},
	"Water": {
		Color:     rl.NewColor(0, 0, 255, 60), // Transparent blue
		IsSolid:   false,
		IsVisible: true,
	},
}

type Chunk struct {
	Voxels    [chunkSize][chunkSize][chunkSize]VoxelData
	Neighbors [6]*Chunk // 0: +X, 1: -X, 2: +Y, 3: -Y, 4: +Z, 5: -Z
	Plants    []PlantData
}

// Strores plant positions
type PlantData struct {
	Position rl.Vector3
	Model    rl.Model
}

type ChunkCache struct {
	chunks     map[rl.Vector3]*Chunk
	cacheMutex sync.Mutex
}

var faceDirections = []rl.Vector3{
	{1, 0, 0},  // Front
	{-1, 0, 0}, // Back
	{0, 1, 0},  // Left
	{0, -1, 0}, // Right
	{0, 0, 1},  // Top
	{0, 0, -1}, // Bottom
}

func shouldDrawFace(chunk *Chunk, x, y, z, faceIndex int) bool {
	direction := faceDirections[faceIndex]

	//  Calculates the new coordinates based on the face direction
	newX, newY, newZ := x+int(direction.X), y+int(direction.Y), z+int(direction.Z)

	// Checks if the new coordinates are within the chunk bounds
	if newX >= 0 && newX < chunkSize && newY >= 0 && newY < chunkSize && newZ >= 0 && newZ < chunkSize {
		// Returns true if the neighboring voxel is not solid
		return !blockTypes[chunk.Voxels[newX][newY][newZ].Type].IsSolid
	}

	// Checks if a neighboring voxel exists and returns true if the face should be drawn
	if chunk.Neighbors[faceIndex] != nil {

		switch faceIndex {
		case 0: // Front (x+1)
			newX = 0
		case 1: // Back (x-1)
			newX = chunkSize - 1
		case 2: // Left (y+1)
			newY = 0
		case 3: // Right (y-1)
			newY = chunkSize - 1
		case 4: // Top (z+1)
			newZ = 0
		case 5: // Bottom (z-1)
			newZ = chunkSize - 1
		}

		return !blockTypes[chunk.Neighbors[faceIndex].Voxels[newX][newY][newZ].Type].IsSolid
	}

	return true
}

// Generate vegetation at random surface positions
func generatePlants(chunk *Chunk, chunkPos rl.Vector3, reusePlants bool) {
	if reusePlants {
		for _, plant := range chunk.Plants {
			chunk.Voxels[int(plant.Position.X)%chunkSize][int(plant.Position.Y)][int(plant.Position.Z)%chunkSize] = VoxelData{Type: "Plant", Model: plant.Model}
		}
	} else {
		plantCount := rand.Intn(20)

		for i := 0; i < plantCount; i++ {
			x := rand.Intn(chunkSize)
			z := rand.Intn(chunkSize)

			// Iterate from the top to the bottom of the chunk to find the surface
			for y := chunkSize - 1; y >= 0; y-- {
				if blockTypes[chunk.Voxels[x][y][z].Type].IsSolid {
					// Ensure plants are only placed above layer 15 (water)
					if y+1 < chunkSize && y+1 > 13 {
						//  Randomly define a model for the plant
						randomModel := rand.Intn(4) // 0 - 3
						chunk.Voxels[x][y+1][z] = VoxelData{Type: "Plant", Model: plantModels[randomModel]}
						plantPos := rl.NewVector3(chunkPos.X+float32(x), float32(y+1), chunkPos.Z+float32(z))
						chunk.Plants = append(chunk.Plants, PlantData{Position: plantPos, Model: plantModels[randomModel]})
					}
					break
				}
			}
		}
	}
}

func addWater(chunk *Chunk) {
	for x := 0; x < chunkSize; x++ {
		for z := 0; z < chunkSize; z++ {
			//	Water shouldn't replace solid blocks (go through them)
			if chunk.Voxels[x][13][z].Type == "Air" {
				chunk.Voxels[x][13][z] = VoxelData{Type: "Water"}
			}
		}
	}
}

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

					if y == height {
						chunk.Voxels[x][y][z] = VoxelData{Type: "Grass"}
					}
				} else {
					chunk.Voxels[x][y][z] = VoxelData{Type: "Air"}
				}
			}
		}
	}
	//  Generate the plants after the terrain generation
	generatePlants(chunk, position, reusePlants)

	// Add water to specific layer
	addWater(chunk)

	return chunk
}

/*
func generateUndergroundChunk(position rl.Vector3, p *perlin.Perlin) *Chunk {
	chunk := &Chunk{}

	for x := 0; x < chunkSize; x++ {
		for z := 0; z < chunkSize; z++ {
			for y := 0; y < chunkSize; y++ {
				isSolid := y <= calculateHeight(position, p, x, z)

				if isSolid {
					chunk.Voxels[x][y][z] = VoxelData{Type: "Dirt"}
				} else {
					chunk.Voxels[x][y][z] = VoxelData{Type: "Air"}
				}
			}
		}
	}

	return chunk
}
*/

func calculateHeight(position rl.Vector3, p *perlin.Perlin, x, z int) int {
	noiseValue := p.Noise2D(float64(position.X+float32(x))*perlinFrequency, float64(position.Z+float32(z))*perlinFrequency)
	return int((noiseValue + 1.0) / 2.0 * float64(chunkSize)) // Normalizes the noise value to [0, chunkSize]
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		chunks: make(map[rl.Vector3]*Chunk),
	}
}

func (cc *ChunkCache) GetChunk(position rl.Vector3, p *perlin.Perlin) *Chunk {
	cc.cacheMutex.Lock()
	defer cc.cacheMutex.Unlock()

	if chunk, exists := cc.chunks[position]; exists {
		generateAbovegroundChunk(position, p, true)
		return chunk
	} else {
		chunk := generateAbovegroundChunk(position, p, false)
		cc.chunks[position] = chunk
		return chunk
	}
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.cacheMutex.Lock()
	defer cc.cacheMutex.Unlock()

	playerChunkX := int(playerPosition.X) / chunkSize
	playerChunkZ := int(playerPosition.Z) / chunkSize

	for position := range cc.chunks {
		if abs(int(position.X)/chunkSize-playerChunkX) > chunkDistance || abs(int(position.Z)/chunkSize-playerChunkZ) > chunkDistance {
			delete(cc.chunks, position)
		}
	}
}

func manageChunks(playerPosition rl.Vector3, chunkCache *ChunkCache, p *perlin.Perlin) {
	playerChunkX := int(playerPosition.X) / chunkSize
	playerChunkZ := int(playerPosition.Z) / chunkSize

	var wg sync.WaitGroup
	// Load chunks within the range
	for x := playerChunkX - chunkDistance; x <= playerChunkX+chunkDistance; x++ {
		for z := playerChunkZ - chunkDistance; z <= playerChunkZ+chunkDistance; z++ {
			chunkPosition := rl.NewVector3(float32(x*chunkSize), 0, float32(z*chunkSize))
			if _, exists := chunkCache.chunks[chunkPosition]; !exists {
				wg.Add(1)
				go func(cp rl.Vector3) {
					defer wg.Done()
					chunkCache.GetChunk(cp, p)
				}(chunkPosition)
			}
		}
	}
	wg.Wait()

	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks in all directions
	for chunkPos, chunk := range chunkCache.chunks {
		for i, direction := range faceDirections {
			neighborPos := rl.NewVector3(chunkPos.X+direction.X*float32(chunkSize), chunkPos.Y, chunkPos.Z+direction.Z*float32(chunkSize))
			if neighbor, exists := chunkCache.chunks[neighborPos]; exists {
				chunk.Neighbors[i] = neighbor
			} else {
				chunk.Neighbors[i] = nil
			}
		}
	}

	// Remove chunks outside the range
	chunkCache.CleanUp(playerPosition)
}

// Function to calculate the absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
