#version 110

// Batched text and plain rectangles, sampled out of the glyph atlas. The atlas
// holds coverage only, so the colour arrives with each vertex - that is what
// lets differently coloured labels and fills share one draw.
uniform sampler2D tex;

varying vec2 fragTexCoord;
varying vec4 fragColor;

void main() {
    float coverage = texture2D(tex, fragTexCoord).a;
    if (coverage <= 0.0)
        discard;

    // Premultiplied, as text was before the atlas: straight alpha would also
    // blend the framebuffer's alpha with itself, and a translucent window would
    // show the desktop through the edges of every glyph.
    float alpha = fragColor.a * coverage;
    gl_FragColor = vec4(fragColor.rgb * alpha, alpha);
}
