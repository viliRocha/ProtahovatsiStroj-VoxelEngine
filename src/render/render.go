package render

import (
	"fmt"
	"unsafe"

	"go-engine/src/load"
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
		for y := 0; y < int(pkg.ChunkHeight); y++ {
			for z := 0; z < pkg.ChunkSize; z++ {
				voxel := chunk.Voxels[x][y][z]
				block := world.BlockTypes[voxel.Type]

				if !block.IsVisible || voxel.Type == "Water" || voxel.Type == "Plant" {
					continue
				}

				for face := 0; face < 6; face++ {
					if !shouldDrawFace(chunk, x, y, z, face) {
						continue
					}

					for i := 0; i < 4; i++ {
                        v := pkg.FaceVertices[face][i]
                        vertices = append(vertices,
				            float32(x)+v[0],
				            float32(y)+v[1],
				            float32(z)+v[2],
                        )

                        c := block.Color
                        // Add color per vertex (RGBA)
                        colors = append(colors, c.R, c.G, c.B, c.A)
                        // texturecoords on common.go
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

func RenderVoxels(game *load.Game, renderTransparent bool) {
	if renderTransparent {
		rl.SetBlendMode(rl.BlendAlpha)
	}
	
    view := rl.GetCameraMatrix(game.Camera)
    projection := rl.GetCameraMatrix(game.Camera)

    rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_proj"), projection)
    rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_view"), view)
    //projection := rl.MatrixPerspective(game.Camera.Fovy, float32(load.ScreenWidth)/float32(load.ScreenHeight), 0.01, 1000.0)
	

	for chunkPosition, chunk := range game.ChunkCache.Chunks {
		// Build mesh only once if needed
		if chunk.IsOutdated {
			BuildChunkMesh(chunk, chunkPosition)
		}

		// If the chunk has mesh, draw directly
		if chunk.HasMesh && chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			rl.DrawModel(chunk.Model, chunkPosition, 1.0, rl.White)
		}

		for x := range pkg.ChunkSize {
			for y := range pkg.ChunkHeight {
				for z := range pkg.ChunkSize {
					voxel := chunk.Voxels[x][y][z]
					block := world.BlockTypes[voxel.Type]

					if !block.IsVisible || (block.Color.A < 255) != renderTransparent {
						continue
					}

					voxelPosition := rl.NewVector3(
						chunkPosition.X+float32(x),
						chunkPosition.Y+float32(y),
						chunkPosition.Z+float32(z),
					)

					switch voxel.Type {
					case "Plant":
						rl.DrawModel(voxel.Model, voxelPosition, 0.4, rl.White)

					case "Water":
						voxelPosition.X += 0.5
						voxelPosition.Y += 0.5
						voxelPosition.Z += 0.5

						rl.DrawPlane(voxelPosition, rl.NewVector2(1.0, 1.0), world.BlockTypes[voxel.Type].Color)

						/*
							modelMatrix := rl.MatrixTranslate(voxelPosition.X, voxelPosition.Y, voxelPosition.Z)
							rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_model"), modelMatrix)
						*/
					}
				}
			}
		}
	}

	//	Disable blending after yielding water
	if renderTransparent {
		rl.SetBlendMode(rl.BlendMode(0))
	}
}

func RenderGame(game *load.Game) {
	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) + 1

	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(150, 208, 233, 255)) //   Light blue

	rl.BeginMode3D(game.Camera)

	//rl.BeginShaderMode(game.Shader)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	RenderVoxels(game, false)

	RenderVoxels(game, true)

	//rl.EndShaderMode()

	rl.EndMode3D()

	// Apply light blue filter - for underwater
	if game.Camera.Position.Y < float32(waterLevel)+0.5 {
		rl.SetBlendMode(rl.BlendMode(0))
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))
	}

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}