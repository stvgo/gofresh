# 🥬 GoFresh (Desarrollo)

Herramienta simple para desarrolladores Go que detecta cambios en archivos `.go` y recompila y reinicia automáticamente tu aplicación.

## Características (Actuales)

*   🔄 Recarga automática al detectar cambios en archivos `.go`.
*   ⏲️ Debounce para evitar reinicios múltiples rápidos.
*   🚀 Logs con emojis para inicio, reinicio y detención.
*   🧹 Limpieza automática del binario temporal al detenerse.
*   💻 Funciona como herramienta de línea de comandos instalable.
*   🛠️ Manejo de procesos optimizado para Windows y Unix-like.

## Instalación

Asegúrate de tener Go instalado y tu `GOPATH` configurado correctamente.

```bash
go install github.com/stvgo/gofresh/cmd/gofresh@latest
```

## Uso

1.  Navega a la carpeta raíz de tu proyecto Go.
2.  Ejecuta el comando:

    ```bash
    gofresh
    ```

`gofresh` observará los cambios en los archivos `.go` del directorio actual, recompilará y reiniciará tu aplicación automáticamente.

Presiona `Ctrl+C` para detener `gofresh`.

## Opciones (Actuales)

*   `-init`: (Funcionalidad futura) Muestra un mensaje indicando que inicializará la configuración.

## Desarrollo Futuro (Ideas)

*   Configuración de directorios y extensiones a observar.
*   Ignorar directorios específicos.
*   Comandos de build y run personalizables.
*   Observación recursiva de subdirectorios.
*   Mejoras en la gestión de procesos hijos.

## Licencia

MIT 