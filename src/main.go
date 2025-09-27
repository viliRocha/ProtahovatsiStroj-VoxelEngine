package main

import (
	"go-engine/src/load"
	"go-engine/src/render"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	game := load.InitGame()
    rl.DisableCursor()

	// Main game loop
	for !rl.WindowShouldClose() {
		//  Update
		if rl.IsKeyDown(rl.KeyOne) {
			game.CameraMode = rl.CameraFree
			game.Camera.Up = rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0} // Reset roll
		}
        
        if rl.IsKeyDown(rl.KeySpace) {
            game.Camera.Position.Y += 0.1
        }

        if rl.IsKeyDown(rl.KeyLeftShift) {
            game.Camera.Position.Y -= 0.1
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
