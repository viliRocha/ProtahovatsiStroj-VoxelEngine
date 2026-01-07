#version 330

in vec3 fragPosition;
in vec3 fragNormal;
in vec4 fragColor;   // alpha = AO factor

uniform vec4 colDiffuse;

out vec4 finalColor;

void main()
{
    // Base color
    vec3 baseRGB = (colDiffuse.rgb * fragColor.rgb);

    // AO factor baked in vertex alpha (clamp for safety)
    float ao = clamp(fragColor.a, 0.0, 1.0);

    // Apply AO (darken by occlusion)
    vec3 shaded = baseRGB * ao;

    finalColor = vec4(shaded, 1.0);
}
