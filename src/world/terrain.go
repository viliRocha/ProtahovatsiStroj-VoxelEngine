package world

import (
	"go-engine/src/pkg"
	"math"
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

func GenerateChunk(position rl.Vector3, p1, p2, p3 *perlin.Perlin, chunkCache *ChunkCache, oldPlants []pkg.PlantData, reusePlants bool, oldTrees []pkg.TreeData, reuseTrees bool) *pkg.Chunk {
	chunk := &pkg.Chunk{
		Plants: []pkg.PlantData{},
		Trees:  []pkg.TreeData{},
	}

	// Registers the chunk in Active before generating caves
	coord := ToChunkCoord(position)
	chunkCache.CacheMutex.Lock()
	chunkCache.Active[coord] = chunk
	chunkCache.CacheMutex.Unlock()

	waterLevel := int(float64(pkg.WorldHeight)*pkg.WaterLevelFraction) - 1

	for x := range pkg.ChunkSize {
		for z := range pkg.ChunkSize {

			// Use Perlin noise to generate the height of the terrain
			height := calculateHeight(position, x, z, p1, p2, p3)

			chunk.HeightMap[x][z] = height

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

	genCaves(chunk, chunkCache, position, waterLevel, p1)

	//  Generate the plants after the terrain generation
	generatePlants(chunk, position, oldPlants, reusePlants)
	generateTrees(chunk, chunkCache, position, chooseRandomTree(), oldTrees, reuseTrees)

	genClouds(chunk, position, p1)

	// Marks the chunk as outdated so that the mesh can be generated
	chunk.IsOutdated = true

	return chunk
}

func calculateHeight(position rl.Vector3, x, z int, p1, p2, p3 *perlin.Perlin) int {
	baseFreq := 0.002
	mountainFreq := 0.09
	detailsFreq := 0.01

	ampl1 := 1.2
	ampl2 := 1.0
	ampl3 := 0.1

	bx := float64(position.X + float32(x))
	bz := float64(position.Z + float32(z))

	n1 := p1.Noise2D(bx*baseFreq, bz*baseFreq) // controls plains and depressions
	n2 := p2.Noise2D(bx*mountainFreq, bz*mountainFreq)
	n3 := p3.Noise2D(bx*detailsFreq, bz*detailsFreq)

	// Narrower mountains
	n2 = math.Pow(math.Abs(n2), 4)

	// Combines both (can be done with sum, average, or another function)
	combined := n1*ampl1 + n2*ampl2 + n3*ampl3

	return int((combined + 1.0) / 2.0 * float64((pkg.WorldHeight - 16))) // Normalizes the noise value to [0, worldHeight - 16]
}

// Perlin worms using 3D perlin noise
func genCaves(chunk *pkg.Chunk, chunkCache *ChunkCache, chunkOrigin rl.Vector3, waterLevel int, p1 *perlin.Perlin) {
	steps := 200 + rand.Intn(601)
	freq := 0.08
	radius := 2
	generationAttempt := rand.Intn(2)

	for i := 0; i < generationAttempt; i++ {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		// find the surface
		surface := chunk.HeightMap[x][z]

		if surface <= waterLevel {
			continue //	Don't create caves next to waterBodies
		}

		pos := rl.NewVector3(chunkOrigin.X+float32(x), float32(surface), chunkOrigin.Z+float32(z))

		for step := 0; step < steps; step++ {
			// direction guided by the perlin
			dx := float32(p1.Noise3D(float64(pos.X)*freq, float64(pos.Y)*freq, float64(pos.Z)*freq))
			dy := float32(p1.Noise3D(float64(pos.X)*freq+100, float64(pos.Y)*freq+100, float64(pos.Z)*freq+100))
			dz := float32(p1.Noise3D(float64(pos.X)*freq+200, float64(pos.Y)*freq+200, float64(pos.Z)*freq+200))

			dir := rl.Vector3{dx, dy, dz}
			// normalize
			length := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
			dir = rl.Vector3{dx / length, dy / length, dz / length}

			// advance
			pos = rl.Vector3{pos.X + dir.X, pos.Y + dir.Y, pos.Z + dir.Z}

			// depth limit
			if pos.Y <= 2 {
				break
			}

			// convert to local chunk
			coord := ToChunkCoord(pos)
			localX := int(pos.X) - coord.X*pkg.ChunkSize
			localY := int(pos.Y)
			localZ := int(pos.Z) - coord.Z*pkg.ChunkSize

			// check if the chunk exists
			chunkCache.CacheMutex.RLock()
			targetChunk := chunkCache.Active[coord]
			chunkCache.CacheMutex.RUnlock()

			if targetChunk != nil {
				if localX >= 0 && localX < pkg.ChunkSize &&
					localY >= 0 && localY < pkg.WorldHeight &&
					localZ >= 0 && localZ < pkg.ChunkSize {
					if targetChunk.Voxels[localX][localY][localZ].Type == "Water" {
						break
					}

					dynamicRadius := radius + rand.Intn(2) // 2 or 3

					carveSphere(targetChunk, localX, localY, localZ, dynamicRadius)
				}
			} else {
				chunkCache.CacheMutex.Lock()
				chunkCache.PendingVoxels[coord] = append(chunkCache.PendingVoxels[coord],
					PendingWrite{
						Pos:   [3]int{localX, localY, localZ},
						Voxel: pkg.VoxelData{Type: "Air"},
					})
				chunkCache.CacheMutex.Unlock()
			}
		}
	}
}

// Function to carve an air sphere (tunnel)
func carveSphere(chunk *pkg.Chunk, cx, cy, cz, radius int) {
	for x := cx - radius; x <= cx+radius; x++ {
		for y := cy - radius; y <= cy+radius; y++ {
			for z := cz - radius; z <= cz+radius; z++ {
				if x >= 0 && x < pkg.ChunkSize &&
					y >= 0 && y < pkg.WorldHeight &&
					z >= 0 && z < pkg.ChunkSize {
					dx, dy, dz := x-cx, y-cy, z-cz
					if chunk.Voxels[x][y][z].Type == "Water" {
						return
					} else if dx*dx+dy*dy+dz*dz <= radius*radius {
						chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Air"}
					}
				}
			}
		}
	}
}
