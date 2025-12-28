package world

import (
	"math"
	"runtime"
	"sync"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var OppositeFaces = [6]int{1, 0, 3, 2, 5, 4}

const MaxChunksPerFrame = 2

type ChunkCoord struct {
	X, Y, Z int
}

type PendingWrite struct {
	Pos   [3]int
	Voxel pkg.VoxelData
}

type ChunkCache struct {
	Active        map[ChunkCoord]*pkg.Chunk      // active chunks
	PlantsCache   map[ChunkCoord][]pkg.PlantData // persistent plants by chunk
	TreesCache    map[ChunkCoord][]pkg.TreeData
	PendingVoxels map[ChunkCoord][]PendingWrite
	CacheMutex    sync.RWMutex
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		Active:        make(map[ChunkCoord]*pkg.Chunk),
		PlantsCache:   make(map[ChunkCoord][]pkg.PlantData),
		TreesCache:    make(map[ChunkCoord][]pkg.TreeData),
		PendingVoxels: make(map[ChunkCoord][]PendingWrite),
	}
}

func ToChunkCoord(pos rl.Vector3) ChunkCoord {
	return ChunkCoord{
		X: int(math.Floor(float64(pos.X) / float64(pkg.ChunkSize))),
		Y: int(math.Floor(float64(pos.Y) / float64(pkg.ChunkSize))),
		Z: int(math.Floor(float64(pos.Z) / float64(pkg.ChunkSize))),
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
			newChunk = GenerateAerialChunk(position, p, cc)
		} else if int(position.Y) == 0 {
			newChunk = GenerateTerrainChunk(position, p, cc, oldPlants, true, oldTrees, true)
		} else {
			//newChunk = GenerateUndergroundChunk(position, p)
			newChunk = GenerateAerialChunk(position, p, cc)
		}
		newChunk.IsOutdated = true // reset flag after reconstruction --> ensures that the mesh is rebuilt and the plant voxels are reapplied
	} else {
		// First time the chunk is generated
		if int(position.Y) > 0 {
			newChunk = GenerateAerialChunk(position, p, cc)
		} else if int(position.Y) == 0 {
			// If there are saved plants, reuse them; if not, create new ones
			if (hasPlants && len(oldPlants) > 0) || (hasTrees && len(oldTrees) > 0) {
				newChunk = GenerateTerrainChunk(position, p, cc, oldPlants, true, oldTrees, true)
			} else {
				newChunk = GenerateTerrainChunk(position, p, cc, nil, false, nil, false)
			}
		} else {
			//newChunk = GenerateUndergroundChunk(position, p)
			newChunk = GenerateAerialChunk(position, p, cc)
		}
		newChunk.IsOutdated = true
	}

	// Update caches
	cc.CacheMutex.Lock()
	cc.Active[coord] = newChunk
	// applies pending issues after the core terrain generation so tree voxels aren’t overwritten.
	if writes, ok := cc.PendingVoxels[coord]; ok {
		for _, w := range writes {
			newChunk.Voxels[w.Pos[0]][w.Pos[1]][w.Pos[2]] = w.Voxel
		}
		newChunk.IsOutdated = true
		delete(cc.PendingVoxels, coord)
	}

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

	// Counter to limit how many chunks are enqueued
	chunksQueued := 0

	// Send the chunk positions to be loaded
	for x := playerCoord.X - pkg.ChunkDistance; x <= playerCoord.X+pkg.ChunkDistance; x++ {
		for z := playerCoord.Z - pkg.ChunkDistance; z <= playerCoord.Z+pkg.ChunkDistance; z++ {
			for y := playerCoord.Y - pkg.ChunkDistance; y <= playerCoord.Y+pkg.ChunkDistance; y++ {
				if chunksQueued >= MaxChunksPerFrame {
					break
				}

				coord := ChunkCoord{X: x, Y: y, Z: z}

				// Allow aerial chunks to load if there are pending writes for them
				chunkCache.CacheMutex.RLock()
				chunk, exists := chunkCache.Active[coord]
				chunkCache.CacheMutex.RUnlock()

				if !exists || (chunk != nil && chunk.IsOutdated) {
					chunkPos := rl.NewVector3(float32(x*pkg.ChunkSize), float32(y*pkg.ChunkSize), float32(z*pkg.ChunkSize))

					chunkRequests <- chunkPos
					chunksQueued++
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

	// Protect map reading
	chunkCache.CacheMutex.RLock()
	chunk := chunkCache.Active[coord]
	chunkCache.CacheMutex.RUnlock()

	// math.Floor prevents inconsistent rounding that throws blocks into the wrong chunk
	localX := int(math.Floor(float64(globalPos.X))) - coord.X*pkg.ChunkSize
	localY := int(math.Floor(float64(globalPos.Y))) - coord.Y*pkg.ChunkSize
	localZ := int(math.Floor(float64(globalPos.Z))) - coord.Z*pkg.ChunkSize

	if localX < 0 && localX >= pkg.ChunkSize &&
		localY < 0 && localY >= pkg.ChunkSize &&
		localZ < 0 && localZ >= pkg.ChunkSize {
		return
	}

	if chunk == nil {
		chunkCache.CacheMutex.Lock()
		chunkCache.PendingVoxels[coord] = append(chunkCache.PendingVoxels[coord], PendingWrite{
			Pos:   [3]int{localX, localY, localZ},
			Voxel: voxel,
		})
		chunkCache.CacheMutex.Unlock()
		return
	}

	// if multiple goroutines write to the same chunk, consider a mutex per chunk
	chunkCache.CacheMutex.Lock()
	chunk.Voxels[localX][localY][localZ] = voxel
	chunk.IsOutdated = true
	chunkCache.CacheMutex.Unlock()
}

// Function to calculate the absolute value
// https://stackoverflow.com/questions/664852/which-is-the-fastest-way-to-get-the-absolute-value-of-a-number#2074403
func Abs(x int) int {
	mask := x >> 31
	return (x + mask) ^ mask
}
