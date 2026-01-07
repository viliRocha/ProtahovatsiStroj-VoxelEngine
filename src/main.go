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
		if rl.IsKeyDown(rl.KeyOne) {
			game.CameraMode = rl.CameraFree
			game.Camera.Up = rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0} // Reset roll
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.DisableCursor()
		}

		if rl.IsKeyDown(rl.KeyLeftShift) {
			game.Camera.Position.Y -= 0.1
		}

		rl.UpdateCamera(&game.Camera, game.CameraMode)

		// Manage chunks based on player's position
		world.ManageChunks(game.Camera.Position, game.ChunkCache, game.Perlin1, game.Perlin2)

		//  Draw
		render.RenderGame(&game)
	}
	rl.UnloadShader(game.FogShader)
	//rl.UnloadShader(game.AOShader)

	// After the loop ends:
	defer rl.CloseWindow()
}
