package load

import (
	"fmt"
	"math/rand"

	"go-engine/src/pkg"
	"go-engine/src/world"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ScreenWidth  int32 = 1000
	ScreenHeight int32 = 480
	//  Control of the intensity/amplitude of the noise
	perlinAlpha = 3
	//  Adjust the frequency of noise, affecting the amount of detail present in the noise by controlling the scale of the variations.
	perlinBeta = 1.5
	//  Dimension of the space in which Perlin Noise is being calculated. For example, in 3D, it would be 3.
	perlinN = int32(2)
)

type Game struct {
	Camera      rl.Camera
	CameraMode  rl.CameraMode
	ChunkCache  *world.ChunkCache
	PerlinNoise *perlin.Perlin
	Shader      rl.Shader
}

func InitGame() Game {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(ScreenWidth, ScreenHeight, "Protahovatsi Stroj - Voxel Game")

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

	shader := rl.LoadShader("shaders/lighting.vs", "shaders/fog.fs")

	// Locations (do not index shader.Locs)
	locFogDensity := rl.GetShaderLocation(shader, "fogDensity")

	// Fog
	fogDensity := float32(0.009)
	rl.SetShaderValue(shader, locFogDensity, []float32{fogDensity}, rl.ShaderUniformFloat)

	// Load .vox models
	for i := 0; i < len(pkg.PlantModels); i++ {
		pkg.PlantModels[i] = rl.LoadModel(fmt.Sprintf("assets/plants/plant_%d.vox", i))

		// Garantir material válido e aplicar shader
		if pkg.PlantModels[i].MaterialCount == 0 || pkg.PlantModels[i].Materials == nil {
			def := rl.LoadMaterialDefault()
			pkg.PlantModels[i].MaterialCount = 1
			pkg.PlantModels[i].Materials = &def
		}
		(*pkg.PlantModels[i].Materials).Shader = shader
	}

	//LightPosition := rl.NewVector3(0, 6, 0)

	chunkCache := world.NewChunkCache() // Initialize ChunkCache

	// Creates the first chunk at the origin
	originCoord := world.ChunkCoord{X: 0, Y: 0, Z: 0}
	originPos := rl.NewVector3(0, 0, 0)

	chunkCache.Active[originCoord] = world.GenerateTerrainChunk(originPos, perlinNoise, chunkCache, nil, false, nil, false)

	rl.SetTargetFPS(120)

	return Game{
		Camera:      camera,
		CameraMode:  cameraMode,
		ChunkCache:  chunkCache,
		PerlinNoise: perlinNoise,
		Shader:      shader,
	}
}
