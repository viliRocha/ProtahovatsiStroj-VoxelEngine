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
		if rl.IsKeyPressed(rl.KeyOne) {
			game.CameraMode = rl.CameraFree
			game.Camera.Up = rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0} // Reset roll
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.DisableCursor()
		}

		rl.UpdateCamera(&game.Camera, game.CameraMode)

		// Manage chunks based on player's position
		world.ManageChunks(game.Camera.Position, game.ChunkCache, game.PerlinNoise) // Passing perlinNoise

		//  Draw
		render.RenderGame(&game)
	}

	// After the loop ends:
	defer rl.CloseWindow()
}
