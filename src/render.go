package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func renderVoxels(game *Game, renderTransparent bool) {
	if renderTransparent {
		rl.SetBlendMode(rl.BlendAlpha)
	}

	for chunkPosition, chunk := range game.chunkCache.chunks {
		for x := 0; x < chunkSize; x++ {
			for y := 0; y < chunkSize; y++ {
				for z := 0; z < chunkSize; z++ {
					voxel := chunk.Voxels[x][y][z]
					if blockTypes[voxel.Type].IsVisible && (blockTypes[voxel.Type].Color.A < 255) == renderTransparent {

						voxelPosition := rl.NewVector3(chunkPosition.X+float32(x), chunkPosition.Y+float32(y), chunkPosition.Z+float32(z))
						for i := 0; i < 6; i++ {
							//	Face culling
							if shouldDrawFace(chunk, x, y, z, i) {
								/*
									lightIntensity := calculateLightIntensity(voxelPosition, game.lightPosition)
									voxelColor := applyLighting(blockTypes[voxel.Type].Color, lightIntensity)
								*/
								switch voxel.Type {
								case "Plant":
									//	Plants are smaller than normal voxels, decrease their heigth so they touch the ground
									voxelPosition.Y -= 0.5

									rl.DrawModel(voxel.Model, voxelPosition, 0.4, rl.White)

									voxelPosition.Y += 0.5 // Reverts the setting for other operations
								case "Water":
									rl.DrawPlane(voxelPosition, rl.NewVector2(1.0, 1.0), blockTypes[voxel.Type].Color)
								default:
									// Ambient occlusion
									occlusion := calculateAmbientOcclusion(chunk, x, y, z)
									rl.SetShaderValue(game.shader, rl.GetShaderLocation(game.shader, "ambientOcclusion"), []float32{occlusion}, rl.ShaderUniformFloat)

									color := rl.Color(blockTypes[chunk.Voxels[x][y][z].Type].Color)

									rl.DrawCube(voxelPosition, 1.0, 1.0, 1.0, darkenColor(color, -100))
								}
							}
						}
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

func renderGame(game *Game) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(150, 208, 233, 255)) //   Light blue

	rl.BeginMode3D(game.camera)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	renderVoxels(game, false)

	renderVoxels(game, true)

	rl.EndMode3D()

	// Apply light blue filter - for underwater
	if game.camera.Position.Y < 13 {
		rl.SetBlendMode(rl.BlendMode(0))
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))
	}

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.camera.Position.X, game.camera.Position.Y, game.camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}

/*
func calculateLightIntensity(voxelPosition, lightPosition rl.Vector3) float32 {
	distance := rl.Vector3Distance(voxelPosition, lightPosition)
	return rl.Clamp(1.0/(distance*distance+1)*50, 0, 1)// Adjusted intensity scale
}

func applyLighting(color rl.Color, intensity float32) rl.Color {
	return rl.NewColor(
		uint8(float32(color.R)*intensity),
		uint8(float32(color.G)*intensity),
		uint8(float32(color.B)*intensity),
		255,
	)
}
*/
