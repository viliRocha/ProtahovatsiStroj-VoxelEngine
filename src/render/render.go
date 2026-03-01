package render

import (
	"fmt"
	"math"
	"sort"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
)

var ShowMenu bool = false
var ShowFPS bool = true
var ShowPosition bool = true
var ShowClouds bool = true

var menuScroll rl.Vector2
var menuView rl.Rectangle

func RenderVoxels(game *load.Game) {
	cam := game.Camera.Position

	rl.SetShaderValue(game.Shader, rl.GetShaderLocation(game.Shader, "viewPos"), []float32{cam.X, cam.Y, cam.Z}, rl.ShaderUniformVec3)

	// --- Round 1: solids ---
	for coord, chunk := range game.ChunkCache.Active {
		// Converts chunk coordinate to actual position
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.IsOutdated {
			BuildChunkMesh(game, chunk, chunkPos)
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
			0,
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
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		for _, voxel := range chunk.SpecialVoxels {
			pos := rl.NewVector3(
				chunkPos.X+float32(voxel.Position.X),
				chunkPos.Y+float32(voxel.Position.Y),
				chunkPos.Z+float32(voxel.Position.Z),
			)

			transparentItems = append(transparentItems, pkg.TransparentItem{
				Position:       pos,
				Type:           voxel.Type,
				Color:          world.BlockTypes[voxel.Type].Color,
				IsSurfaceWater: voxel.Type == "Water" && voxel.IsSurface,
			})
		}
	}

	// --- Back-to-front sorting ---
	if game.Camera.Position.Y >= float32(pkg.CloudHeight) {
		sort.Slice(transparentItems, func(i, j int) bool {
			di := rl.Vector3Length(rl.Vector3Subtract(transparentItems[i].Position, cam))
			dj := rl.Vector3Length(rl.Vector3Subtract(transparentItems[j].Position, cam))
			return di > dj // furthest first
		})
	}

	// --- Round 3: transparent ---
	rl.SetBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()
	//rl.BeginShaderMode(game.Shader)
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
			if ShowClouds {
				rl.DrawCube(it.Position, 1.0, 0.0, 1.0, it.Color)
			}
		}
	}

	//rl.EndShaderMode()
	rl.EnableDepthMask()
	rl.SetBlendMode(rl.BlendMode(0))
}

/*
func RenderVoxels(game *load.Game) {
	cam := game.Camera.Position
	rl.SetShaderValue(game.Shader, rl.GetShaderLocation(game.Shader, "viewPos"),
		[]float32{cam.X, cam.Y, cam.Z}, rl.ShaderUniformVec3)

	var solids []struct {
		chunk *pkg.Chunk
		pos   rl.Vector3
	}
	var plants []struct {
		voxel pkg.SpecialVoxel
		pos   rl.Vector3
	}
	var transparents []pkg.TransparentItem
	var rebuild []*pkg.Chunk

	var chunkPos rl.Vector3

	// --- Coleta de dados com lock ---
	game.ChunkCache.CacheMutex.RLock()
	for coord, chunk := range game.ChunkCache.Active {
		chunkPos = rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize))

		if chunk.IsOutdated {
			rebuild = append(rebuild, chunk)
		}

		if chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			solids = append(solids, struct {
				chunk *pkg.Chunk
				pos   rl.Vector3
			}{chunk, chunkPos})
		}

		for _, voxel := range chunk.SpecialVoxels {
			pos := rl.NewVector3(
				chunkPos.X+float32(voxel.Position.X),
				chunkPos.Y+float32(voxel.Position.Y),
				chunkPos.Z+float32(voxel.Position.Z),
			)
			if voxel.Type == "Plant" {
				plants = append(plants, struct {
					voxel pkg.SpecialVoxel
					pos   rl.Vector3
				}{voxel, pos})
			} else {
				transparents = append(transparents, pkg.TransparentItem{
					Position:       pos,
					Type:           voxel.Type,
					Color:          world.BlockTypes[voxel.Type].Color,
					IsSurfaceWater: voxel.Type == "Water" && voxel.IsSurface,
				})
			}
		}
	}
	game.ChunkCache.CacheMutex.RUnlock()

	// --- Rebuild fora do lock ---
	for _, chunk := range rebuild {
		BuildChunkMesh(game, chunk, rl.NewVector3(chunkPos.X*float32(pkg.ChunkSize), 0, chunkPos.Z*float32(pkg.ChunkSize)))
		chunk.IsOutdated = false
	}

	// --- Desenho fora do lock ---
	for _, s := range solids {
		rl.DrawModel(s.chunk.Model, s.pos, 1.0, rl.White)
	}
	for _, p := range plants {
		rl.DrawModel(p.voxel.Model, p.pos, 0.4, rl.White)
	}

	// Transparências: ordenar e desenhar
	sort.Slice(transparents, func(i, j int) bool {
		di := rl.Vector3Length(rl.Vector3Subtract(transparents[i].Position, cam))
		dj := rl.Vector3Length(rl.Vector3Subtract(transparents[j].Position, cam))
		return di > dj
	})
	rl.SetBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()
	for _, it := range transparents {
		switch it.Type {
		case "Water":
			p := rl.NewVector3(it.Position.X+0.5, it.Position.Y+0.5, it.Position.Z+0.5)
			rl.DrawPlane(p, rl.NewVector2(1.0, 1.0), it.Color)
		case "Cloud":
			rl.DrawCube(it.Position, 1.0, 0.0, 1.0, it.Color)
		}
	}
	rl.EnableDepthMask()
	rl.SetBlendMode(rl.BlendMode(0))
}
*/

