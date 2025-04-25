module github.com/stvgo/gofresh

go 1.24.0

require github.com/fsnotify/fsnotify v1.9.0

require golang.org/x/sys v0.13.0 // indirect

// Nota: 'golang.org/x/sys' es una dependencia indirecta de fsnotify, go mod tidy la manejará.
// Forzar fsnotify como directa es bueno para la claridad.
