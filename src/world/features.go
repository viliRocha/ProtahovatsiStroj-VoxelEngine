package world

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BlockProperties struct {
	Color     rl.Color
	IsSolid   bool
	IsVisible bool
}

var BlockTypes = map[string]BlockProperties{
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

// Generate vegetation at random surface positions
func generatePlants(chunk *pkg.Chunk, chunkPos rl.Vector3, reusePlants bool) {
	if reusePlants {
		for _, plant := range chunk.Plants {
			chunk.Voxels[int(plant.Position.X)%pkg.ChunkSize][int(plant.Position.Y)][int(plant.Position.Z)%pkg.ChunkSize] = pkg.VoxelData{Type: "Plant", Model: plant.Model}
		}
	} else {
		plantCount := rand.Intn(20)

		for i := 0; i < plantCount; i++ {
			x := rand.Intn(pkg.ChunkSize)
			z := rand.Intn(pkg.ChunkSize)

			// Iterate from the top to the bottom of the chunk to find the surface
			for y := pkg.ChunkSize - 1; y >= 0; y-- {
				if BlockTypes[chunk.Voxels[x][y][z].Type].IsSolid {
					// Ensure plants are only placed above layer 13 (water)
					if y < pkg.ChunkSize && y > 13 {
						//  Randomly define a model for the plant
						randomModel := rand.Intn(4) // 0 - 3
						chunk.Voxels[x][y+1][z] = pkg.VoxelData{Type: "Plant", Model: pkg.PlantModels[randomModel]}
						plantPos := rl.NewVector3(chunkPos.X+float32(x), float32(y+1), chunkPos.Z+float32(z))
						chunk.Plants = append(chunk.Plants, pkg.PlantData{Position: plantPos, Model: pkg.PlantModels[randomModel]})
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

// Função para interpretar ângulos e ramificações
func interpretAngleAndBranch(command string) (float64, int) {
	angle := 15.0 // Valor padrão
	branchDepth := 0

	// Verifica se há um comando de ângulo
	if strings.HasPrefix(command, "A(") {
		// Extrai o valor de 'a' da regra A(a)
		parts := strings.Split(command[2:len(command)-1], ",")
		if len(parts) == 2 {
			if a, err := strconv.Atoi(parts[0]); err == nil {
				branchDepth = a
			}
			if angleValue, err := strconv.ParseFloat(parts[1], 64); err == nil {
				angle = angleValue
			}
		}
	}

	return angle, branchDepth
}

func placeTree(chunk *pkg.Chunk, position rl.Vector3, treeStructure string) {
	stack := []rl.Vector3{position} // Array to hold both position and direction
	currentPos := position
	direction := rl.Vector3{0, 1, 0} //	Initial direction (upwards)

	for _, char := range treeStructure {
		switch char {
		case 'F': // Create wood blocks for tree tunks
			//	Gurantee values are within bounds
			if int(currentPos.X) >= 0 && int(currentPos.X) < pkg.ChunkSize &&
				int(currentPos.Y) >= 0 && int(currentPos.Y) < pkg.ChunkSize &&
				int(currentPos.Z) >= 0 && int(currentPos.Z) < pkg.ChunkSize {
				chunk.Voxels[int(currentPos.X)][int(currentPos.Y)][int(currentPos.Z)] = pkg.VoxelData{Type: "Wood"}
			}
			currentPos = rl.Vector3{currentPos.X + direction.X, currentPos.Y + direction.Y, currentPos.Z + direction.Z}

		case '+': // Turn right (around the Y-axis)
			//direction = rotateY(direction, 25.0)

			currentPos.X += 1

		case '-': // Turn left (around the Y-axis)
			//direction = rotateY(direction, -25.0)

			currentPos.X -= 1

		case '[': // Save the current position and direction
			stack = append(stack, currentPos)

		case ']': // Restore the last saved position
			currentPos = stack[len(stack)-1]
			stack = stack[:len(stack)-1] //	Removes the last item from the stack

		case 'A': // Interpretar a regra A(a)
			branchDepth := 2 // Defina a profundidade da ramificação
			for j := 0; j < branchDepth; j++ {
				// Criar uma nova ramificação
				currentPos = calculateDiagonalPosition(currentPos, 30.0, 1.0) // Aumenta a altura em 1.0
				if int(currentPos.X) >= 0 && int(currentPos.X) < pkg.ChunkSize &&
					int(currentPos.Y) >= 0 && int(currentPos.Y) < pkg.ChunkSize &&
					int(currentPos.Z) >= 0 && int(currentPos.Z) < pkg.ChunkSize {
					chunk.Voxels[int(currentPos.X)][int(currentPos.Y)][int(currentPos.Z)] = pkg.VoxelData{Type: "Wood"}
				}
				// Voltar para a posição anterior
				currentPos = stack[len(stack)-1]
			}
		}
	}
}

// Função para calcular a nova posição diagonal com base em um ângulo
func calculateDiagonalPosition(currentPos rl.Vector3, angle float64, heightIncrement float32) rl.Vector3 {
	// Converte o ângulo para radianos
	radians := angle * (math.Pi / 180.0)
	// Calcula os incrementos de movimento baseados no ângulo
	moveX := float32(math.Sin(radians))
	moveZ := float32(math.Cos(radians))

	// Retorna a nova posição
	return rl.Vector3{
		X: currentPos.X + moveX,
		Y: currentPos.Y + heightIncrement, // Aumenta a altura a cada bloco
		Z: currentPos.Z + moveZ,
	}
}

func generateTrees(chunk *pkg.Chunk, lsystemRule string) {
	rules := parseLSystemRule(lsystemRule)

	treeStructure := applyLSystem("F", rules, 5)

	treeCount := rand.Intn(3)

	for range treeCount {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		// Iterate over the chunk to find the surface height of the terrain
		surfaceY := -1
		for y := pkg.ChunkSize - 1; y >= 0; y-- {
			if BlockTypes[chunk.Voxels[x][y][z].Type].IsSolid {
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

/*
func placeTree(chunk *pkg.Chunk, position rl.Vector3, treeStructure string) {
	stack := []rl.Vector3{position} // Array to hold both position and direction
	currentPos := position

	for _, char := range treeStructure {
		switch char {
		case 'F': // Create wood blocks for tree tunks
			//	Gurantee values are within the limits
			if int(currentPos.X) >= 0 && int(currentPos.X) < pkg.ChunkSize &&
				int(currentPos.Y) >= 0 && int(currentPos.Y) < pkg.ChunkSize &&
				int(currentPos.Z) >= 0 && int(currentPos.Z) < pkg.ChunkSize {
				chunk.Voxels[int(currentPos.X)][int(currentPos.Y)][int(currentPos.Z)] = pkg.VoxelData{Type: "Wood"}
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
*/

func genWaterFormations(chunk *pkg.Chunk) {
	// Creates a Perlin Noise generator
	perlinNoise := perlin.NewPerlin(2, 2, 4, 0)

	for x := 0; x < pkg.ChunkSize; x++ {
		for z := 0; z < pkg.ChunkSize; z++ {
			//	Water shouldn't replace solid blocks (go through them)
			if chunk.Voxels[x][13][z].Type == "Air" {
				chunk.Voxels[x][13][z] = pkg.VoxelData{Type: "Water"}

				for y := 0; y < pkg.ChunkSize; y++ {
					// Checks adjacent blocks for generating sand
					for dy := -3; dy <= 1; dy++ {
						for dx := -3; dx <= 3; dx++ {
							adjX := x + dx
							adjZ := z + dy

							//	Ensures that adjX and adjZ are within the valid limits of the chunk.Voxels array
							if adjX >= 0 && adjX < pkg.ChunkSize && adjZ >= 0 && adjZ < pkg.ChunkSize {
								// Generate a Perlin Noise value
								noiseValue := perlinNoise.Noise2D(float64(adjX)/10, float64(adjZ)/10)

								if chunk.Voxels[adjX][y][adjZ].Type == "Grass" || chunk.Voxels[adjX][y][adjZ].Type == "Dirt" {
									// Replaces dirt and grass with sand
									if noiseValue > 0.32 {
										chunk.Voxels[adjX][y][adjZ] = pkg.VoxelData{Type: "Sand"}
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
