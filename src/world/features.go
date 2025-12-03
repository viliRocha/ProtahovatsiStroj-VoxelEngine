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

// see https://github.com/adct-the-experimenter/Raylib_VoxelEngine/blob/main/blockfacehelper.c for inspiration
type BlockProperties struct {
	Color     rl.Color
	IsSolid   bool
	IsVisible bool
}

var BlockTypes = map[string]BlockProperties{
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
	"Wood": {
		Color:     rl.NewColor(126, 90, 57, 255), // Light brown
		IsSolid:   true,
		IsVisible: true,
	},
	"Leaves": {
		Color:     rl.NewColor(73, 129, 49, 255), // Dark green
		IsSolid:   true,
		IsVisible: true,
	},
	"Plant": {
		Color:     rl.Red,
		IsSolid:   false,
		IsVisible: true,
	},
	"Water": {
		Color:     rl.NewColor(0, 0, 255, 110), // Transparent blue
		IsSolid:   false,
		IsVisible: true,
	},
	"Air": {
		Color:     rl.NewColor(0, 0, 0, 0), // Transparent
		IsSolid:   false,
		IsVisible: false,
	},
}

// Generate vegetation at random surface positions
func generatePlants(chunk *pkg.Chunk, chunkPos rl.Vector3, oldPlants []pkg.PlantData, reusePlants bool) {
	waterLevel := int(float64(pkg.ChunkSize) * pkg.WaterLevelFraction)

	if reusePlants && oldPlants != nil {
		fmt.Println("Reusing plants:", len(oldPlants))
		for _, plant := range oldPlants {
			localX := int(plant.Position.X) - int(chunkPos.X)
			localY := int(plant.Position.Y)
			localZ := int(plant.Position.Z) - int(chunkPos.Z)

			chunk.Voxels[localX][localY][localZ] = pkg.VoxelData{
				Type:  "Plant",
				Model: pkg.PlantModels[plant.ModelID],
			}

			chunk.Plants = append(chunk.Plants, plant)
		}
		return
	}
	plantCount := pkg.ChunkSize / 2

	for i := 0; i < plantCount; i++ {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		// Iterate from the top to the bottom of the chunk to find the surface
		for y := pkg.ChunkSize - 1; y >= 0; y-- {

			// Ensure plants are only placed above layer 13 (water)
			if chunk.Voxels[x][y][z].Type == "Grass" && y+1 < pkg.ChunkSize && chunk.Voxels[x][y+1][z].Type == "Air" && y > waterLevel {

				// Randomly define a model for the plant
				randomModel := rand.Intn(4) // 0 - 3
				chunk.Voxels[x][y+1][z] = pkg.VoxelData{
					Type:  "Plant",
					Model: pkg.PlantModels[randomModel],
				}
				plantPos := rl.NewVector3(chunkPos.X+float32(x), float32(y+1), chunkPos.Z+float32(z))
				chunk.Plants = append(chunk.Plants, pkg.PlantData{
					Position: plantPos,
					ModelID:  randomModel,
				})
				break
			}
		}
	}
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
	for range iterations {
		var newResult string
		for _, char := range result {
			if replacement, ok := rules[char]; ok {
				newResult += replacement
				continue
			}
			newResult += string(char)
		}
		result = newResult
	}
	return result
}

