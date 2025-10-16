package world

import (
	"runtime"
	"sync"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var OppositeFaces = [6]int{1, 0, 3, 2, 5, 4}

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
	chunk, exists := cc.Chunks[position]
	cc.CacheMutex.Unlock()

	if exists {
		// Reuses plants from saved chunk
		updatedChunk := GenerateAbovegroundChunk(position, p, true)
		updatedChunk.Plants = chunk.Plants // Reassigns the old plants
		updatedChunk.IsOutdated = true

		cc.CacheMutex.Lock()
		cc.Chunks[position] = updatedChunk
		cc.CacheMutex.Unlock()

		return updatedChunk
	}

	chunk = GenerateAbovegroundChunk(position, p, false)

	cc.CacheMutex.Lock()
	cc.Chunks[position] = chunk
	cc.CacheMutex.Unlock()

	return chunk
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.CacheMutex.Lock()
	defer cc.CacheMutex.Unlock()

	playerChunkX := int(playerPosition.X) / pkg.ChunkSize
	playerChunkY := int(playerPosition.Y) / pkg.ChunkSize
	playerChunkZ := int(playerPosition.Z) / pkg.ChunkSize
	chDist := int(pkg.ChunkDistance)

	for position := range cc.Chunks {
		if Abs(int(position.X)/pkg.ChunkSize-playerChunkX) > chDist || Abs(int(position.Y)/pkg.ChunkSize-playerChunkY) > chDist || Abs(int(position.Z)/pkg.ChunkSize-playerChunkZ) > chDist {
			delete(cc.Chunks, position)
		}
	}
}

func ManageChunks(playerPosition rl.Vector3, chunkCache *ChunkCache, p *perlin.Perlin) {
	playerChunkX := int(playerPosition.X) / pkg.ChunkSize
	playerChunkY := int(playerPosition.Y) / pkg.ChunkSize
	playerChunkZ := int(playerPosition.Z) / pkg.ChunkSize

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
	for x := playerChunkX - pkg.ChunkDistance; x <= playerChunkX+pkg.ChunkDistance; x++ {
		for z := playerChunkZ - pkg.ChunkDistance; z <= playerChunkZ+pkg.ChunkDistance; z++ {
			for y := playerChunkY - pkg.ChunkDistance; y <= playerChunkY+pkg.ChunkDistance; y++ {
				chunkPosition := rl.NewVector3(float32(x*pkg.ChunkSize), float32(y*pkg.ChunkSize), float32(z*pkg.ChunkSize))
				if _, exists := chunkCache.Chunks[chunkPosition]; !exists {
					//fmt.Printf("Starting chunk loading in %v...\n", chunkPosition)
					chunkRequests <- chunkPosition
				}
			}
		}
	}
	close(chunkRequests)

	// Wait for all workers to finish
	for i := 0; i < runtime.NumCPU(); i++ {
		<-done
	}

	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks in all directions
	for chunkPos, chunk := range chunkCache.Chunks {
		for i, direction := range pkg.FaceDirections {
			neighborPos := rl.NewVector3(chunkPos.X+direction.X*float32(pkg.ChunkSize), chunkPos.Y+direction.Y*float32(pkg.ChunkSize), chunkPos.Z+direction.Z*float32(pkg.ChunkSize))
			if neighbor, exists := chunkCache.Chunks[neighborPos]; exists {
				chunk.Neighbors[i] = neighbor
				/*
					neighbor.Neighbors[OppositeFaces[i]] = chunk

					// Marcar como desatualizado para forçar reconstrução da mesh
					chunk.IsOutdated = true
					neighbor.IsOutdated = true
				*/
			} else {
				chunk.Neighbors[i] = nil
			}
		}
	}

	// Remove chunks outside the range
	chunkCache.CleanUp(playerPosition)
}

// Function to calculate the absolute value
// https://stackoverflow.com/questions/664852/which-is-the-fastest-way-to-get-the-absolute-value-of-a-number#2074403
func Abs(x int) int {
	mask := x >> 31
	return (x + mask) ^ mask
}