func applyUnderwaterEffect(game *load.Game) {
	waterLevel := int(float64(pkg.WorldHeight)*pkg.WaterLevelFraction) + 1

	coord := world.ToChunkCoord(game.Camera.Position)
	chunk := game.ChunkCache.Active[coord]

	if chunk != nil {
		localX := int(game.Camera.Position.X) - coord.X*pkg.ChunkSize
		localY := int(game.Camera.Position.Y) - coord.Y*pkg.WorldHeight
		localZ := int(game.Camera.Position.Z) - coord.Z*pkg.ChunkSize

		if localX >= 0 && localX < pkg.ChunkSize &&
			localY >= 0 && localY < pkg.WorldHeight &&
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

	if ShowMenu {
		menuWidth := int32(rl.GetScreenWidth()) / 2
		menuHeight := int32(rl.GetScreenHeight()) / 2
		menuX := (int32(rl.GetScreenWidth()) - menuWidth) / 2
		menuY := (int32(rl.GetScreenHeight()) - menuHeight) / 2

		// Visible menu area
		menuBounds := rl.NewRectangle(float32(menuX), float32(menuY), float32(menuWidth), float32(menuHeight))

		contentHeight := float32(500) // Actual height of the content, including what is not visible.
		contentBounds := rl.NewRectangle(0, 0, float32(menuWidth-20), contentHeight)

		gui.ScrollPanel(menuBounds, "Game settings", contentBounds, &menuScroll, &menuView)

		renderMenu(menuX, menuY, menuWidth)
	}

	if ShowFPS {
		rl.DrawFPS(10, 30)
	}

	if ShowPosition {
		positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
		rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)
	}

	rl.EndDrawing()
}

func renderMenu(menuX, menuY, width int32) {
	rl.BeginScissorMode(
		int32(menuView.X),
		int32(menuView.Y),
		int32(menuView.Width),
		int32(menuView.Height),
	)

	offsetY := int32(menuScroll.Y)

	newButton(menuX+20, menuY+40+offsetY, float32(width-40), 40.0, &ShowPosition, "Show Player Position")

	newButton(menuX+20, menuY+90+offsetY, float32(width-40), 40.0, &ShowFPS, "Show FPS")

	newButton(menuX+20, menuY+140+offsetY, float32(width-40), 40.0, &ShowClouds, "Clouds")

	newGuiSlider(menuX+20, menuY+190+offsetY, float32(width-40), 40.0,
		&pkg.ChunkDistance, 1, 10,
		fmt.Sprintf("View Distance: %d", pkg.ChunkDistance),
	)

	//newGuiSlider(menuX+20, menuY+290, float32(menuWidth-40), 40.0, &load.FogCoefficient, 0.0, 1.0, fmt.Sprintf("Fog Density: %.3f", load.FogCoefficient))

	rl.EndScissorMode()
}

func newButton(menuX, menuY int32, buttonWidth, buttonHeight float32, isOn *bool, text string) {
	// raygui button
	if gui.Button(rl.NewRectangle(float32(menuX), float32(menuY), buttonWidth, buttonHeight),
		func() string {
			if *isOn {
				return fmt.Sprintf("%s: ON", text)
			}
			return fmt.Sprintf("%s: OFF", text)
		}()) {
		*isOn = !*isOn
	}
}

func newGuiSlider[T constraints.Integer | constraints.Float](menuX, menuY int32, barWidth, barHeight float32, value *T, minVal, maxVal float32, text string) {
	floatVal := float32(*value)

	rl.DrawText(text, menuX, menuY, 20, rl.DarkGray)

	floatVal = gui.Slider(rl.NewRectangle(float32(menuX), float32(menuY+30), barWidth, barHeight),
		"", "",
		floatVal, minVal, maxVal,
	)

	switch any(value).(type) {
	case *int:
		*value = T(int(math.Floor(float64(floatVal))))
	case *float32:
		*value = T(floatVal)
	}
}
