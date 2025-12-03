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
	Active      map[ChunkCoord]*pkg.Chunk      // active chunks
	PlantsCache map[ChunkCoord][]pkg.PlantData // persistent plants by chunk
	TreesCache  map[ChunkCoord][]pkg.TreeData
	CacheMutex  sync.RWMutex
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		Active:      make(map[ChunkCoord]*pkg.Chunk),
		PlantsCache: make(map[ChunkCoord][]pkg.PlantData),
		TreesCache:  make(map[ChunkCoord][]pkg.TreeData),
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

	cc.CacheMutex.RLock()
	_, exists := cc.Active[coord]
	oldPlants, hasPlants := cc.PlantsCache[coord]
	oldTrees, hasTrees := cc.TreesCache[coord]
	cc.CacheMutex.RUnlock()

	var newChunk *pkg.Chunk

	if exists {
		if int(position.Y) > 0 {
			newChunk = GenerateAerialChunk(position, cc)
		} else if int(position.Y) == 0 {
			newChunk = GenerateTerrainChunk(position, p, cc, oldPlants, true, oldTrees, true)
		} else {
			newChunk = GenerateUndergroundChunk(position, p)
		}
		newChunk.IsOutdated = true // reset flag after reconstruction --> ensures that the mesh is rebuilt and the plant voxels are reapplied
	} else {
		// First time the chunk is generated
		if int(position.Y) > 0 {
			newChunk = GenerateAerialChunk(position, cc)
		} else if int(position.Y) == 0 {
			// If there are saved plants, reuse them; if not, create new ones
			if (hasPlants && len(oldPlants) > 0) || (hasTrees && len(oldTrees) > 0) {
				newChunk = GenerateTerrainChunk(position, p, cc, oldPlants, true, oldTrees, true)
			} else {
				newChunk = GenerateTerrainChunk(position, p, cc, nil, false, nil, false)
			}
		} else {
			newChunk = GenerateUndergroundChunk(position, p)
		}
		newChunk.IsOutdated = true
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

	if len(newChunk.Trees) > 0 {
		cc.TreesCache[coord] = newChunk.Trees
	} else if !hasTrees {
		cc.TreesCache[coord] = nil
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
				chunkCache.GetChunk(cp, p)
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

				chunkCache.CacheMutex.RLock()
				chunk, exists := chunkCache.Active[coord]
				chunkCache.CacheMutex.RUnlock()

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
				if chunk.Neighbors[i] != neighbor {
					// Update reference
					chunk.Neighbors[i] = neighbor
					// Mark both as outdated to rebuild the mesh (exposed/hidden faces are recalculated, fixing the 'holes' in the terrain)
					chunk.IsOutdated = true
					neighbor.IsOutdated = true
				}
			} else {
				if chunk.Neighbors[i] != nil {
					// Neighbor was removed → mark chunk as outdated
					chunk.Neighbors[i] = nil
					chunk.IsOutdated = true
				}
			}
		}
	}
	chunkCache.CacheMutex.Unlock()

	// Remove chunks outside the range
	chunkCache.CleanUp(playerPosition)
}

func setVoxelGlobal(chunkCache *ChunkCache, globalPos rl.Vector3, voxel pkg.VoxelData) {
	coord := ToChunkCoord(globalPos)

	// Protege leitura do mapa
	chunkCache.CacheMutex.RLock()
	chunk := chunkCache.Active[coord]
	chunkCache.CacheMutex.RUnlock()

	//	Neighboring chunk wasn't rendered yet
	if chunk == nil {
		return
	}

	localX := int(globalPos.X) - coord.X*pkg.ChunkSize
	localY := int(globalPos.Y) - coord.Y*pkg.ChunkSize
	localZ := int(globalPos.Z) - coord.Z*pkg.ChunkSize

	if localX >= 0 && localX < pkg.ChunkSize &&
		localY >= 0 && localY < pkg.ChunkSize &&
		localZ >= 0 && localZ < pkg.ChunkSize {
		// Protege escrita no array de voxels
		chunkCache.CacheMutex.Lock()
		chunk.Voxels[localX][localY][localZ] = voxel
		chunkCache.CacheMutex.Unlock()
	}
}

// Function to calculate the absolute value
// https://stackoverflow.com/questions/664852/which-is-the-fastest-way-to-get-the-absolute-value-of-a-number#2074403
func Abs(x int) int {
	mask := x >> 31
	return (x + mask) ^ mask
}
