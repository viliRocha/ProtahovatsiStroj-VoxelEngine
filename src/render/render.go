package render

import (
	"fmt"
	"unsafe"
    "math/rand"

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
    
    /* Multidimensional Arrays Linearization, docs and extras that may come in handy
     * https://ic.unicamp.br/~bit/mc102/aulas/aula15.pdf (introdução)
     * https://felippe.ubi.pt/texts3/contr_av_ppt01p.pdf (pág. 13)
     * https://www.aussieai.com/book/ch36-linearized-multidimensional-arrays
     * https://teotl.dev/vischunk/ (may be useful)
     * (AI was used to help the interpretation of some of those docs)
     */
    Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.ChunkHeight), int(pkg.ChunkSize)
    for i := 0; i < Nx*Ny*Nz; i++ {
        pos := pkg.Coords{
            X: i / (Ny * Nz),
            Y: (i / Nz) % Ny,
            Z: i % Nz,
        }

        voxel := chunk.Voxels[pos.X][pos.Y][pos.Z]
        block := world.BlockTypes[voxel.Type]

        if !block.IsVisible || voxel.Type == "Water" || voxel.Type == "Plant" {
            continue
        }

        for face := 0; face < 6; face++ {
            if !shouldDrawFace(chunk, pos, face) {
                continue
            }

            for i := 0; i < 4; i++ {
                v := pkg.FaceVertices[face][i]
                vertices = append(vertices,
                    float32(pos.X)+v[0],
                    float32(pos.Y)+v[1],
                    float32(pos.Z)+v[2],
                )

                c := block.Color
                // Add color per vertex (RGBA)
                colorModifier := uint8(rand.Intn(16))//uint8(y * 20)

                colors = append(colors, c.R + colorModifier, c.G + colorModifier, c.B + colorModifier, c.A)
            }

            //	Add the two triangles of the face
            indices = append(indices,
                indexOffset, indexOffset+1, indexOffset+2,
                indexOffset, indexOffset+2, indexOffset+3,
            )
            indexOffset += 4
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

func shouldDrawFace(chunk *pkg.Chunk, pos pkg.Coords, faceIndex int) bool {
	direction, max_size, max_height :=
        pkg.FaceDirections[faceIndex], int(pkg.ChunkSize - 1), int(pkg.ChunkHeight) - 1

    // Calculates the new coordinates based on the face direction
    pos.X = int(pos.X+int(direction.X))
    pos.Y = int(pos.Y+int(direction.Y))
    pos.Z = int(pos.Z+int(direction.Z))
    
    // normalize the value of the coords to be within the range of 0-15
    x, y, z := pkg.Clamp(pos.X, 0, max_size), pkg.Clamp(pos.Y, 0, max_height), pkg.Clamp(pos.Z, 0, max_size)

	// Checks if the new coordinates are within the chunk bounds
	if x == pos.X && y == pos.Y && z == pos.Z {
        voxel_type := chunk.Voxels[x][y][z].Type
        return !world.BlockTypes[voxel_type].IsSolid
	}

    neighbor_index := chunk.Neighbors[faceIndex]

	if neighbor_index == nil {
		return false
	}
    
    switch faceIndex {
	case 0: // Right (X+)
        x = 0
	case 1: // Left (X-)
        x = max_size
	case 2: // Top (Y+)
        x = max_height
	case 3: // Bottom (Y-)
        return false
	case 4: // Front (Z+)
        z = 0
	case 5: // Back (Z-)
        z = max_size
	}

	return !world.BlockTypes[neighbor_index.Voxels[x][y][z].Type].IsSolid
}

func RenderVoxels(game *load.Game, is_transparent bool) {
	if is_transparent {
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

        Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.ChunkHeight), int(pkg.ChunkSize)
        for i := 0; i < Nx*Ny*Nz; i++ {
            pos := pkg.Coords{
                X: i / (Ny * Nz),
                Y: (i / Nz) % Ny,
                Z: i % Nz,
            }

            voxel := chunk.Voxels[pos.X][pos.Y][pos.Z]
            block := world.BlockTypes[voxel.Type]

            if !block.IsVisible || (block.Color.A < 255) != is_transparent {
                continue
            }

            voxelPosition := rl.NewVector3(
                chunkPosition.X+float32(pos.X),
                chunkPosition.Y+float32(pos.Y),
                chunkPosition.Z+float32(pos.Z),
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

	//	Disable blending after yielding water
	if is_transparent {
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