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
	Camera        rl.Camera
	CameraMode    rl.CameraMode
	ChunkCache    *world.ChunkCache
	PerlinNoise   *perlin.Perlin
	Shader        rl.Shader
	LightPosition rl.Vector3
}

func loadShader(camera rl.Camera, chunkCache *world.ChunkCache) rl.Shader {
	shader := rl.LoadShader("shaders/shading.vs", "shaders/shading.fs")
	*shader.Locs = rl.GetShaderLocation(shader, "viewPos")

	ambientLoc := rl.GetShaderLocation(shader, "ambient")
	shaderValue := []float32{0.1, 0.1, 0.1, 1.0}
	rl.SetShaderValue(shader, ambientLoc, shaderValue, rl.ShaderUniformVec4)

	cameraPos := []float32{camera.Position.X, camera.Position.Y, camera.Position.Z}
	rl.SetShaderValue(shader, *shader.Locs, cameraPos, rl.ShaderUniformVec3)

	lights := make([]Light, 4)
	lights[0] = NewLight(LightTypePoint, rl.NewVector3(-2, 1, -2), rl.NewVector3(0, 0, 0), rl.Yellow, shader)

	rl.SetShaderValue(shader, *shader.Locs, cameraPos, rl.ShaderUniformVec3)

	return shader
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

	// Load .vox models
	for i := range 4 {
		pkg.PlantModels[i] = rl.LoadModel(fmt.Sprintf("assets/plants/plant_%d.vox", i))
	}

	LightPosition := rl.NewVector3(0, 6, 0)

	chunkCache := world.NewChunkCache() // Initialize ChunkCache
	shader := loadShader(camera, chunkCache)

	// Cria o primeiro chunk na origem
	originCoord := world.ChunkCoord{X: 0, Y: 0, Z: 0}
	originPos := rl.NewVector3(0, 0, 0)

	chunkCache.Active[originCoord] = world.GenerateTerrainChunk(originPos, perlinNoise, nil, false)

	rl.SetTargetFPS(60)

	return Game{
		Camera:        camera,
		CameraMode:    cameraMode,
		ChunkCache:    chunkCache,
		PerlinNoise:   perlinNoise,
		Shader:        shader,
		LightPosition: LightPosition,
	}
}
