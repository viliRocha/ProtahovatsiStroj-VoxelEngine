#version 330

in vec4 fragColor;
in float fragAO;

out vec4 finalColor;

void main() {
    finalColor = vec4(fragColor.rgb * fragAO, fragColor.a);
}
