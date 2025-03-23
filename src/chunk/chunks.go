package main

import (
	"sync"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	chunkSize     int = 32
	chunkDistance int = 1
)

type VoxelData struct {
	Type  string
	Model rl.Model
}

type Chunk struct {
	Voxels    [chunkSize][chunkSize][chunkSize]VoxelData
	Neighbors [6]*Chunk // 0: +X, 1: -X, 2: +Y, 3: -Y, 4: +Z, 5: -Z
	Plants    []PlantData
}

type ChunkCache struct {
	chunks     map[rl.Vector3]*Chunk
	cacheMutex sync.Mutex
}

var faceDirections = []rl.Vector3{
	{1, 0, 0},  // Front
	{-1, 0, 0}, // Back
	{0, 1, 0},  // Left
	{0, -1, 0}, // Right
	{0, 0, 1},  // Top
	{0, 0, -1}, // Bottom
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		chunks: make(map[rl.Vector3]*Chunk),
	}
}

func (cc *ChunkCache) GetChunk(position rl.Vector3, p *perlin.Perlin) *Chunk {
	cc.cacheMutex.Lock()
	defer cc.cacheMutex.Unlock()

	if chunk, exists := cc.chunks[position]; exists {
		generateAbovegroundChunk(position, p, true)
		return chunk
	} else {
		chunk := generateAbovegroundChunk(position, p, false)
		cc.chunks[position] = chunk
		return chunk
	}
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.cacheMutex.Lock()
	defer cc.cacheMutex.Unlock()

	playerChunkX := int(playerPosition.X) / chunkSize
	playerChunkZ := int(playerPosition.Z) / chunkSize

	for position := range cc.chunks {
		if abs(int(position.X)/chunkSize-playerChunkX) > chunkDistance || abs(int(position.Z)/chunkSize-playerChunkZ) > chunkDistance {
			delete(cc.chunks, position)
		}
	}
}

func manageChunks(playerPosition rl.Vector3, chunkCache *ChunkCache, p *perlin.Perlin) {
	playerChunkX := int(playerPosition.X) / chunkSize
	playerChunkZ := int(playerPosition.Z) / chunkSize

	var wg sync.WaitGroup
	// Load chunks within the range
	for x := playerChunkX - chunkDistance; x <= playerChunkX+chunkDistance; x++ {
		for z := playerChunkZ - chunkDistance; z <= playerChunkZ+chunkDistance; z++ {
			chunkPosition := rl.NewVector3(float32(x*chunkSize), 0, float32(z*chunkSize))
			if _, exists := chunkCache.chunks[chunkPosition]; !exists {
				wg.Add(1)
				go func(cp rl.Vector3) {
					defer wg.Done()
					chunkCache.GetChunk(cp, p)
				}(chunkPosition)
			}
		}
	}
	wg.Wait()

	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks in all directions
	for chunkPos, chunk := range chunkCache.chunks {
		for i, direction := range faceDirections {
			neighborPos := rl.NewVector3(chunkPos.X+direction.X*float32(chunkSize), chunkPos.Y, chunkPos.Z+direction.Z*float32(chunkSize))
			if neighbor, exists := chunkCache.chunks[neighborPos]; exists {
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
