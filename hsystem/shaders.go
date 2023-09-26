package hsystem

const VertexShaderSource = `
attribute vec3 position;
attribute vec2 uv;

varying vec2 vUV;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main(void) {
  gl_Position = projection * view * model * vec4(position, 1.0);
  vUV = uv;
}
`

const FragmentShaderSource = `
precision mediump float;

varying vec2 vUV;
 
uniform sampler2D tileset;

void main(void) {
  gl_FragColor = texture2D(tileset, vUV);
}
`

const QuadVertexShaderSource = `
attribute vec3 position;
attribute vec2 uv;

varying vec2 vUV;

void main(void)
{
    gl_Position = vec4(position, 1.0); 
    vUV = uv;
}`

const QuadFragmentShaderSource = `
precision mediump float;

varying vec2 vUV;

uniform sampler2D quad;

void main()
{ 
    gl_FragColor = texture2D(quad, vUV);
}`
