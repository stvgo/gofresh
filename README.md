# 🥬 GoFresh

GoFresh es una herramienta simple para desarrolladores de Go que detecta cambios en archivos y recompila automáticamente tus aplicaciones.

## Características

- 🔄 Recarga automática cuando detecta cambios en archivos Go
- 🔍 Monitoreo configurable de múltiples tipos de archivos importantes en microservicios:
  - Archivos Go (`.go`)
  - Configuración (`.env`, `.yaml`, `.yml`, `.json`, `.toml`, `.ini`)
  - Definiciones de API (`.proto`, `.graphql`)
  - Esquemas de base de datos (`.sql`)
  - Archivos de dependencias (`go.mod`)
  - Configuración de Docker (`Dockerfile`, `docker-compose.yml`)
- 🚫 Ignora directorios que no necesitas observar
- ⚡ Liviano y eficiente, sin consumir recursos excesivos
- 🛠️ Funciona de manera optimizada en Windows y sistemas Unix
- 📊 Muestra estadísticas de rendimiento de compilación

## Instalación

```bash
go install github.com/stvgo/gofresh@latest
```

## Uso

Posiciónate en el directorio de tu proyecto y ejecuta:

```bash
gofresh
```

### Opciones

```
-d, --dir string        Directorio a observar para cambios (default directorio actual)
-b, --build string      Comando para compilar (default "go build -o app")
-r, --run string        Comando para ejecutar la aplicación (default "app" en Windows, "./app" en Unix)
-e, --ext string        Extensiones de archivo a observar (default ".go,.env,.yaml,.yml,.json,.toml,.ini,.proto,.sql,.graphql,.mod")
-i, --ignore strings    Directorios a ignorar (default [".git","node_modules","vendor",".cursor","tmp","dist","build",".bin",".cache"])
-v, --verbose           Mostrar información detallada
-t, --debounce duration Tiempo de espera entre detección y compilación (default 300ms en Unix, 500ms en Windows)
-V, --version           Muestra la versión actual de GoFresh
```

## Ejemplos

Monitorear solo archivos específicos:
```bash
gofresh --ext ".go,.env,.yaml"
```

Comando de compilación personalizado:
```bash
gofresh --build "go build -o miapp cmd/main.go" --run "miapp"
```

Observar un directorio específico de microservicios:
```bash
gofresh --dir ./services/auth
```

Ignorar directorios adicionales:
```bash
gofresh --ignore ".git,node_modules,vendor,.cursor,logs,data"
```

## Uso en microservicios

GoFresh es ideal para desarrollo de microservicios en Go:

1. Cada microservicio puede tener su propia instancia de GoFresh
2. Monitorea archivos de configuración distribuida como `.env`, `.yaml` o `.json`
3. Detecta cambios en definiciones de API como archivos Protocol Buffers (`.proto`)
4. Recarga cuando hay cambios en esquemas SQL o configuraciones de GraphQL

## Compatibilidad y optimización multiplataforma

GoFresh detecta automáticamente el sistema operativo y se ajusta para un rendimiento óptimo:

- Usa la sintaxis de comandos correcta según el sistema (`cmd /C` en Windows, `sh -c` en Unix)
- Optimiza el tiempo de debounce según el sistema operativo (500ms en Windows, 300ms en Unix)
- Detecta y elimina eventos duplicados del sistema de archivos para evitar recompilaciones innecesarias
- Muestra estadísticas de tiempo de compilación para ayudar a optimizar el ciclo de desarrollo
- Utiliza los mecanismos adecuados de terminación de procesos para cada plataforma

## Licencia

[MIT](LICENSE)

## Contribuciones

Las contribuciones son bienvenidas. Por favor, abre un issue o pull request en [GitHub](https://github.com/stvgo/gofresh).

