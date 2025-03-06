#version 330

in vec4 fragColor;
out vec4 finalColor;

uniform float ambientOcclusion;

void main() {
    finalColor = fragColor * ambientOcclusion;
}
