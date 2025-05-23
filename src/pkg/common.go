package pkg

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var PlantModels [4]rl.Model

const (
	ChunkHeight        int     = 96
	ChunkSize          int     = 32
	ChunkDistance      int     = 1
	WaterLevelFraction float64 = 0.375 // 3/8
)

type VoxelData struct {
	Type  string
	Model rl.Model
}

// Stores plant positions
type PlantData struct {
	Position rl.Vector3
	ModelID  int
}

type Chunk struct {
	Voxels    [ChunkSize][ChunkHeight][ChunkSize]VoxelData
	Neighbors [6]*Chunk // 0: +X, 1: -X, 2: +Y, 3: -Y, 4: +Z, 5: -Z
	Plants    []PlantData
}

var FaceDirections = []rl.Vector3{
	{1, 0, 0},  // Front
	{-1, 0, 0}, // Back
	{0, 1, 0},  // Left
	{0, -1, 0}, // Right
	{0, 0, 1},  // Top
	{0, 0, -1}, // Bottom
}

var FaceScales = []rl.Vector3{
	{1.0, 1.0, 0.0}, // Front
	{1.0, 1.0, 0.0}, // Back
	{0.0, 1.0, 1.0}, // Left
	{0.0, 1.0, 1.0}, // Right
	{1.0, 0.0, 1.0}, // Top
	{1.0, 0.0, 1.0}, // Bottom
}
