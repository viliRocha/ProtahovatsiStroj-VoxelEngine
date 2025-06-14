package world

import (
	"sync"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ChunkCache struct {
	Chunks     map[rl.Vector3]*pkg.Chunk
	CacheMutex sync.Mutex
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		Chunks: make(map[rl.Vector3]*pkg.Chunk),
	}
}

func (cc *ChunkCache) GetChunk(position rl.Vector3, p *perlin.Perlin) *pkg.Chunk {
	cc.CacheMutex.Lock()
	defer cc.CacheMutex.Unlock()

	if chunk, exists := cc.Chunks[position]; exists {
		// Reuses plants from saved chunk
		updatedChunk := GenerateAbovegroundChunk(position, p, true)
		updatedChunk.Plants = chunk.Plants // Reassigns the old plants
		updatedChunk.IsOutdated = true
		cc.Chunks[position] = updatedChunk
		return updatedChunk
	} else {
		chunk := GenerateAbovegroundChunk(position, p, false)
		cc.Chunks[position] = chunk
		return chunk
	}
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.CacheMutex.Lock()
	defer cc.CacheMutex.Unlock()

	playerChunkX := int(playerPosition.X) / pkg.ChunkSize
	playerChunkZ := int(playerPosition.Z) / pkg.ChunkSize

	for position := range cc.Chunks {
		if abs(int(position.X)/pkg.ChunkSize-playerChunkX) > pkg.ChunkDistance || abs(int(position.Z)/pkg.ChunkSize-playerChunkZ) > pkg.ChunkDistance {
			delete(cc.Chunks, position)
		}
	}
}

func ManageChunks(playerPosition rl.Vector3, chunkCache *ChunkCache, p *perlin.Perlin) {
	playerChunkX := int(playerPosition.X) / pkg.ChunkSize
	playerChunkZ := int(playerPosition.Z) / pkg.ChunkSize

	var wg sync.WaitGroup
	// Load chunks within the range
	for x := playerChunkX - pkg.ChunkDistance; x <= playerChunkX+pkg.ChunkDistance; x++ {
		for z := playerChunkZ - pkg.ChunkDistance; z <= playerChunkZ+pkg.ChunkDistance; z++ {
			chunkPosition := rl.NewVector3(float32(x*pkg.ChunkSize), 0, float32(z*pkg.ChunkSize))
			if _, exists := chunkCache.Chunks[chunkPosition]; !exists {
				wg.Add(1)
				//fmt.Printf("Starting chunk loading in %v...\n", chunkPosition)
				go func(cp rl.Vector3) {
					defer wg.Done()
					//fmt.Printf("[%s] Loading chunk in %v\n", time.Now().Format("15:04:05.000"), cp)
					chunkCache.GetChunk(cp, p)
					//fmt.Printf("[%s] Finished chunk in %v\n", time.Now().Format("15:04:05.000"), cp)
				}(chunkPosition)
			}
		}
	}
	wg.Wait()

	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks in all directions
	for chunkPos, chunk := range chunkCache.Chunks {
		for i, direction := range pkg.FaceDirections {
			neighborPos := rl.NewVector3(chunkPos.X+direction.X*float32(pkg.ChunkSize), chunkPos.Y, chunkPos.Z+direction.Z*float32(pkg.ChunkSize))
			if neighbor, exists := chunkCache.Chunks[neighborPos]; exists {
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
