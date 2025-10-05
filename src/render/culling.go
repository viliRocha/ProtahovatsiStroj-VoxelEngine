package render

import (
	"math/rand"
	"unsafe"

	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func BuildChunkMesh(chunk *pkg.Chunk, chunkPos rl.Vector3) {
	var vertices []float32
	var indices []uint16
	var colors []uint8

	indexOffset := uint16(0)

	for x := 0; x < pkg.ChunkSize; x++ {
		for y := 0; y < pkg.ChunkHeight; y++ {
			for z := 0; z < pkg.ChunkSize; z++ {
				voxel := chunk.Voxels[x][y][z]
				block := world.BlockTypes[voxel.Type]

				if !block.IsVisible || voxel.Type == "Water" || voxel.Type == "Plant" {
					continue
				}

				for face := 0; face < 6; face++ {
					if !shouldDrawFace(chunk, x, y, z, face) && voxel.Type != "Leaves" {
						continue
					}

					for i := 0; i < 4; i++ {
						v := pkg.FaceVertices[face][i]
						vertices = append(vertices,
							float32(x)+v[0],
							float32(y)+v[1],
							float32(z)+v[2],
						)

						// Add color per vertex (RGBA)
						c := block.Color
						colorVariation := uint8(rand.Intn(8))
						colors = append(colors, c.R+colorVariation, c.G+colorVariation, c.B+colorVariation, c.A)
					}

					//	Add the two triangles of the face
					indices = append(indices,
						indexOffset, indexOffset+1, indexOffset+2,
						indexOffset, indexOffset+2, indexOffset+3,
					)
					indexOffset += 4
				}
			}
		}
	}

	mesh := rl.Mesh{
		VertexCount:   int32(len(vertices) / 3),
		TriangleCount: int32(len(indices) / 3),
	}

	if len(vertices) > 0 {
		mesh.Vertices = (*float32)(unsafe.Pointer(&vertices[0]))
	}
	if len(indices) > 0 {
		mesh.Indices = (*uint16)(unsafe.Pointer(&indices[0]))
	}
	if len(colors) > 0 {
		mesh.Colors = (*uint8)(unsafe.Pointer(&colors[0]))
	}

	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)

	// Create material and assign it
	material := rl.LoadMaterialDefault()
	materials := []rl.Material{material}
	model.MaterialCount = int32(len(materials))
	model.Materials = &materials[0]

	// Assign to chunk
	chunk.Model = model
	chunk.HasMesh = true
	chunk.IsOutdated = false
}

func shouldDrawFace(chunk *pkg.Chunk, x, y, z, faceIndex int) bool {
	direction := pkg.FaceDirections[faceIndex]

	//  Calculates the new coordinates based on the face direction
	newX, newY, newZ := x+int(direction.X), y+int(direction.Y), z+int(direction.Z)

	// Checks if the new coordinates are within the chunk bounds
	if newX >= 0 && newX < pkg.ChunkSize && newY >= 0 && newY < pkg.ChunkSize && newZ >= 0 && newZ < pkg.ChunkSize {
		// Returns true if the neighboring voxel is not solid
		return !world.BlockTypes[chunk.Voxels[newX][newY][newZ].Type].IsSolid
	}

	if chunk.Neighbors[faceIndex] == nil {
		return false
	}

	// Checks if a neighboring voxel exists and returns true if the face should be drawn
	switch faceIndex {
	case 0: // Right (X+)
		newX = 0
	case 1: // Left (X-)
		newX = pkg.ChunkSize - 1
	case 2: // Top (Y+)
		newY = 0
	case 3: // Bottom (Y-)
		//	No need to render the chunks bottom
		return false
	case 4: // Front (Z+)
		newZ = 0
	case 5: // Back (Z-)
		newZ = pkg.ChunkSize - 1
	}

	return !world.BlockTypes[chunk.Neighbors[faceIndex].Voxels[newX][newY][newZ].Type].IsSolid
}
