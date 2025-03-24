# PtotahovatsiStroj
### v0.8
<center><img src="./img/engine_logo.png" alt="Logo" width="100%"></center>

A voxel engine built using **raylib-go** and **OpenGL**

##  Features 🌟
- **Infinite Random World Generation**: Utilizes a Perlin noise algorithm for creating expansive landscapes.
- **Dynamic Lighting**: Implemented basic lighting with shading and ambient occlusion for better depth perception.
- **Water Formations**: Realistic water bodies.
- **Surface Feature System**: Precedurally generated trees and randomly placed flowers and tall grass.
- **Cache System**: Efficiently store the position of surface features for better world consistency.

## Upcoming Features 📋
- **Layered Landscapes**: Add more geological layers, such as stone.
- **Fog Effects**: Create atmospheric depth with fog.
- **Block Interaction**: Enable players to place and destroy blocks.
- **Biome Diversity**: Implement various biomes for a richer exploration experience.
- **Cave Generation**: Create intricate cave systems for players to discover.
- **Web Build**: Compile the project to WASM.

## Screenshots 🖼️
<img src="./img/lighting.png" alt="light" width="1000px">
<img src="./img/landscape1.png" alt="world_gen" width="1000px">
<img src="./img/landscape2.png" alt="world_gen" width="1000px">

## Getting Started 🚀
To get started with the voxel engine, clone the repository and open the folder. Make sure you have Go installed on your device. Then, run the command `go mod tidy` and finally, to compile the project, run `go run ./src`.

## License 📄
This project is licensed under the MIT License - see the `LICENSE` file for details.

## Acknowledgments 🙏
Inspired by [CubeWorld](https://store.steampowered.com/app/1128000/Cube_World/) and [TanTan's](https://github.com/TanTanDev) Voxel engine built with the beavy engine. I would also like to acknowledge the use of voxelized vegetation assets from [MangoVoxel](https://mangovoxel.itch.io/voxelfoliage) in my voxel engine. <!--Special thanks to the resources and tutorials that helped shape this project.-->
