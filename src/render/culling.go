package render

import (
	"go-engine/src/pkg"
	"go-engine/src/world"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Função auxiliar: calcula fator de occlusion para um vértice
func calculateVoxelAO(chunk *pkg.Chunk, pos pkg.Coords, face int, corner int) float32 {
	// Cada vértice da face tem 3 vizinhos potenciais (duas arestas + um diagonal)
	// Se todos são sólidos, o vértice fica mais escuro.
	offsets := pkg.VertexAOOffsets[face][corner] // tabela de offsets (dx,dy,dz) para cada vizinho
	occlusion := 0
	for _, off := range offsets {
		nx := pos.X + off[0]
		ny := pos.Y + off[1]
		nz := pos.Z + off[2]
		if nx >= 0 && nx < pkg.ChunkSize &&
			ny >= 0 && ny < pkg.ChunkSize &&
			nz >= 0 && nz < pkg.ChunkSize {
			if world.BlockTypes[chunk.Voxels[nx][ny][nz].Type].IsSolid {
				occlusion++
			}
		} else {
			// fora do chunk → trata como sólido para não deixar borda clara demais
			occlusion++
		}
	}

	// Normaliza: 0 vizinhos sólidos = claro, 3 vizinhos sólidos = escuro
	return 0.7 + 0.5*(1.0-float32(occlusion)/3.0)
}

func calculateFaceAO(chunk *pkg.Chunk, pos pkg.Coords, face int) float32 {
	total := 0.0
	for corner := 0; corner < 4; corner++ {
		total += float64(calculateVoxelAO(chunk, pos, face, corner))
	}
	// média dos 4 cantos
	return float32(total / 4.0)
}

func BuildChunkMesh(chunk *pkg.Chunk, chunkPos rl.Vector3 /*, lightPosition rl.Vector3*/) {
	var vertices []float32
	var indices []uint16
	var colors []uint8

	indexOffset := uint16(0)

	/* Multidimensional Arrays Linearization, docs and extras that may come in handy
	 * https://ic.unicamp.br/~bit/mc102/aulas/aula15.pdf (introdução)
	 * https://felippe.ubi.pt/texts3/contr_av_ppt01p.pdf (pág. 13)
	 * https://www.aussieai.com/book/ch36-linearized-multidimensional-arrays
	 * https://teotl.dev/vischunk/ (may be useful)
	 * (AI was used to help the interpretation of some of those docs)
	 */

	Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.ChunkSize), int(pkg.ChunkSize)
	for i := 0; i < Nx*Ny*Nz; i++ {
		pos := pkg.Coords{
			X: i / (Ny * Nz),
			Y: (i / Nz) % Ny,
			Z: i % Nz,
		}

		voxel := chunk.Voxels[pos.X][pos.Y][pos.Z]
		block := world.BlockTypes[voxel.Type]

		if !block.IsVisible || voxel.Type == "Water" || voxel.Type == "Plant" || voxel.Type == "Cloud" {
			continue
		}

		c := block.Color
		// Add color per block (RGBA) on a deterministic way (so it can be recalculated when a new chunk is created and still look the same)
		colorModifier := uint8(
			((pos.X*73856093 + pos.Y*19349663) ^
				(pos.Z*83492791 + pos.X*19349663) ^
				(pos.Y*83492791 + pos.Z*73856093)) % 8)

		for face := 0; face < 6; face++ {
			if !shouldDrawFace(chunk, pos, face) && voxel.Type != "Leaves" {
				continue
			}

			/*
				voxelPosition := rl.NewVector3(
					chunkPos.X+float32(x),
					chunkPos.Y+float32(y),
					chunkPos.Z+float32(z),
				)

				lightIntensity := calculateLightIntensity(voxelPosition, lightPosition)
				voxelColor := applyLighting(block.Color, lightIntensity)
			*/

			//ao := calculateFaceAO(chunk, pos, face)

			for vertice := 0; vertice < 4; vertice++ {
				v := pkg.FaceVertices[face][vertice]
				vertices = append(vertices,
					float32(pos.X)+v[0],
					float32(pos.Y)+v[1],
					float32(pos.Z)+v[2],
				)

				/*
					colors = append(colors,
						uint8(float32(c.R+colorModifier)*ao),
						uint8(float32(c.G+colorModifier)*ao),
						uint8(float32(c.B+colorModifier)*ao),
						c.A,
					)
				*/

				colors = append(colors, c.R+colorModifier, c.G+colorModifier, c.B+colorModifier, c.A)
			}

			//	Add the two triangles of the face
			indices = append(indices,
				indexOffset, indexOffset+1, indexOffset+2,
				indexOffset, indexOffset+2, indexOffset+3,
			)
			indexOffset += 4
		}
	}

	mesh := rl.Mesh{
		VertexCount:   int32(len(vertices) / 3),
		TriangleCount: int32(len(indices) / 3),
	}

	if len(vertices) > 0 {
		mesh.Vertices = (*float32)(unsafe.Pointer(&vertices[0]))
	}
	if len(indices) > 0 {
		mesh.Indices = (*uint16)(unsafe.Pointer(&indices[0]))
	}
	if len(colors) > 0 {
		mesh.Colors = (*uint8)(unsafe.Pointer(&colors[0]))
	}
	/*
		if len(aoValues) > 0 {
			// envia AO como "Texcoords" (cada vértice recebe um valor)
			mesh.Texcoords = (*float32)(unsafe.Pointer(&aoValues[0]))
		}
	*/

	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)

	// Create material and assign it
	material := rl.LoadMaterialDefault()
	materials := []rl.Material{material}
	model.MaterialCount = int32(len(materials))
	model.Materials = &materials[0]

	// Assign to chunk
	chunk.Model = model
	chunk.HasMesh = true
	chunk.IsOutdated = false
}

func shouldDrawFace(chunk *pkg.Chunk, pos pkg.Coords, faceIndex int) bool {
	direction := pkg.FaceDirections[faceIndex]
	maxSize := int(pkg.ChunkSize - 1)
	maxHeight := int(pkg.ChunkSize) - 1

	// Calculates the new coordinates based on the face direction
	nx := pos.X + int(direction.X)
	ny := pos.Y + int(direction.Y)
	nz := pos.Z + int(direction.Z)

	// Checks if the new coordinates are within the chunk bounds
	if nx >= 0 && nx <= maxSize &&
		ny >= 0 && ny <= maxHeight &&
		nz >= 0 && nz <= maxSize {
		voxelType := chunk.Voxels[nx][ny][nz].Type
		return !world.BlockTypes[voxelType].IsSolid
	}

	// Outside chunk boundries → depends on the neighbor
	neighbor := chunk.Neighbors[faceIndex]
	if neighbor == nil {
		return true // no neighbor → exposed face
	}

	// Adjusts coordinates relative to the neighbor.
	nx = (nx + int(pkg.ChunkSize)) % int(pkg.ChunkSize)
	ny = (ny + int(pkg.ChunkSize)) % int(pkg.ChunkSize)
	nz = (nz + int(pkg.ChunkSize)) % int(pkg.ChunkSize)

	voxelType := neighbor.Voxels[nx][ny][nz].Type
	return !world.BlockTypes[voxelType].IsSolid
}