func placeTree(chunkCache *ChunkCache, position rl.Vector3, treeStructure string) {
	stack := []rl.Vector3{position} // Array to hold both position and direction
	currentPos := position
	direction := rl.Vector3{0, 1, 0} //	Initial direction (upwards)

	// Variables to Extend Angular Spacing
	angleIncrement := 45.0 * (1.0 + rand.Float64()*0.2) // Initial separation angle between branches
	currentAngle := 0.0

	for i := 0; i < len(treeStructure); i++ {
		char := treeStructure[i]

		switch char {
		case 'F': // Create wood blocks for tree tunks
			setVoxelGlobal(chunkCache, currentPos, pkg.VoxelData{Type: "Wood"})

			// Moving in the current direction
			currentPos = rl.Vector3{
				currentPos.X + direction.X,
				currentPos.Y + direction.Y,
				currentPos.Z + direction.Z,
			}

		case '+': // Turn right (around the Y-axis)
			direction = rl.Vector3{1, 0, 0}

		case '-': // Turn left (around the Y-axis)
			direction = rl.Vector3{-1, 0, 0}

		case '/': // Move forward (positive Z-axis)
			direction = rl.Vector3{0, 0, 1}

		case '\\': //  Move back (negative Z-axis)
			direction = rl.Vector3{0, 0, -1}

		case '[': // Save the current position and direction
			stack = append(stack, currentPos)

		case ']': // Restore the last saved position
			currentPos = stack[len(stack)-1]
			stack = stack[:len(stack)-1] //	Removes the last item from the stack

		case 'A': // Create simple branch (diagonal movement without rotation)
			// Checks if the next character is '('
			if i+1 < len(treeStructure) && treeStructure[i+1] == '(' {
				// Reads the numeric value in parentheses
				end := i + 2
				for end < len(treeStructure) && treeStructure[end] != ')' {
					end++
				}

				if end < len(treeStructure) && treeStructure[end] == ')' {
					// Extracts the number as string and converts to integer
					branchDepthStr := treeStructure[i+2 : end]
					branchDepth, err := strconv.Atoi(branchDepthStr)
					if err == nil {
						// Handles the branch with the defined depth
						for range branchDepth {
							// Calculates new direction based on current angle
							radians := currentAngle * (math.Pi / 180.0) // Converts to radians
							// Adjusts X and Z based on angle and keeps Y rising
							newDir := rl.Vector3{
								float32(math.Cos(radians)),
								1.0,
								float32(math.Sin(radians)),
							}

							newPos := rl.Vector3{
								currentPos.X + newDir.X,
								currentPos.Y + newDir.Y,
								currentPos.Z + newDir.Z,
							}

							setVoxelGlobal(chunkCache, newPos, pkg.VoxelData{Type: "Wood"})

							currentPos = newPos
							// Increases the angle to open the next branch
							currentAngle += angleIncrement

						}
					}
					// Updates the index to continue after
					i = end
				}
			}

		case 'L':
			numLeaves := 32
			radius := 2.0 // Radius of the leaf circle

			for i := range numLeaves {
				angle := float64(i) * (2 * math.Pi / float64(numLeaves)) // Calculate leaf cluster angle

				// "l": leafPos
				lx := currentPos.X + float32(radius*math.Cos(angle))
				ly := currentPos.Y + float32(rand.Intn(2)) // variação vertical)
				lz := currentPos.Z + float32(radius*math.Sin(angle))

				leafPos := rl.Vector3{lx, ly, lz}
				setVoxelGlobal(chunkCache, leafPos, pkg.VoxelData{Type: "Leaves"})
			}
		}
	}
}

func generateTrees(chunk *pkg.Chunk, chunkCache *ChunkCache, lsystemRule string, oldTrees []pkg.TreeData, reuseTrees bool) {
	waterLevel := int(float64(pkg.ChunkSize) * pkg.WaterLevelFraction)

	if reuseTrees && oldTrees != nil {
		fmt.Println("Reusing trees:", len(oldTrees))
		for _, tree := range oldTrees {
			placeTree(chunkCache, tree.Position, tree.StructureStr)
			chunk.Trees = append(chunk.Trees, tree)
		}
		return
	}

	rules := parseLSystemRule(lsystemRule)
	treeStructure := applyLSystem("F", rules, 2)
	treeCount := rand.Intn(pkg.ChunkSize / 8)

	for range treeCount {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		// Iterate over the chunk to find the surface height of the terrain
		surfaceY := -1
		for y := pkg.ChunkSize - 1; y >= 0; y-- {
			if chunk.Voxels[x][y][z].Type != "Grass" {
				continue
			}
			surfaceY = y
			break
		}

		// Make sure the surface is valid and not in the water
		if surfaceY < waterLevel {
			continue
		}
		treePos := rl.NewVector3(float32(x), float32(surfaceY+1), float32(z))

		// Build the tree with the generated structure
		placeTree(chunkCache, treePos, treeStructure)

		chunk.Trees = append(chunk.Trees, pkg.TreeData{
			Position:     treePos,
			StructureStr: treeStructure,
		})
	}
}

func genWaterFormations(chunk *pkg.Chunk) {
	waterLevel := int(float64(pkg.ChunkSize) * pkg.WaterLevelFraction)

	// Creates a Perlin Noise generator
	perlinNoise := perlin.NewPerlin(2, 2, 4, 0)

	for x := range pkg.ChunkSize {
		for z := range pkg.ChunkSize {
			//	Water shouldn't replace solid blocks (go through them)
			if chunk.Voxels[x][waterLevel][z].Type != "Air" {
				continue
			}
			chunk.Voxels[x][waterLevel][z] = pkg.VoxelData{Type: "Water"}

			for y := range pkg.ChunkSize {
				// Checks adjacent blocks for generating sand
				for dy := -3; dy <= 1; dy++ {
					for dx := -3; dx <= 3; dx++ {
						adjX := x + dx
						adjZ := z + dy

						//  adjX and adjZ >= 0 ensures that it does not access negative indices.
						//  adjX and adjZ < pkg.ChunkSize ensures that it does not exceed the chunk size.
						if adjX < 0 || adjX >= pkg.ChunkSize || adjZ < 0 || adjZ >= pkg.ChunkSize {
							continue
						}
						// Generate a Perlin Noise value
						noiseValue := perlinNoise.Noise2D(float64(adjX)/8, float64(adjZ)/8)
						voxel := chunk.Voxels[adjX][y][adjZ].Type

						// Replaces dirt and grass with sand
						if (voxel == "Grass" || voxel == "Dirt") && noiseValue > 0.32 {
							chunk.Voxels[adjX][y][adjZ] = pkg.VoxelData{Type: "Sand"}
						}
					}
				}
			}
		}
	}
}
