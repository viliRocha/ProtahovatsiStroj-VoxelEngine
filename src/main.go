package main

import (
	"go-engine/src/load"
	"go-engine/src/render"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	game := load.InitGame()

	// Main game loop
	for !rl.WindowShouldClose() {
		//  Update
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.DisableCursor()
		}

		rl.UpdateCamera(&game.Camera, game.CameraMode)

		// Manage chunks based on player's position
		world.ManageChunks(game.Camera.Position, game.ChunkCache, game.Perlin1, game.Perlin2, game.Perlin3)

		//  Draw
		render.RenderGame(&game)
	}
	rl.UnloadShader(game.FogShader)
	//rl.UnloadShader(game.AOShader)

	// After the loop ends:
	defer rl.CloseWindow()
}
