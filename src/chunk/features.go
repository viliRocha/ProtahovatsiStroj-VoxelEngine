package main

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var plantModels [4]rl.Model

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
	"Sand": {
		Color:     rl.NewColor(236, 221, 178, 255), //	Beige
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

// Stores plant positions
type PlantData struct {
	Position rl.Vector3
	Model    rl.Model
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

func genTreePattern(iterations int) string {
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
	fmt.Println("L-system result:", result) // Debug output
	return result
}

func parseLSystemRule(ruleStr string) map[rune]string {
	rules := make(map[rune]string)
	parts := strings.Split(ruleStr, "=")
	if len(parts) == 2 {
		rules[rune(parts[0][0])] = parts[1]
	}
	return rules
}

func applyLSystem(axiom string, rules map[rune]string, iterations int) string {
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
	return result
}

func placeTree(chunk *Chunk, position rl.Vector3, treeStructure string) {
	stack := []rl.Vector3{position} // Array to hold both position and direction
	currentPos := position

	for _, char := range treeStructure {
		switch char {
		case 'F': // Create wood blocks for tree tunks
			//	Gurantee values are within the limits
			if int(currentPos.X) >= 0 && int(currentPos.X) < chunkSize &&
				int(currentPos.Y) >= 0 && int(currentPos.Y) < chunkSize &&
				int(currentPos.Z) >= 0 && int(currentPos.Z) < chunkSize {
				chunk.Voxels[int(currentPos.X)][int(currentPos.Y)][int(currentPos.Z)] = VoxelData{Type: "Wood"}
			}

			currentPos.Y += 1 // Going up a level
		case '+': // Turn right
			currentPos.X += 1
		case '-': // Turn left
			currentPos.X -= 1
		case '[': // Save the current position and direction
			stack = append(stack, currentPos)
		case ']': // Restore the last saved position and direction
			currentPos = stack[len(stack)-1]
			stack = stack[:len(stack)-1] // Remove o último item da pilha
		}
	}
}

func generateTrees(chunk *Chunk, lsystemRule string) {
	rules := parseLSystemRule(lsystemRule)

	treeStructure := applyLSystem("F", rules, 2)

	treeCount := rand.Intn(3)

	for i := 0; i < treeCount; i++ {
		x := rand.Intn(chunkSize)
		z := rand.Intn(chunkSize)

		// Iterate over the chunk to find the surface height of the terrain
		surfaceY := -1
		for y := chunkSize - 1; y >= 0; y-- {
			if blockTypes[chunk.Voxels[x][y][z].Type].IsSolid {
				surfaceY = y
				break
			}
		}

		// Make sure the surface is valid and not in the water
		if surfaceY > 13 {
			treePos := rl.NewVector3(float32(x), float32(surfaceY+1), float32(z))

			// Build the tree with the generated structure
			placeTree(chunk, treePos, treeStructure)
		}
	}
}

func genWaterFormations(chunk *Chunk) {
	// Creates a Perlin Noise generator
	perlinNoise := perlin.NewPerlin(2, 2, 4, 0)

	for x := 0; x < chunkSize; x++ {
		for z := 0; z < chunkSize; z++ {
			//	Water shouldn't replace solid blocks (go through them)
			if chunk.Voxels[x][13][z].Type == "Air" {
				chunk.Voxels[x][13][z] = VoxelData{Type: "Water"}

				for y := 0; y < chunkSize; y++ {
					// Checks adjacent blocks for generating sand
					for dy := -3; dy <= 1; dy++ {
						for dx := -3; dx <= 3; dx++ {
							adjX := x + dx
							adjZ := z + dy

							//	Ensures that adjX and adjZ are within the valid limits of the chunk.Voxels array
							if adjX >= 0 && adjX < chunkSize && adjZ >= 0 && adjZ < chunkSize {
								// Generate a Perlin Noise value
								noiseValue := perlinNoise.Noise2D(float64(adjX)/10, float64(adjZ)/10)

								if chunk.Voxels[adjX][y][adjZ].Type == "Grass" || chunk.Voxels[adjX][y][adjZ].Type == "Dirt" {
									// Replaces dirt and grass with sand
									if noiseValue > 0.32 {
										chunk.Voxels[adjX][y][adjZ] = VoxelData{Type: "Sand"}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
