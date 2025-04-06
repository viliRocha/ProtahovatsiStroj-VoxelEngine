package render

import (
	"fmt"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func shouldDrawFace(chunk *pkg.Chunk, x, y, z, faceIndex int) bool {
	// If there is no chunk below, hide the base
	if y == 0 && chunk.Neighbors[5] == nil {
		return false
	}

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
		newY = pkg.ChunkSize - 1
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
	/*
		view := rl.GetCameraMatrix(game.Camera)
		projection := rl.GetCameraMatrix(game.Camera)

		rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_proj"), projection)
		rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_view"), view)
	*/
	//projection := rl.MatrixPerspective(game.Camera.Fovy, float32(load.ScreenWidth)/float32(load.ScreenHeight), 0.01, 1000.0)

	for chunkPosition, chunk := range game.ChunkCache.Chunks {
		for x := 0; x < pkg.ChunkSize; x++ {
			for y := 0; y < pkg.ChunkSize; y++ {
				for z := 0; z < pkg.ChunkSize; z++ {
					voxel := chunk.Voxels[x][y][z]
					if world.BlockTypes[voxel.Type].IsVisible && (world.BlockTypes[voxel.Type].Color.A < 255) == renderTransparent {

						voxelPosition := rl.NewVector3(chunkPosition.X+float32(x), chunkPosition.Y+float32(y), chunkPosition.Z+float32(z))
						for i := 0; i < 6; i++ {
							//	Face culling
							if shouldDrawFace(chunk, x, y, z, i) {
								lightIntensity := calculateLightIntensity(voxelPosition, game.LightPosition)
								voxelColor := applyLighting(world.BlockTypes[voxel.Type].Color, lightIntensity)
								/*
									modelMatrix := rl.MatrixTranslate(voxelPosition.X, voxelPosition.Y, voxelPosition.Z)
									rl.SetShaderValueMatrix(game.Shader, rl.GetShaderLocation(game.Shader, "m_model"), modelMatrix)
								*/

								switch voxel.Type {
								case "Plant":
									//	Plants are smaller than normal voxels, decrease their heigth so they touch the ground
									voxelPosition.Y -= 0.5

									rl.DrawModel(voxel.Model, voxelPosition, 0.4, rl.White)

									voxelPosition.Y += 0.5 // Reverts the setting for other operations
								case "Water":
									rl.DrawPlane(voxelPosition, rl.NewVector2(1.0, 1.0), voxelColor)
								default:
									// Ambient occlusion
									/*
										rl.PushMatrix()
										rl.Translatef(voxelPosition.X, voxelPosition.Y, voxelPosition.Z)
										model := rl.GetMatrixModelview()
										rl.PopMatrix()

										mvp := rl.MatrixMultiply(projection, rl.MatrixMultiply(view, model))
										rl.SetShaderValueMatrix(game.shader, rl.GetShaderLocation(game.shader, "mvp"), mvp)

										voxelColor := blockTypes[voxel.Type].Color
										color := []float32{float32(voxelColor.R) / 255.0, float32(voxelColor.G) / 255.0, float32(voxelColor.B) / 255.0, float32(voxelColor.A) / 255.0}
										rl.SetShaderValue(game.shader, rl.GetShaderLocation(game.shader, "color"), color, rl.ShaderUniformVec4)

										normal := rl.NewVector3(0, 0, -1) // Normal para a face correspondente
									*/

									rl.DrawCube(voxelPosition, 1.0, 1.0, 1.0, voxelColor)
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
	if game.Camera.Position.Y < float32(waterLevel) {
		rl.SetBlendMode(rl.BlendMode(0))
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))
	}

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}

func calculateLightIntensity(voxelPosition, lightPosition rl.Vector3) float32 {
	distance := rl.Vector3Distance(voxelPosition, lightPosition)
	return rl.Clamp(1.0/(distance*distance+1)*50, 0, 1) // Adjusted intensity scale
}

func applyLighting(color rl.Color, intensity float32) rl.Color {
	return rl.NewColor(
		uint8(float32(color.R)*intensity),
		uint8(float32(color.G)*intensity),
		uint8(float32(color.B)*intensity),
		255,
	)
}
