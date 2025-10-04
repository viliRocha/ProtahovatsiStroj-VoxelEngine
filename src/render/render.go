package render

import (
	"fmt"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RenderVoxels(game *load.Game, renderTransparent bool) {
	if renderTransparent {
		rl.SetBlendMode(rl.BlendAlpha)
	}
	/*
		view := rl.GetCameraMatrix(game.Camera)
		projection := rl.GetCameraMatrix(game.Camera)

		rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_proj"), projection)
		rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_view"), view)
		projection := rl.MatrixPerspective(game.Camera.Fovy, float32(load.ScreenWidth)/float32(load.ScreenHeight), 0.01, 1000.0)
	*/

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
	if game.Camera.Position.Y < float32(waterLevel)-0.5 {
		rl.SetBlendMode(rl.BlendMode(0))
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))
	}

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}
