package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	game := initGame()

	// Main game loop
	for !rl.WindowShouldClose() {
		//  Update
		if rl.IsKeyPressed(rl.KeyOne) {
			game.cameraMode = rl.CameraFree
			game.camera.Up = rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0} // Reset roll
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.DisableCursor()
		}

		rl.UpdateCamera(&game.camera, game.cameraMode)

		// Manage chunks based on player's position
		manageChunks(game.camera.Position, game.chunkCache, game.perlinNoise) // Passing perlinNoise

		//  Draw
		renderGame(&game)
	}

	// After the loop ends:
	defer rl.CloseWindow()
}
