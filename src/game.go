package main

import (
	"fmt"
	"math/rand"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  int32 = 1000
	screenHeight int32 = 480
	//  Control of the intensity/amplitude of the noise
	perlinAlpha = 3.0
	//  Adjust the frequency of noise, affecting the amount of detail present in the noise by controlling the scale of the variations.
	perlinBeta = 1.5
	//  Dimension of the space in which Perlin Noise is being calculated. For example, in 3D, it would be 3.
	perlinN = int32(3)
)

type Game struct {
	camera      rl.Camera
	cameraMode  rl.CameraMode
	chunkCache  *ChunkCache
	perlinNoise *perlin.Perlin
	shader      rl.Shader
	//lightPosition rl.Vector3
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

	//	Load .vox models
	for i := 0; i < 4; i++ {
		plantModels[i] = rl.LoadModel(fmt.Sprintf("./assets/plants/plant_%d.vox", i))
	}

	//lightPosition := rl.NewVector3(5, 5, 5)
	shader := rl.LoadShader("", "./shaders/occlusion.fs")

	chunkCache := NewChunkCache()                                                                                    // Initialize ChunkCache
	chunkCache.chunks[rl.NewVector3(0, 0, 0)] = generateAbovegroundChunk(rl.NewVector3(0, 0, 0), perlinNoise, false) // Passing perlinNoise

	rl.SetTargetFPS(120)

	return Game{
		camera:      camera,
		cameraMode:  cameraMode,
		chunkCache:  chunkCache,
		perlinNoise: perlinNoise,
		shader:      shader,
		//lightPosition: lightPosition,
	}
}
