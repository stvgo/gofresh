package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
)

var (
	watchDir      string
	runCmd        string
	buildDebounce time.Duration
	cmd           *exec.Cmd
	version       = "0.1.0" // Versión de la aplicación
	verbose       bool
	selfPath      string // Ruta del propio ejecutable
)

func init() {
	// Obtener el directorio actual
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error obteniendo directorio actual: %v", err)
	}

	// Guardar el path absoluto de este archivo
	ex, err := os.Executable()
	if err == nil {
		selfPath, _ = filepath.Abs(ex)
	}

	// Si no tenemos el ejecutable, intentar con el archivo actual
	if selfPath == "" {
		selfFile, _ := filepath.Abs(filepath.Join(currentDir, "main.go"))
		if _, err := os.Stat(selfFile); err == nil {
			selfPath = selfFile
		}
	}

	// Definir las flags
	pflag.StringVarP(&watchDir, "dir", "d", currentDir, "Directorio a observar para cambios")
	pflag.StringVarP(&runCmd, "run", "r", "go run .", "Comando para ejecutar la aplicación")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Mostrar información detallada")
	pflag.DurationVarP(&buildDebounce, "debounce", "t", 500*time.Millisecond, "Tiempo de espera entre detección y compilación")

	// Versión
	showVersion := pflag.BoolP("version", "V", false, "Muestra la versión actual de GoFresh")

	pflag.Parse()

	// Mostrar versión y salir si se solicitó
	if *showVersion {
		fmt.Printf("GoFresh versión %s\n", version)
		os.Exit(0)
	}
}

func main() {
	// Verificar si hay argumentos
	if len(os.Args) > 1 {
		// Si el primer argumento después del nombre del programa es "init"
		if os.Args[1] == "init" {
			initProject()
			return
		}
	}

	// Modo normal (monitorear)
	startMonitoring()
}

// Inicialización del proyecto
func initProject() {
	fmt.Println("🥬 GoFresh - Inicializando proyecto...")

	// Verificar si ya hay un archivo go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("🔍 No se encontró go.mod, inicializando módulo...")
		moduleName := askForInput("Nombre del módulo (ej: github.com/usuario/proyecto): ")

		// Ejecutar go mod init
		initCmd := exec.Command("go", "mod", "init", moduleName)
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		if err := initCmd.Run(); err != nil {
			fmt.Printf("❌ Error inicializando módulo: %v\n", err)
			return
		}
		fmt.Println("✅ Módulo inicializado correctamente")
	} else {
		fmt.Println("✅ El módulo ya está inicializado (go.mod existe)")
	}

	// Verificar si existe main.go
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("🔍 No se encontró main.go, creando archivo de ejemplo...")

		// Preguntar por el tipo de proyecto
		fmt.Println("Selecciona el tipo de proyecto:")
		fmt.Println("1) API REST con Gin (recomendado)")
		fmt.Println("2) Aplicación CLI simple")
		fmt.Println("3) Servidor HTTP básico")

		choice := askForInput("Selección (1-3): ")

		// Crear archivo main.go según la selección
		var template string
		switch choice {
		case "1":
			template = `package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	
	r.Run() // escucha en 0.0.0.0:8080
}
`
			// Agregar gin como dependencia
			fmt.Println("📦 Instalando dependencia: gin-gonic/gin...")
			getCmd := exec.Command("go", "get", "github.com/gin-gonic/gin")
			getCmd.Run()

		case "2":
			template = `package main

import (
	"fmt"
	"os"
)

func main() {
	// Verificar argumentos
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./app comando [argumentos]")
		os.Exit(1)
	}
	
	// Procesar comando
	cmd := os.Args[1]
	switch cmd {
	case "hola":
		fmt.Println("¡Hola, mundo!")
	default:
		fmt.Printf("Comando desconocido: %s\\n", cmd)
	}
}
`
		case "3":
			template = `package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "¡Hola, %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Servidor iniciado en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
`
		default:
			template = `package main

import "fmt"

func main() {
	fmt.Println("¡Hola, mundo!")
}
`
		}

		// Escribir el archivo
		if err := os.WriteFile("main.go", []byte(template), 0644); err != nil {
			fmt.Printf("❌ Error creando main.go: %v\n", err)
			return
		}
		fmt.Println("✅ Archivo main.go creado correctamente")
	} else {
		fmt.Println("✅ El archivo main.go ya existe")
	}

	fmt.Println("\n🚀 Proyecto inicializado correctamente. Para comenzar a desarrollar ejecuta:")
	fmt.Println("    gofresh")
	fmt.Println("\nEsto monitoreará cambios en tus archivos .go y ejecutará automáticamente tu aplicación.")
}

// Función para pedir input al usuario
func askForInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

