package pkg

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var PlantModels [4]rl.Model

const (
	ChunkHeight        int     = 80
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
	//Materials rl.Material
	Voxels    [ChunkSize][ChunkSize][ChunkSize]VoxelData
	Neighbors [6]*Chunk // 0: +X, 1: -X, 2: +Y, 3: -Y, 4: +Z, 5: -Z
	Plants    []PlantData

	Mesh       rl.Mesh
	Model      rl.Model
	IsOutdated bool // Flag to know if you need to update the mesh
	HasMesh    bool // Flag to know if the mesh has already been created
}

type Coords struct {
	X int
	Y int
	Z int
}

var FaceDirections = []rl.Vector3{
	{1, 0, 0},  // Front
	{-1, 0, 0}, // Back
	{0, 1, 0},  // Left
	{0, -1, 0}, // Right
	{0, 0, 1},  // Top
	{0, 0, -1}, // Bottom
}

var FaceVertices = [6][4][3]float32{
	// Face 0: Right (+X)
	{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
	// Face 1: Left (-X)
	{{0, 0, 1}, {0, 1, 1}, {0, 1, 0}, {0, 0, 0}},
	// Face 2: Top (+Y)
	{{0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}},
	// Face 3: Bottom (-Y)
	{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
	// Face 4: Front (+Z)
	{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	// Face 5: Back (-Z)
	{{1, 0, 0}, {0, 0, 0}, {0, 1, 0}, {1, 1, 0}},
}

// Returns the vertices, normals and UVs of a voxel face
func GetFaceGeometry(faceIndex int, x, y, z float32) ([]rl.Vector3, []rl.Vector3, []rl.Vector2) {
	px, py, pz := x, y, z

	// Relative positions of face vertices
	var vertices []rl.Vector3
	var normal rl.Vector3

	switch faceIndex {
	case 0: // Right (+X)
		vertices = []rl.Vector3{
			{px + 1, py, pz},
			{px + 1, py + 1, pz},
			{px + 1, py + 1, pz + 1},
			{px + 1, py, pz + 1},
		}
		normal = rl.NewVector3(1, 0, 0)
	case 1: // Left (-X)
		vertices = []rl.Vector3{
			{px, py, pz + 1},
			{px, py + 1, pz + 1},
			{px, py + 1, pz},
			{px, py, pz},
		}
		normal = rl.NewVector3(-1, 0, 0)
	case 2: // Top (+Y)
		vertices = []rl.Vector3{
			{px, py + 1, pz},
			{px, py + 1, pz + 1},
			{px + 1, py + 1, pz + 1},
			{px + 1, py + 1, pz},
		}
		normal = rl.NewVector3(0, 1, 0)
	case 3: // Bottom (-Y)
		vertices = []rl.Vector3{
			{px, py, pz},
			{px + 1, py, pz},
			{px + 1, py, pz + 1},
			{px, py, pz + 1},
		}
		normal = rl.NewVector3(0, -1, 0)
	case 4: // Front (+Z)
		vertices = []rl.Vector3{
			{px + 1, py, pz + 1},
			{px + 1, py + 1, pz + 1},
			{px, py + 1, pz + 1},
			{px, py, pz + 1},
		}
		normal = rl.NewVector3(0, 0, 1)
	case 5: // Back (-Z)
		vertices = []rl.Vector3{
			{px, py, pz},
			{px, py + 1, pz},
			{px + 1, py + 1, pz},
			{px + 1, py, pz},
		}
		normal = rl.NewVector3(0, 0, -1)
	}

	// Normals (1 per vertex)
	normals := []rl.Vector3{normal, normal, normal, normal}

	// Default texture coordinates (UV)
	texcoords := []rl.Vector2{
		{0, 1},
		{0, 0},
		{1, 0},
		{1, 1},
	}

	return vertices, normals, texcoords
}

func Transform(val, min, max int) (int, bool) {
	if val < min {
		return max, false
	}
	if val > max {
		return min, false
	}
	return val, true
}
