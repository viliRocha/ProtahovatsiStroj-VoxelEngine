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
	"Stone": {
		Color:     rl.Gray,
		IsSolid:   true,
		IsVisible: true,
	},
	"Plant": {
		Color:     rl.Red,
		IsSolid:   false,
		IsVisible: true,
	},
	"Wood": {
		Color:     rl.NewColor(126, 90, 57, 255), // Light brown
		IsSolid:   false,
		IsVisible: true,
	},
	"Leaves": {
		Color:     rl.NewColor(73, 129, 49, 255), // Dark green
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
	// If there is no chunk below, hide the base
	if y == 0 && chunk.Neighbors[5] == nil {
		return false
	}

	direction := faceDirections[faceIndex]

	//  Calculates the new coordinates based on the face direction
	newX, newY, newZ := x+int(direction.X), y+int(direction.Y), z+int(direction.Z)

	// Checks if the new coordinates are within the chunk bounds
	if newX >= 0 && newX < chunkSize && newY >= 0 && newY < chunkSize && newZ >= 0 && newZ < chunkSize {
		// Returns true if the neighboring voxel is not solid
		return !blockTypes[chunk.Voxels[newX][newY][newZ].Type].IsSolid
	}

	if chunk.Neighbors[faceIndex] == nil {
		return false
	}

	// Checks if a neighboring voxel exists and returns true if the face should be drawn
	switch faceIndex {
	case 0: // Right (X+)
		newX = 0
	case 1: // Left (X-)
		newX = chunkSize - 1
	case 2: // Top (Y+)
		newY = 0
	case 3: // Bottom (Y-)
		newY = chunkSize - 1
	case 4: // Front (Z+)
		newZ = 0
	case 5: // Back (Z-)
		newZ = chunkSize - 1
	}

	return !blockTypes[chunk.Neighbors[faceIndex].Voxels[newX][newY][newZ].Type].IsSolid
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
					// Ensure plants are only placed above layer 13 (water)
					if y < chunkSize && y > 13 {
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

/*
func applyLSystem(iterations int) string {
	axiom := "X"
	rules := map[rune]string{
		'X': "F[-X][X]F[-X]+X",
		'F': "FF",
	}
	result := axiom
	for i := 0; i < iterations; i++ {
		newResult := ""
		for _, char := range result {
			if replacement, ok := rules[char]; ok {
				newResult += replacement
			} else {
				newResult += string(char)
			}
		}
		result = newResult
	}
	//fmt.Println("L-system result:", result) // Debug output
	return result
}

func generateTrees(chunk *Chunk, startPosition rl.Vector3, lSystem string) {
	stack := make([][2]rl.Vector3, 0) // Array to hold both position and direction
	position := startPosition
	direction := rl.NewVector3(0, 1, 0) // Initial direction is upwards

	for _, char := range lSystem {
		//fmt.Printf("Processing char: %c, Position: %v, Direction: %v\n", char, position, direction) // Debug output

		switch char {
		case 'F':
			if position.Y >= 0 && int(position.Y) < chunkSize {
				chunk.Voxels[int(position.X)][int(position.Y)][int(position.Z)] = VoxelData{Type: "Wood"}
				position.X += direction.X
				position.Y += direction.Y
				position.Z += direction.Z
			}
		case '-':
			// Turn left
			angle := -math.Pi / 4 // Adjust the angle as needed
			direction = rl.NewVector3(
				direction.X*float32(math.Cos(angle))-direction.Z*float32(math.Sin(angle)),
				direction.Y,
				direction.X*float32(math.Sin(angle))+direction.Z*float32(math.Cos(angle)),
			)
		case '+':
			// Turn right
			angle := math.Pi / 4 // Adjust the angle as needed
			direction = rl.NewVector3(
				direction.X*float32(math.Cos(angle))-direction.Z*float32(math.Sin(angle)),
				direction.Y,
				direction.X*float32(math.Sin(angle))+direction.Z*float32(math.Cos(angle)),
			)
		case '[':
			// Save the current position and direction
			stack = append(stack, [2]rl.Vector3{position, direction})
		case ']':
			// Restore the last saved position and direction
			if len(stack) > 0 {
				position, direction = stack[len(stack)-1][0], stack[len(stack)-1][1]
				stack = stack[:len(stack)-1]
			}
		}

		// Print current position and direction after processing the character
		//fmt.Printf("After processing char: %c, Position: %v, Direction: %v\n", char, position, direction) // Debug output
	}
}
*/

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

	//generateTrees(chunk, rl.Vector3{X: float32(rand.Intn(chunkSize)), Y: float32(chunkSize - 1), Z: float32(rand.Intn(chunkSize))}, applyLSystem(4))

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
