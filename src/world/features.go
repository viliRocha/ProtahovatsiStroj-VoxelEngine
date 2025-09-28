package world

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
    "time"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BlockProperties struct {
    Texture   rl.Texture2D
    Color     rl.Color
    IsSolid   bool
    IsVisible bool
}

var BlockTypes = map[string]BlockProperties{
	"Grass": {
        Texture:     rl.LoadTexture("../assets/blocks/block.png"),
        Color:       rl.NewColor(72, 174, 34, 255), // Green
        IsSolid:     true,
        IsVisible:   true,
	},
	"Dirt": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.Brown,
        IsSolid:      true,
        IsVisible:    true,
	},
	"Sand": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.NewColor(236, 221, 178, 255), //	Beige
        IsSolid:      true,
        IsVisible:    true,
	},
	"Stone": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.Gray,
        IsSolid:      true,
        IsVisible:    true,
	},
	"Wood": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.NewColor(126, 90, 57, 255), // Light brown
        IsSolid:      true,
        IsVisible:    true,
	},
	"Leaves": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.NewColor(73, 129, 49, 255), // Dark green
        IsSolid:      true,
        IsVisible:    true,
	},
	"Plant": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.Red,
        IsSolid:      false,
        IsVisible:    true,
	},
	"Water": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.NewColor(0, 0, 255, 110), // Transparent blue
        IsSolid:      false,
        IsVisible:    true,
	},
	"Air": {
        Texture:      rl.LoadTexture("../assets/blocks/block.png"),
        Color:        rl.NewColor(0, 0, 0, 0), // Transparent
        IsSolid:      false,
        IsVisible:    false,
	},
}

// Generate vegetation at random surface positions
func generatePlants(chunk *pkg.Chunk, chunkPos rl.Vector3, reusePlants bool) {
	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) + 1

	if reusePlants {
		fmt.Println("Reusing plants:", len(chunk.Plants))
		for _, plant := range chunk.Plants {
			x := mod(int(plant.Position.X), pkg.ChunkSize)
			y := int(plant.Position.Y)
			z := mod(int(plant.Position.Z), pkg.ChunkSize)

			chunk.Voxels[x][y][z] = pkg.VoxelData{
				Type:  "Plant",
				Model: pkg.PlantModels[plant.ModelID],
			}

			chunk.Plants = append(chunk.Plants, plant)
		}
		return
	} else {
		plantCount := rand.Intn(pkg.ChunkSize / 2)

		for i := 0; i < plantCount; i++ {
			x := rand.Intn(pkg.ChunkSize)
			z := rand.Intn(pkg.ChunkSize)

			// Iterate from the top to the bottom of the chunk to find the surface
			for y := pkg.ChunkSize - 1; y >= 0; y-- {
				if BlockTypes[chunk.Voxels[x][y][z].Type].IsSolid {
					// Ensure plants are only placed above layer 13 (water)
					if y < pkg.ChunkSize && y > waterLevel {
						//  Randomly define a model for the plant
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
					}
					break
				}
			}
		}
	}
}

func mod(a, b int) int {
	return (a%b + b) % b
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
		newResult := ""
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

func placeTree(chunk *pkg.Chunk, position rl.Vector3, treeStructure string) {
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
            //	Gurantee values are within bounds
            PosY := int(currentPos.Y)
            PosX := int(currentPos.X)
            PosZ := int(currentPos.Z)

			if PosX >= 0 && PosX < int(pkg.ChunkSize) &&
				PosY >= 0 && PosY < int(pkg.ChunkHeight) &&
				PosZ >= 0 && PosZ < int(pkg.ChunkSize) {
				chunk.Voxels[PosX][PosY][PosZ] = pkg.VoxelData{Type: "Wood"}
			}
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

							// Ensures that the blocks position is within bounds
							if int(newPos.X) >= 0 && int(newPos.X) < pkg.ChunkSize &&
								int(newPos.Y) >= 0 && int(newPos.Y) < int(pkg.ChunkHeight) &&
								int(newPos.Z) >= 0 && int(newPos.Z) < pkg.ChunkSize {
								chunk.Voxels[int(newPos.X)][int(newPos.Y)][int(newPos.Z)] = pkg.VoxelData{Type: "Wood"}
							}
							currentPos = newPos
							// Increases the angle to open the next branch
							currentAngle += angleIncrement
						}
					}
					// Updates the index to continue after ')'
					i = end
				}
			}

		case 'L':
			numLeaves := 32
			radius := 2.0 // Radius of the leaf circle

			for i := range numLeaves {
                angle := float64(i) * (2 * math.Pi / float64(numLeaves)) // Calculate leaf cluster angle
                mCos := float32(radius * math.Cos(angle))

                // "l": leafPos
                lx := int(currentPos.X + mCos)
                ly := int(currentPos.Y + mCos)
                lz := int(currentPos.Z + float32(radius * math.Sin(angle)))

                if lx >= 0 && lx < pkg.ChunkSize && ly >= 0 && ly < int(pkg.ChunkHeight) && lz >= 0 && lz < pkg.ChunkSize {
                    chunk.Voxels[lx][ly][lz] = pkg.VoxelData{Type: "Leaves"}
                }
			}
		}
	}
}

func generateTrees(chunk *pkg.Chunk, lsystemRule string) {
	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) + 1

	rules := parseLSystemRule(lsystemRule)

	treeStructure := applyLSystem("F", rules, 2)

	treeCount := rand.Intn(pkg.ChunkSize / 8)

	for range treeCount {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		// Iterate over the chunk to find the surface height of the terrain
		surfaceY := -1
		for y := pkg.ChunkSize - 1; y >= 0; y-- {
            if chunk.Voxels[x][y][z].Type != "Grass" { continue }
            surfaceY = y
            break
		}

		// Make sure the surface is valid and not in the water
		if surfaceY > waterLevel {
            treePos := rl.NewVector3(float32(x), float32(surfaceY+1), float32(z))

            // Build the tree with the generated structure
            placeTree(chunk, treePos, treeStructure)
		}
	}
}

func genWaterFormations(chunk *pkg.Chunk) {
	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) + 1

	// Creates a Perlin Noise generator
	perlinNoise := perlin.NewPerlin(2, 2, 4, int64(time.Now().Unix()))

	for x := range pkg.ChunkSize {
		for z := range pkg.ChunkSize {
			//	Water shouldn't replace solid blocks (go through them)
			if chunk.Voxels[x][waterLevel][z].Type == "Air" {
                chunk.Voxels[x][waterLevel][z] = pkg.VoxelData{Type: "Water"}

                for y := range pkg.ChunkSize {
                    // Checks adjacent blocks for generating sand
                    for dy := -3; dy <= 1; dy++ {
                        for dx := -3; dx <= 3; dx++ {
                            adjX := x + dx
                            adjZ := z + dy

                            //	Ensures that adjX and adjZ are within the valid limits of the chunk.Voxels array
                            if adjX >= 0 && adjX < pkg.ChunkSize && adjZ >= 0 && adjZ < pkg.ChunkSize {
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
	}
}
