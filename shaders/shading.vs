#version 330

in vec3 vertexPosition;
in vec4 vertexColor;
in vec2 vertexTexCoord;

uniform mat4 m_proj;
uniform mat4 m_view;
uniform mat4 m_model;

out vec4 fragColor;
out float fragAO;

void main() {
    gl_Position = m_proj * m_view * m_model * vec4(vertexPosition, 1.0);
    fragColor = vertexColor;
    fragAO = vertexTexCoord.x; // usa o AO que você colocou em Texcoords
}
