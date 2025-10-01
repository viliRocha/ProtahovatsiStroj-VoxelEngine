package load

import (
	"fmt"
	"math/rand"

    "go-engine/src/pkg"
    "go-engine/src/world"

    "github.com/aquilax/go-perlin"
    rl "github.com/gen2brain/raylib-go/raylib"
)

type Game struct {
	Camera           rl.Camera
	CameraMode       rl.CameraMode
	ChunkCache       *world.ChunkCache
	PerlinNoise      *perlin.Perlin
	Shader           rl.Shader
	LightPosition    rl.Vector3
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
    rl.SetConfigFlags(rl.FlagMsaa4xHint)
    rl.InitWindow(pkg.ScreenWidth, pkg.ScreenHeight, "Protahovatsi Stroj - Voxel Game")

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
    perlinNoise := perlin.NewPerlin(pkg.PerlinAlpha, pkg.PerlinBeta, pkg.PerlinN, seed)

    // Load .vox models
    for i := range 4 {
        pkg.PlantModels[i] = rl.LoadModel(fmt.Sprintf("assets/plants/plant_%d.vox", i))
    }

    LightPosition := rl.NewVector3(0, 6, 0)

    chunkCache := world.NewChunkCache()    // Initialize ChunkCache
    shader := loadShader(camera, chunkCache)

    chunkCache.Chunks[rl.NewVector3(0, 0, 0)] = world.GenerateAbovegroundChunk(rl.NewVector3(0, 0, 0), perlinNoise, false) // Passing perlinNoise

	rl.SetTargetFPS(60)

    return Game{
        Camera:          camera,
        CameraMode:      cameraMode,
        ChunkCache:      chunkCache,
        PerlinNoise:     perlinNoise,
        Shader:          shader,
        LightPosition:   LightPosition,
	}
}
