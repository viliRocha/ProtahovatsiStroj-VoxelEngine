package render

import (
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Those functions are only used if you want to apply a light source to a specific point (like a torch)
func calculateLightIntensity(voxelPosition, lightPosition rl.Vector3) float32 {
	distance := rl.Vector3Distance(voxelPosition, lightPosition)
	return rl.Clamp(1.0/(distance*distance+1)*100, 0, 1) // Adjusted intensity scale
}

func applyLighting(color rl.Color, intensity float32) rl.Color {
	return rl.NewColor(
		uint8(float32(color.R)*intensity),
		uint8(float32(color.G)*intensity),
		uint8(float32(color.B)*intensity),
		color.A,
	)
}

// calculates occlusion factor per vertex
func calculateVoxelAO(chunk *pkg.Chunk, pos pkg.Coords, face int, corner int) float32 {
	// Each vertex of the face has 3 potential neighbors (two edges + one diagonal)
	// If everything is solid, the vertex gets darker.
	offsets := pkg.VertexAOOffsets[face][corner] // offset table (dx,dy,dz) for each neighbor
	occlusion := 0
	for _, off := range offsets {
		nx := pos.X + off[0]
		ny := pos.Y + off[1]
		nz := pos.Z + off[2]
		if nx >= 0 && nx < pkg.ChunkSize &&
			ny >= 0 && ny < pkg.ChunkSize &&
			nz >= 0 && nz < pkg.ChunkSize {
			if world.BlockTypes[chunk.Voxels[nx][ny][nz].Type].IsSolid {
				occlusion++
			}
		} else {
			// outside the chunk → treats it as solid to avoid leaving the edge too light
			occlusion++
		}
	}

	// Normalize: 0 solid neighbors = light, 3 solid neighbors = dark
	return 0.6 + 0.5*(1.0-float32(occlusion)/3.0)
}

func calculateFaceAO(chunk *pkg.Chunk, pos pkg.Coords, face int) float32 {
	total := 0.0
	for corner := 0; corner < 4; corner++ {
		total += float64(calculateVoxelAO(chunk, pos, face, corner))
	}
	// average of the 4 corners
	return float32(total / 4.0)
}
