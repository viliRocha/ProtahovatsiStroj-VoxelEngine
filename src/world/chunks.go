package world

import (
	"runtime"
	"sync"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var OppositeFaces = [6]int{1, 0, 3, 2, 5, 4}

type ChunkCoord struct {
	X, Y, Z int
}

type ChunkCache struct {
	Active      map[ChunkCoord]*pkg.Chunk      // chunks ativos
	PlantsCache map[ChunkCoord][]pkg.PlantData // plantas persistentes por chunk
	CacheMutex  sync.Mutex
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		Active:      make(map[ChunkCoord]*pkg.Chunk),
		PlantsCache: make(map[ChunkCoord][]pkg.PlantData),
	}
}

func ToChunkCoord(pos rl.Vector3) ChunkCoord {
	return ChunkCoord{
		X: int(pos.X) / pkg.ChunkSize,
		Y: int(pos.Y) / pkg.ChunkSize,
		Z: int(pos.Z) / pkg.ChunkSize,
	}
}

func (cc *ChunkCache) GetChunk(position rl.Vector3, p *perlin.Perlin) *pkg.Chunk {
	coord := ToChunkCoord(position)

	cc.CacheMutex.Lock()
	_, exists := cc.Active[coord]
	oldPlants, hasPlants := cc.PlantsCache[coord]
	cc.CacheMutex.Unlock()

	var newChunk *pkg.Chunk

	if exists {
		if int(position.Y) > 0 {
			newChunk = GenerateAerialChunk(position)
		} else if int(position.Y) == 0 {
			newChunk = GenerateTerrainChunk(position, p, oldPlants, true)
		} else {
			newChunk = GenerateUndergroundChunk(position, p)
		}
		newChunk.IsOutdated = true // reset flag after reconstruction --> ensures that the mesh is rebuilt and the plant voxels are reapplied
	} else {
		// First time the chunk is generated
		if int(position.Y) > 0 {
			newChunk = GenerateAerialChunk(position)
		} else if int(position.Y) == 0 {
			// If there are saved plants, reuse them; if not, create new ones
			if hasPlants && len(oldPlants) > 0 {
				newChunk = GenerateTerrainChunk(position, p, oldPlants, true)
			} else {
				newChunk = GenerateTerrainChunk(position, p, nil, false)
			}
		} else {
			newChunk = GenerateUndergroundChunk(position, p)
		}
	}

	// Update caches
	cc.CacheMutex.Lock()
	cc.Active[coord] = newChunk
	if len(newChunk.Plants) > 0 {
		// Always save when generating/rebuilding
		cc.PlantsCache[coord] = newChunk.Plants
	} else if !hasPlants {
		// Ensures key exists even if empty to avoid future nil checks
		cc.PlantsCache[coord] = nil
	}
	cc.CacheMutex.Unlock()

	return newChunk
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.CacheMutex.Lock()
	defer cc.CacheMutex.Unlock()

	playerCoord := ToChunkCoord(playerPosition)
	chDist := int(pkg.ChunkDistance)

	for coord := range cc.Active {
		if Abs(coord.X-playerCoord.X) > chDist ||
			Abs(coord.Y-playerCoord.Y) > chDist ||
			Abs(coord.Z-playerCoord.Z) > chDist {
			delete(cc.Active, coord)
			// DO NOT delete cc.PlantsCache[coord] — plants remain stored
		}
	}
}

func ManageChunks(playerPosition rl.Vector3, chunkCache *ChunkCache, p *perlin.Perlin) {
	playerCoord := ToChunkCoord(playerPosition)

	chunkRequests := make(chan rl.Vector3, 100)
	done := make(chan struct{})

	// Worker pool
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for cp := range chunkRequests {
				//fmt.Printf("[%s] Loading chunk in %v\n", time.Now().Format("15:04:05.000"), cp)
				chunkCache.GetChunk(cp, p)
				//fmt.Printf("[%s] Finished chunk in %v\n", time.Now().Format("15:04:05.000"), cp)
			}
			done <- struct{}{}
		}()
	}

	// Send the chunk positions to be loaded
	for x := playerCoord.X - pkg.ChunkDistance; x <= playerCoord.X+pkg.ChunkDistance; x++ {
		for z := playerCoord.Z - pkg.ChunkDistance; z <= playerCoord.Z+pkg.ChunkDistance; z++ {
			for y := playerCoord.Y - pkg.ChunkDistance; y <= playerCoord.Y+pkg.ChunkDistance; y++ {
				if y > 0 {
					continue // Ignore chunks above the surface
				}

				chunkPos := rl.NewVector3(float32(x*pkg.ChunkSize), float32(y*pkg.ChunkSize), float32(z*pkg.ChunkSize))
				coord := ChunkCoord{X: x, Y: y, Z: z}

				chunkCache.CacheMutex.Lock()
				chunk, exists := chunkCache.Active[coord]
				chunkCache.CacheMutex.Unlock()

				if !exists || (chunk != nil && chunk.IsOutdated) {
					chunkRequests <- chunkPos
				}
			}
		}
	}
	close(chunkRequests)

	// Wait for all workers to finish
	for i := 0; i < runtime.NumCPU(); i++ {
		<-done
	}

	// Updates neighbors
	chunkCache.CacheMutex.Lock()
	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks in all directions
	for coord, chunk := range chunkCache.Active {
		for i, direction := range pkg.FaceDirections {
			neighborCoord := ChunkCoord{
				X: coord.X + int(direction.X),
				Y: coord.Y + int(direction.Y),
				Z: coord.Z + int(direction.Z),
			}
			if neighbor, exists := chunkCache.Active[neighborCoord]; exists {
				chunk.Neighbors[i] = neighbor
			} else {
				chunk.Neighbors[i] = nil
			}
		}
	}
	chunkCache.CacheMutex.Unlock()

	// Remove chunks outside the range
	chunkCache.CleanUp(playerPosition)
}

// Function to calculate the absolute value
// https://stackoverflow.com/questions/664852/which-is-the-fastest-way-to-get-the-absolute-value-of-a-number#2074403
func Abs(x int) int {
	mask := x >> 31
	return (x + mask) ^ mask
}