// Inicia el monitoreo de cambios
func startMonitoring() {
	fmt.Println("🥬 GoFresh v" + version + " - Recarga automática para aplicaciones Go")
	fmt.Printf("Observando cambios en archivos .go en: %s\n", watchDir)
	fmt.Printf("Comando de ejecución: %s\n", runCmd)

	// Verificar que no estamos monitoreando nuestro propio código
	if isSelfMonitoring() {
		fmt.Println("⚠️  ADVERTENCIA: Parece que estás monitoreando el directorio que contiene GoFresh.")
		fmt.Println("   Esto puede causar bucles infinitos. Se recomienda usar GoFresh en otro directorio.")
		fmt.Println("   Si realmente quieres continuar, presiona Enter. Para salir, presiona Ctrl+C.")
		fmt.Scanln() // Esperar confirmación
	}

	// Crear el watcher para monitorear cambios en el sistema de archivos
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creando watcher: %v", err)
	}
	defer watcher.Close()

	// Manejar señales para limpiar recursos
	done := make(chan bool)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Agregar los directorios a monitorear
	if err := addDirsToWatch(watcher, watchDir); err != nil {
		log.Fatalf("Error agregando directorios: %v", err)
	}

	// Ejecutar inicialmente
	runApp()

	// Monitorear cambios
	go watchChanges(watcher)

	// Manejar señales
	go handleSignals(sigs, done)

	<-done
	fmt.Println("\n👋 Saliendo de GoFresh...")
}

// Verifica si estamos monitoreando nuestro propio código
func isSelfMonitoring() bool {
	// Si estamos en el directorio de GoFresh
	mainFile := filepath.Join(watchDir, "main.go")
	if _, err := os.Stat(mainFile); err == nil {
		// Leer el archivo para ver si es GoFresh
		content, err := os.ReadFile(mainFile)
		if err == nil {
			return strings.Contains(string(content), "GoFresh") &&
				strings.Contains(string(content), "fsnotify")
		}
	}
	return false
}

func addDirsToWatch(watcher *fsnotify.Watcher, root string) error {
	dirsCount := 0
	ignoreDirs := []string{".git", "node_modules", "vendor", ".cursor", "tmp", "dist", "build", ".bin", ".cache"}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Si es un directorio
		if info.IsDir() {
			// Verificar si es un directorio a ignorar
			base := filepath.Base(path)
			for _, ignoreDir := range ignoreDirs {
				if base == ignoreDir {
					if verbose {
						fmt.Printf("Ignorando directorio: %s\n", path)
					}
					return filepath.SkipDir
				}
			}

			// Agregar el directorio al watcher
			if err := watcher.Add(path); err != nil {
				return err
			}
			dirsCount++
			if verbose {
				fmt.Printf("Observando directorio: %s\n", path)
			}
		}
		return nil
	})

	fmt.Printf("Total de directorios monitoreados: %d\n", dirsCount)
	return err
}

func watchChanges(watcher *fsnotify.Watcher) {
	throttled := false
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Solo nos interesa cuando un archivo es creado, modificado o eliminado
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Ignorar nuestro propio archivo para evitar bucles infinitos
				if selfPath != "" && strings.Contains(event.Name, selfPath) {
					if verbose {
						fmt.Printf("Ignorando cambio en el propio código de GoFresh: %s\n", event.Name)
					}
					continue
				}

				// Solo monitorear archivos .go y archivos de dependencias importantes
				shouldTrigger := strings.HasSuffix(strings.ToLower(event.Name), ".go") ||
					strings.HasSuffix(event.Name, "go.mod") ||
					strings.HasSuffix(event.Name, "go.sum")

				if shouldTrigger && !throttled {
					throttled = true
					go func() {
						// Mostrar información del cambio
						fmt.Printf("🔄 Cambio detectado en: %s\n", event.Name)

						// Esperar un poco para debounce (evitar ejecuciones frecuentes)
						time.Sleep(buildDebounce)

						// Ejecutar
						runApp()

						// Desbloquear throttling
						throttled = false
					}()
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error en watcher: %v", err)
		}
	}
}

func runApp() {
	// Si hay un proceso en ejecución, terminarlo
	if cmd != nil && cmd.Process != nil {
		fmt.Println("🛑 Deteniendo proceso anterior...")

		// Intentar terminar el proceso de la manera más simple posible
		cmd.Process.Kill()
		cmd.Wait()
	}

	// Ejecutar la aplicación directamente con go run
	fmt.Println("🚀 Iniciando aplicación...")
	cmd = exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Error iniciando aplicación: %v\n", err)
		return
	}

	fmt.Println("✅ Aplicación en ejecución")
}

func handleSignals(sigs chan os.Signal, done chan bool) {
	<-sigs

	// Matar el proceso hijo si existe
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}

	// Señalar terminación
	done <- true
}
