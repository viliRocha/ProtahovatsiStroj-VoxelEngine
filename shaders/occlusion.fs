#version 330 core
in vec3 fragNormal;

out vec4 finalColor;

uniform vec3 lightDir;
uniform vec4 voxelColor;

void main() {
    vec3 normal = normalize(fragNormal);
    float intensity = max(dot(normal, -lightDir), 0.2);  // Adicionando iluminação mínima
    finalColor = voxelColor * intensity;
}
