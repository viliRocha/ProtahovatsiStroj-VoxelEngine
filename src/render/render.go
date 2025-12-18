package render

import (
	"fmt"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RenderVoxels(game *load.Game, is_transparent bool) {
	if is_transparent {
		rl.SetBlendMode(rl.BlendAlpha)
	}

	view := rl.GetCameraMatrix(game.Camera)
	projection := rl.GetMatrixProjection()

	rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_proj"), projection)
	rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_view"), view)
	//projection := rl.MatrixPerspective(game.Camera.Fovy, float32(load.ScreenWidth)/float32(load.ScreenHeight), 0.01, 1000.0)

	for coord, chunk := range game.ChunkCache.Active {
		// Converte coordenada de chunk para posição real
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			float32(coord.Y*pkg.ChunkSize),
			float32(coord.Z*pkg.ChunkSize),
		)

		// Build mesh only once if needed
		if chunk.IsOutdated {
			BuildChunkMesh(chunk, chunkPos /*, game.LightPosition*/)
			chunk.IsOutdated = false // reset flag → do not rebuild each frame
			chunk.HasMesh = true     // note that already has ready-made fabric
		}

		// If the chunk has mesh, draw directly
		if chunk.HasMesh && chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			rl.DrawModel(chunk.Model, chunkPos, 1.0, rl.White)
		}

		Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.ChunkSize), int(pkg.ChunkSize)
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
				chunkPos.X+float32(pos.X),
				chunkPos.Y+float32(pos.Y),
				chunkPos.Z+float32(pos.Z),
			)

			/*
				lightIntensity := calculateLightIntensity(voxelPosition, game.LightPosition)
				voxelColor := applyLighting(world.BlockTypes[voxel.Type].Color, lightIntensity)
			*/

			switch voxel.Type {
			case "Plant":
				rl.DrawModel(voxel.Model, voxelPosition, 0.4, rl.White)
			case "Water":
				// only draws if the voxel above is not water
				if pos.Y+1 < pkg.ChunkSize {
					above := chunk.Voxels[pos.X][pos.Y+1][pos.Z]
					if above.Type == "Water" {
						continue // do not draw an internal plane
					}
				}

				voxelPosition.X += 0.5
				voxelPosition.Y += 0.5
				voxelPosition.Z += 0.5

				rl.DrawPlane(voxelPosition, rl.NewVector2(1.0, 1.0), world.BlockTypes[voxel.Type].Color)

				/*
				   modelMatrix := rl.MatrixTranslate(voxelPosition.X, voxelPosition.Y, voxelPosition.Z)
				   rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_model"), modelMatrix)
				*/
			case "Cloud":
				rl.DrawCube(voxelPosition, 1.0, 0.0, 1.0, world.BlockTypes[voxel.Type].Color)
			}
		}
	}

	//	Disable blending after yielding water
	if is_transparent {
		rl.SetBlendMode(rl.BlendMode(0))
	}
}

func applyUnderwaterEffect(game *load.Game) {
	waterLevel := int(float64(pkg.ChunkSize)*pkg.WaterLevelFraction) + 1

	coord := world.ToChunkCoord(game.Camera.Position)
	chunk := game.ChunkCache.Active[coord]

	if chunk != nil {
		localX := int(game.Camera.Position.X) - coord.X*pkg.ChunkSize
		localY := int(game.Camera.Position.Y) - coord.Y*pkg.ChunkSize
		localZ := int(game.Camera.Position.Z) - coord.Z*pkg.ChunkSize

		if localX >= 0 && localX < pkg.ChunkSize &&
			localY >= 0 && localY < pkg.ChunkSize &&
			localZ >= 0 && localZ < pkg.ChunkSize {

			voxel := chunk.Voxels[localX][localY][localZ]
			if voxel.Type == "Water" && game.Camera.Position.Y < float32(waterLevel)-0.5 {
				// apply blue overlay
				rl.SetBlendMode(rl.BlendMode(0))
				rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))
			}
		}
	}
}

func RenderGame(game *load.Game) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(150, 208, 233, 255)) //   Light blue

	rl.BeginMode3D(game.Camera)

	//rl.BeginShaderMode(game.Shader)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	RenderVoxels(game, false)

	RenderVoxels(game, true)

	//rl.EndShaderMode()

	rl.EndMode3D()

	applyUnderwaterEffect(game)

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}
