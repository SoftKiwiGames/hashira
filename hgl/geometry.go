package hgl

var QuadVertices = []float32{
	-1.0, -1.0, 0.0, // Bottom-left corner
	1.0, -1.0, 0.0, // Bottom-right corner
	-1.0, 1.0, 0.0, // Top-left corner
	1.0, -1.0, 0.0, // Bottom-right corner
	1.0, 1.0, 0.0, // Top-right corner
	-1.0, 1.0, 0.0, // Top-left corner
}

var QuadUV = []float32{
	0.0, 0.0, // Bottom-left corner
	1.0, 0.0, // Bottom-right corner
	0.0, 1.0, // Top-left corner
	1.0, 0.0, // Bottom-right corner
	1.0, 1.0, // Top-right corner
	0.0, 1.0, // Top-left corner
}
