package main

import (
    "github.com/gen2brain/raylib-go/raylib"
    "github.com/aquilax/go-perlin"
    "math/rand"
    "fmt"
)

const (
	screenWidth  int32 = 1000
	screenHeight int32 = 480
	//  Control of the intensity/amplitude of the noise
	perlinAlpha = 3.0
	//  Adjust the frequency of noise, affecting the amount of detail present in the noise by controlling the scale of the variations.
	perlinBeta = 2.0
	//  Dimension of the space in which Perlin Noise is being calculated. For example, in 3D, it would be 3.
	perlinN = int32(3)
)

type Game struct {
	camera      rl.Camera
	cameraMode  rl.CameraMode
	voxelChunks map[rl.Vector3]*Chunk
	perlinNoise *perlin.Perlin
}

func renderGame(game *Game) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(150, 208, 233, 255)) //   Light blue

	rl.BeginMode3D(game.camera)

	for chunkPosition, chunk := range game.voxelChunks {

		for x := 0; x < chunkSize; x++ {

			for y := 0; y < chunkSize; y++ {

				for z := 0; z < chunkSize; z++ {
					voxel := chunk.Voxels[x][y][z]

					if voxel.IsSolid {
						voxelPosition := rl.NewVector3(chunkPosition.X+float32(x), chunkPosition.Y+float32(y), chunkPosition.Z+float32(z))

						// Face culling
						for i := 0; i < 6; i++ {
							if shouldDrawFace(chunk, x, y, z, i) {
								rl.DrawCube(voxelPosition, 1.0, 1.0, 1.0, voxel.Color)
							}
						}
					}
				}
			}
		}
	}
	rl.EndMode3D()

	// Draw debug text
	rl.DrawFPS(10, 30)

	positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.camera.Position.X, game.camera.Position.Y, game.camera.Position.Z)
	rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)

	rl.EndDrawing()
}

func initGame() Game {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "Protahovatsi Stroj - Basic Voxel Game")

	camera := rl.Camera{
		Position:   rl.NewVector3(2.79, 19.45, 10.0),
		Target:     rl.NewVector3(0.0, 0.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}
	cameraMode := rl.CameraFirstPerson

	// Initializes Perlin noise
	seed := rand.Int63()
	perlinNoise := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, seed)

	// Creates a hash map to store voxel data
	voxelChunks := make(map[rl.Vector3]*Chunk)
	voxelChunks[rl.NewVector3(0, 0, 0)] = generateChunk(rl.NewVector3(0, 0, 0), perlinNoise) // Passing perlinNoise

	rl.SetTargetFPS(120)

	return Game{
		camera:      camera,
		cameraMode:  cameraMode,
		voxelChunks: voxelChunks,
		perlinNoise: perlinNoise,
	}
}
