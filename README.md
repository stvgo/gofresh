# GoFresh - Monitor de cambios y recompilación automática para Go

## Descripción del proyecto
GoFresh es una herramienta de desarrollo para Go que monitorea cambios en archivos de código fuente y automáticamente recompila y redespliega tu aplicación. Similar a Air pero con enfoque en simplicidad, velocidad y configuración mínima.

## Características clave
- Tiempo de inicio rápido
- Mínima configuración requerida (funciona out-of-the-box)
- Soporte para múltiples directorios de monitoreo
- Filtrado configurable de archivos y directorios
- Diferentes modos de recompilación (completa vs incremental)
- Soporte de variables de entorno
- Notificaciones de escritorio opcionales

## Requisitos
- Go 1.24 o superior
- Dependencias gestionadas con Go Modules

## Instalación

1. Clona el repositorio:
   ```bash
   git clone https://github.com/stvgo/gofresh.git
   cd gofresh
   ```

2. Instala las dependencias:
   ```bash
   go mod tidy
   ```

3. Compila e instala la herramienta:
   ```bash
   go install
   ```

## Uso básico

