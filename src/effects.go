package main

// Calculate ambient occlusion based on neighbors
func calculateAmbientOcclusion(chunk *Chunk, x, y, z int) float32 {
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
func isValidPosition(chunk *Chunk, x, y, z int) bool {
	return x >= 0 && x < chunkSize && y >= 0 && y < chunkSize && z >= 0 && z < chunkSize
}
