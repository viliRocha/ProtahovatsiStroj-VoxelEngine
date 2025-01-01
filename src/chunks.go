package main

import (
    "github.com/aquilax/go-perlin"
    "github.com/gen2brain/raylib-go/raylib"
)

const (
    chunkSize     int = 16
    chunkDistance int = 2
    perlinFrequency = 0.1
)

type VoxelData struct {
    Type    string
    Color   rl.Color
    IsSolid bool
}

func generateChunk(position rl.Vector3, p *perlin.Perlin) []VoxelData {
    voxels := make([]VoxelData, chunkSize*chunkSize*chunkSize)

    for x := 0; x < chunkSize; x++ {
        for z := 0; z < chunkSize; z++ {
            // Use Perlin noise to generate the height of the terrain
            height := calculateHeight(position, p, x, z)

            for y := 0; y < chunkSize; y++ {
                index := x + y*chunkSize + z*chunkSize*chunkSize
                isSolid := y <= height

                if isSolid {
                    voxels[index] = createVoxelData(y, height)
                } else {
                    voxels[index] = VoxelData{
                        Type:    "Air",
                        Color:   rl.NewColor(0, 0, 0, 0),// Transparent
                        IsSolid: false,
                    }
                }
            }
        }
    }
    return voxels
}

func calculateHeight(position rl.Vector3, p *perlin.Perlin, x, z int) int {
    noiseValue := p.Noise2D(float64(position.X+float32(x))*perlinFrequency, float64(position.Z+float32(z))*perlinFrequency)
    return int((noiseValue + 1.0) / 2.0 * float64(chunkSize)) // Normaliza o valor do ruído para [0, chunkSize]
}

func createVoxelData(y, height int) VoxelData {
    // Checks if the block above is air
    if y == height {
        // If the block above is air, sets it to grass
        return VoxelData{
            Type:    "Grass",
            Color:   rl.NewColor(72, 174, 34, 255),// Green
            IsSolid: true,
        }
    } else {
        return VoxelData{
            Type:    "Dirt",
            Color:   rl.Brown,
            IsSolid: true,
        }
    }
}

func manageChunks(playerPosition rl.Vector3, voxelChunks map[rl.Vector3][]VoxelData, p *perlin.Perlin) {
    playerChunkX := int(playerPosition.X) / chunkSize
    playerChunkZ := int(playerPosition.Z) / chunkSize

    // Load chunks within the range
    for x := playerChunkX - chunkDistance; x <= playerChunkX + chunkDistance; x++ {
        for z := playerChunkZ - chunkDistance; z <= playerChunkZ + chunkDistance; z++ {
            chunkPosition := rl.NewVector3(float32(x*chunkSize), 0, float32(z*chunkSize))
            if _, exists := voxelChunks[chunkPosition]; !exists {
                voxelChunks[chunkPosition] = generateChunk(chunkPosition, p)
            }
        }
    }

    // Remove chunks outside the range
    for position := range voxelChunks {
        if abs(int(position.X)/chunkSize-playerChunkX) > chunkDistance ||
            abs(int(position.Z)/chunkSize-playerChunkZ) > chunkDistance {
            delete(voxelChunks, position)
        }
    }
}

// Function to calculate the absolute value
func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

// Checks if a neighboring voxel exists and returns true if the face should be drawn
func shouldDrawFace(chunk []VoxelData, x, y, z, faceDirection int) bool {
    // Function that maps coordinates to indexes in the chunk array
    index := func(x, y, z int) int {
        return x + y*chunkSize + z*chunkSize*chunkSize
    }

    switch faceDirection {
    case 0: // Front (x+1)
        if x+1 < chunkSize && chunk[index(x+1, y, z)].IsSolid {
            return false
        }
    case 1: // Back (x-1)
        if x-1 >= 0 && chunk[index(x-1, y, z)].IsSolid {
            return false
        }
    case 2: // Left (y+1)
        if y+1 < chunkSize && chunk[index(x, y+1, z)].IsSolid {
            return false
        }
    case 3: // Right (y-1)
        if y-1 >= 0 && chunk[index(x, y-1, z)].IsSolid {
            return false
        }
    case 4: // Top (z+1)
        if z+1 < chunkSize && chunk[index(x, y, z+1)].IsSolid {
            return false
        }
    case 5: // Bottom (z-1)
        if z-1 >= 0 && chunk[index(x, y, z-1)].IsSolid {
            return false
        }
    }

    return true
}