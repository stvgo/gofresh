package main

import (
	"flag"
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
	debounceTime   = 500 * time.Millisecond
	outputBinary   = ".gofresh-app"
	cwd            string
	initFlag       = flag.Bool("init", false, "Inicializa gofresh en el directorio actual (futuro).")
)

func main() {
	flag.Parse()

	if *initFlag {
		fmt.Println("Modo init detectado. Funcionalidad futura: crearía archivo de configuración.")
		os.Exit(0)
	}

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatalf("Error obteniendo directorio de trabajo: %v", err)
	}
	log.Printf("🚀 Iniciando gofresh en: %s", cwd)

	defer cleanup()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creando el watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(cwd)
	if err != nil {
		log.Fatalf("Error añadiendo directorio '%s' al watcher: %v", cwd, err)
	}
	log.Printf("Observando cambios en: %s", cwd)

	buildAndRun()

	go watchEvents(watcher)

	handleInterrupt()
}

func watchEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if shouldRestart(event) {
				log.Printf("Cambio detectado en: %s", event.Name)
				scheduleRestart()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error del watcher: %v", err)
		}
	}
}

func shouldRestart(event fsnotify.Event) bool {
	if !strings.HasSuffix(event.Name, ".go") {
		return false
	}
	if event.Name == filepath.Join(cwd, outputBinary) {
		return false
	}
	return event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
}

func scheduleRestart() {
	processMutex.Lock()
	if restartTimer != nil {
		restartTimer.Reset(debounceTime)
		log.Println("Debounce: reinicio reprogramado.")
	} else {
		log.Println("Debounce: programando reinicio...")
		restartTimer = time.AfterFunc(debounceTime, func() {
			log.Println("🔄 ¡Tiempo de debounce cumplido! Reiniciando aplicación...")
			stopApp()
			buildAndRun()
			processMutex.Lock()
			restartTimer = nil
			processMutex.Unlock()
		})
	}
	processMutex.Unlock()
}

func stopApp() {
	processMutex.Lock()
	defer processMutex.Unlock()

	if currentProcess == nil {
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

func killProcess(p *os.Process) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", p.Pid))
		return cmd.Run()
	}
	return p.Kill()
}

func buildAndRun() {
	processMutex.Lock()
	defer processMutex.Unlock()

	outputBinaryPath := filepath.Join(cwd, outputBinary)
	log.Printf("Compilando aplicación en %s...", cwd)
	buildCmd := exec.Command("go", "build", "-o", outputBinaryPath, ".")
	buildCmd.Dir = cwd
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		log.Printf("Error en la compilación: %v", err)
		return
	}
	log.Println("Compilación exitosa.")

	log.Printf("Ejecutando: %s", outputBinaryPath)
	runCmd := exec.Command(outputBinaryPath)
	runCmd.Dir = cwd
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	configureCmdSysProcAttr(runCmd)

	err = runCmd.Start()
	if err != nil {
		log.Printf("Error al iniciar la aplicación %s: %v", outputBinaryPath, err)
		return
	}

	currentProcess = runCmd.Process
	log.Printf("Aplicación iniciada con PID: %d en %s", currentProcess.Pid, cwd)

	go func(p *os.Process, binaryPath string) {
		waitErr := runCmd.Wait()
		processMutex.Lock()
		if currentProcess == p {
			currentProcess = nil
			log.Printf("Proceso (PID: %d, Binario: %s) terminó. Código de salida: %v", p.Pid, binaryPath, waitErr)
		} else {
		}
		processMutex.Unlock()
	}(runCmd.Process, outputBinaryPath)
}

func handleInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Señal de interrupción recibida. Deteniendo...")
	stopApp()
	log.Println("👋 gofresh detenido.")
}

func cleanup() {
	outputBinaryPath := filepath.Join(cwd, outputBinary)
	log.Printf("Limpiando binario temporal: %s...", outputBinaryPath)
	err := os.Remove(outputBinaryPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error limpiando el binario %s: %v", outputBinaryPath, err)
	}
}
