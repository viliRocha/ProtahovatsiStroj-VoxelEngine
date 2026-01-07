package render

import (
	"fmt"
	"sort"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func RenderVoxels(game *load.Game) {
	cam := game.Camera.Position

	rl.SetShaderValue(game.FogShader, rl.GetShaderLocation(game.FogShader, "viewPos"), []float32{cam.X, cam.Y, cam.Z}, rl.ShaderUniformVec3)

	// --- Round 1: solids ---
	for coord, chunk := range game.ChunkCache.Active {
		// Converts chunk coordinate to actual position
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			float32(coord.Y*pkg.ChunkSize),
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.IsOutdated {
			BuildChunkMesh(game, chunk, chunkPos)
			//BuildCloudGreddyMesh(game, chunk)
			chunk.IsOutdated = false // reset flag → do not rebuild each frame
		}

		// If the chunk has mesh, draw directly
		if chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			rl.DrawModel(chunk.Model, chunkPos, 1.0, rl.White)
		}
	}

	// --- Passage 2: plants per chunk (without global sorting) ---
	for coord, chunk := range game.ChunkCache.Active {
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			float32(coord.Y*pkg.ChunkSize),
			float32(coord.Z*pkg.ChunkSize),
		)

		for _, voxel := range chunk.SpecialVoxels {
			if voxel.Type == "Plant" {
				pos := rl.NewVector3(
					chunkPos.X+float32(voxel.Position.X),
					chunkPos.Y+float32(voxel.Position.Y),
					chunkPos.Z+float32(voxel.Position.Z))
				rl.DrawModel(voxel.Model, pos, 0.4, rl.White)
			}
		}
	}

	// --- Global collection of transparencies ---
	var transparentItems []pkg.TransparentItem

	for coord, chunk := range game.ChunkCache.Active {
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			float32(coord.Y*pkg.ChunkSize),
			float32(coord.Z*pkg.ChunkSize),
		)

		for _, voxel := range chunk.SpecialVoxels {
			voxelPosition := rl.NewVector3(
				chunkPos.X+float32(voxel.Position.X),
				chunkPos.Y+float32(voxel.Position.Y),
				chunkPos.Z+float32(voxel.Position.Z),
			)

			transparentItems = append(transparentItems, pkg.TransparentItem{
				Position:       voxelPosition,
				Type:           voxel.Type,
				Color:          world.BlockTypes[voxel.Type].Color,
				IsSurfaceWater: voxel.Type == "Water" && voxel.IsSurface,
			})
		}
	}

	// --- Back-to-front sorting ---
	sort.Slice(transparentItems, func(i, j int) bool {
		di := rl.Vector3Length(rl.Vector3Subtract(transparentItems[i].Position, cam))
		dj := rl.Vector3Length(rl.Vector3Subtract(transparentItems[j].Position, cam))
		return di > dj // furthest first
	})

	// --- Round 3: transparent ---
	rl.SetBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()
	for _, it := range transparentItems {
		switch it.Type {
		case "Water":
			p := rl.NewVector3(it.Position.X+0.5, it.Position.Y+0.5, it.Position.Z+0.5)

			/*
				// calculates light intensity for this position
				lightIntensity := calculateLightIntensity(p, game.LightPosition)
				litColor := applyLighting(it.Color, lightIntensity)
			*/

			rl.DrawPlane(p, rl.NewVector2(1.0, 1.0), it.Color)
		case "Cloud":
			/*
				lightIntensity := calculateLightIntensity(it.Position, game.LightPosition)
				litColor := applyLighting(it.Color, lightIntensity)
			*/

			rl.DrawCube(it.Position, 1.0, 0.0, 1.0, it.Color)
		}
	}

	rl.EnableDepthMask()
	rl.SetBlendMode(rl.BlendMode(0))
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
	//rl.ClearBackground(rl.NewColor(134, 13, 13, 255))  Red

	rl.BeginMode3D(game.Camera)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	RenderVoxels(game)

	rl.EndMode3D()

	applyUnderwaterEffect(game)

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}
