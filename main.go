package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	currentProcess *os.Process
	processMutex   sync.Mutex
	restartTimer   *time.Timer
	debounceTime   = 500 * time.Millisecond // Tiempo de espera antes de reiniciar
	outputBinary   = "gofresh-app"          // Nombre del binario temporal
)

func main() {
	log.Println("Iniciando gofresh...")
	// Asegurarse de limpiar el binario al salir
	defer cleanup()

	// Configurar watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creando el watcher: %v", err)
	}
	defer watcher.Close()

	// Añadir directorio actual (podría expandirse para añadir más o ser configurable)
	// TODO: Añadir lógica para observar subdirectorios recursivamente si es necesario
	err = watcher.Add(".")
	if err != nil {
		log.Fatalf("Error añadiendo directorio al watcher: %v", err)
	}
	log.Println("Observando cambios en '.'")

	// Compilar y ejecutar por primera vez
	buildAndRun()

	// Goroutine para manejar eventos del watcher
	go watchEvents(watcher)

	// Esperar señal de interrupción (Ctrl+C)
	handleInterrupt()
}

func watchEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return // Canal cerrado
			}
			// log.Printf("Evento raw: %s %s", event.Op, event.Name) // Debug

			// Filtrar eventos relevantes (archivos .go, operaciones de cambio)
			if shouldRestart(event) {
				log.Printf("Cambio detectado en: %s", event.Name)
				scheduleRestart()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return // Canal cerrado
			}
			log.Printf("Error del watcher: %v", err)
		}
	}
}

func shouldRestart(event fsnotify.Event) bool {
	// Ignorar directorios o archivos no Go
	if !strings.HasSuffix(event.Name, ".go") {
		return false
	}
	// Ignorar el propio binario temporal si cae en el watcher
	if filepath.Base(event.Name) == outputBinary {
		return false
	}
	// Reaccionar a escrituras, creaciones o eliminaciones
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
}

func scheduleRestart() {
	processMutex.Lock()
	// Si ya hay un timer, reiniciarlo
	if restartTimer != nil {
		restartTimer.Reset(debounceTime)
		log.Println("Debounce: reinicio reprogramado.")
	} else {
		// Si no hay timer, crear uno nuevo
		log.Println("Debounce: programando reinicio...")
		restartTimer = time.AfterFunc(debounceTime, func() {
			log.Println("¡Tiempo de debounce cumplido! Reiniciando aplicación...")
			stopApp() // Asegura parar antes de construir
			buildAndRun()
			processMutex.Lock() // Bloquear para limpiar el timer
			restartTimer = nil  // Indicar que el timer ya no está activo
			processMutex.Unlock()
		})
	}
	processMutex.Unlock()
}

func stopApp() {
	processMutex.Lock()
	defer processMutex.Unlock()

	if currentProcess == nil {
		// log.Println("No hay proceso activo para detener.")
		return
	}

	log.Printf("Deteniendo proceso anterior (PID: %d)...", currentProcess.Pid)
	err := killProcess(currentProcess)
	if err != nil {
		log.Printf("Error al detener el proceso (PID: %d): %v", currentProcess.Pid, err)
	} else {
		log.Printf("Proceso (PID: %d) detenido.", currentProcess.Pid)
	}
	currentProcess = nil
}

// killProcess intenta detener el proceso de forma robusta según el SO.
func killProcess(p *os.Process) error {
	if runtime.GOOS == "windows" {
		// En Windows, taskkill es más fiable para árboles de procesos
		cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", p.Pid))
		// cmd.Stdout = os.Stdout // Descomentar para debug
		// cmd.Stderr = os.Stderr // Descomentar para debug
		return cmd.Run()
	}
	// En Linux/macOS, intentar matar el grupo de procesos.
	// Esto requiere que el proceso hijo se inicie en su propio grupo (cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true})
	// Por ahora, usamos Kill simple. Podría mejorarse.
	return p.Kill()
}

func buildAndRun() {
	processMutex.Lock() // Bloquear mientras se compila y lanza
	defer processMutex.Unlock()

	log.Println("Compilando aplicación...")
	buildCmd := exec.Command("go", "build", "-o", outputBinary, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		log.Printf("Error en la compilación: %v", err)
		return // No intentar ejecutar si la compilación falló
	}
	log.Println("Compilación exitosa.")

	log.Println("Ejecutando aplicación...")
	runCmd := exec.Command("./" + outputBinary) // Asume que está en el directorio actual
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Configurar atributos específicos del SO usando build constraints
	configureCmdSysProcAttr(runCmd)

	err = runCmd.Start()
	if err != nil {
		log.Printf("Error al iniciar la aplicación: %v", err)
		return
	}

	currentProcess = runCmd.Process
	log.Printf("Aplicación iniciada con PID: %d", currentProcess.Pid)

	// Goroutine para esperar a que el proceso termine por sí mismo
	go func(p *os.Process) {
		waitErr := runCmd.Wait() // Espera a que termine
		processMutex.Lock()
		// Solo limpiar si ESTE es el proceso actual (evita race conditions si se reinició rápido)
		if currentProcess == p {
			currentProcess = nil
			log.Printf("Proceso (PID: %d) terminó. Código de salida: %v", p.Pid, waitErr)
		} else {
			// El proceso ya fue reemplazado por uno nuevo, no hacemos nada
			// log.Printf("Proceso (PID: %d) terminó, pero ya no era el proceso actual.", p.Pid)
		}
		processMutex.Unlock()
	}(runCmd.Process) // Pasar el proceso actual a la goroutine
}

func handleInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig // Espera la señal

	log.Println("Señal de interrupción recibida. Deteniendo...")
	stopApp() // Detener la aplicación hija
	log.Println("gofresh detenido.")
	// La limpieza del binario se hace con defer en main()
}

func cleanup() {
	log.Println("Limpiando binario temporal...")
	err := os.Remove(outputBinary)
	if err != nil && !os.IsNotExist(err) { // No mostrar error si ya no existe
		log.Printf("Error limpiando el binario %s: %v", outputBinary, err)
	}
}
