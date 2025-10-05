package render

import (
	"go-engine/src/pkg"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Calculate ambient occlusion based on neighbors
func calculateAmbientOcclusion(chunk *pkg.Chunk, x, y, z int) float32 {
	occlusion := 0.0
	neighbors := []struct{ X, Y, Z int }{
		{X: x - 1, Y: y, Z: z},
		{X: x + 1, Y: y, Z: z},
		{X: x, Y: y - 1, Z: z},
		{X: x, Y: y + 1, Z: z},
		{X: x, Y: y, Z: z - 1},
		{X: x, Y: y, Z: z + 1},
	}
	for _, neighbor := range neighbors {
		if isValidPosition(chunk, neighbor.X, neighbor.Y, neighbor.Z) && chunk.Voxels[neighbor.X][neighbor.Y][neighbor.Z].Type != "Air" {
			occlusion += 0.025
		}
	}
	occlusionValue := 1 - float32(occlusion) // Normalizes the value between 0 and 1
	if occlusionValue < 0 {
		occlusionValue = 0
	}
	return occlusionValue
}

// Make sure that the position is valid in the chunk
func isValidPosition(chunk *pkg.Chunk, x, y, z int) bool {
	return x >= 0 && x < pkg.ChunkSize && y >= 0 && y < pkg.ChunkSize && z >= 0 && z < pkg.ChunkSize
}

func calculateLightIntensity(voxelPosition, lightPosition rl.Vector3) float32 {
	distance := rl.Vector3Distance(voxelPosition, lightPosition)
	return rl.Clamp(1.0/(distance*distance+1)*100, 0, 1) // Adjusted intensity scale
}

func applyLighting(color rl.Color, intensity float32) rl.Color {
	return rl.NewColor(
		uint8(float32(color.R)*intensity),
		uint8(float32(color.G)*intensity),
		uint8(float32(color.B)*intensity),
		255,
	)
}
