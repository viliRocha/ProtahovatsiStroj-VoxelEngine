package render

import (
	"math/rand"
	"unsafe"

	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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
	Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.ChunkHeight), int(pkg.ChunkSize)
	for i := 0; i < Nx*Ny*Nz; i++ {
		pos := pkg.Coords{
			X: i / (Ny * Nz),
			Y: (i / Nz) % Ny,
			Z: i % Nz,
		}

		voxel := chunk.Voxels[pos.X][pos.Y][pos.Z]
		block := world.BlockTypes[voxel.Type]

		if !block.IsVisible || voxel.Type == "Water" || voxel.Type == "Plant" {
			continue
		}

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

			for i := 0; i < 4; i++ {
				v := pkg.FaceVertices[face][i]
				vertices = append(vertices,
					float32(pos.X)+v[0],
					float32(pos.Y)+v[1],
					float32(pos.Z)+v[2],
				)

				c := block.Color
				// Add color per vertex (RGBA)
				colorModifier := uint8(rand.Intn(8)) //uint8(y * 20)

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
	direction, max_size, max_height :=
		pkg.FaceDirections[faceIndex], int(pkg.ChunkSize-1), int(pkg.ChunkHeight)-1

	var vX bool
	var vY bool
	var vZ bool

	// Calculates the new coordinates based on the face direction
	pos.X, vX = pkg.Transform(int(pos.X+int(direction.X)), 0, max_size)
	pos.Y, vY = pkg.Transform(int(pos.Y+int(direction.Y)), 0, max_height)
	pos.Z, vZ = pkg.Transform(int(pos.Z+int(direction.Z)), 0, max_size)

	// Checks if the new coordinates are within the chunk bounds
	if vX && vY && vZ {
		voxel_type := chunk.Voxels[pos.X][pos.Y][pos.Z].Type
		return !world.BlockTypes[voxel_type].IsSolid
	}

	neighbor_index := chunk.Neighbors[faceIndex]

	if neighbor_index == nil {
		return false
	}

	return !world.BlockTypes[neighbor_index.Voxels[pos.X][pos.Y][pos.Z].Type].IsSolid
}
