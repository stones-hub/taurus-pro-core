package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

// ANSI escape sequences define colors
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
)

// DefaultHost and DefaultPort are the default server address and port
var (
	env        = ".env.local"
	configPath = "./config"
	T          *Taurus
	Cleanup    = make([]func(), 0)
	Err        error
)

// init is automatically called before the main function
// --env .env.local --config ./config
func init() {
	// custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n", Cyan+"==================== Usage ===================="+Reset)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s-e, --env <file>%s      Specify the environment file (default \".env.local\")\n", Green, Reset)
		fmt.Fprintf(os.Stderr, "  %s-c, --config <path>%s   Specify the configuration file or directory (default \"config\")\n", Green, Reset)
		fmt.Fprintf(os.Stderr, "  %s-h, --help%s            Show this help message\n", Green, Reset)
		fmt.Fprintf(os.Stderr, "%s\n", Cyan+"==============================================="+Reset)
	}

	// set command line arguments and their aliases
	flag.StringVar(&env, "env", ".env.local", "Environment file")
	flag.StringVar(&env, "e", ".env.local", "Environment file (alias)")
	flag.StringVar(&configPath, "config", "config", "Path to the configuration file or directory")
	flag.StringVar(&configPath, "c", "config", "Path to the configuration file or directory (alias)")

	// parse command line arguments
	flag.Parse()

	// initialize all modules.
	// the env file is not needed, because the makefile has already written the environment variables into the env file, but for the sake of rigor, we still pass the env file to the initialize function
	buildComponents(configPath, env)
}

func StartAndWait(httpServer *server.Server) {
	// use errChan to receive http server startup error
	errChan := make(chan error, 1)
	httpServer.Start(errChan)

	// Block until a signal is received or an error is returned.
	// If an error is returned, it is a fatal error and the program will exit.
	if err := signalWaiter(errChan); err != nil {
		log.Fatalf("%sServer startup failed: %v %s\n", Red, err, Reset)
	}

	// If signalWaiter returns nil, it means the server is running. But received a signal, so we need to shutdown the server.
	// Create a deadline to wait for, 5 seconds or cancel() are all called ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("%sServer forced to shutdown: %v %s\n", Red, err, Reset)
	}

	log.Printf("%s🔗 -> Server shutdown successfully. %s\n", Green, Reset)
	gracefulCleanup(ctx)
}

// signalWaiter waits for a signal or an error, then return
func signalWaiter(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)

	// Block until a signal is received or an error is returned
	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			log.Printf("%s🔗 -> Received signal: %s, graceful shutdown... %s\n", Yellow, sig, Reset)
			// graceful shutdown
			return nil
		}
	case err := <-errCh:
		return err
	}

	return nil
}

// gracefulCleanup is called when the server is shutting down. we can do some cleanup work here.
func gracefulCleanup(ctx context.Context) {

	log.Printf("%s🔗 -> Waiting for all requests to be processed... %s\n", Yellow, Reset)
	done := make(chan struct{})

	go func() {
		for _, cleanup := range Cleanup {
			if cleanup != nil {
				cleanup()
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Printf("%s🔗 -> Server stopped successfully. %s\n", Green, Reset)
	case <-ctx.Done():
		// If 5 seconds have passed and the server has not stopped, it means the server is not responding, so we need to force it to stop.
		log.Printf("%s🔗 -> Server stopped forcefully. %s\n", Red, Reset)
	}
}

// buildComponents builds all components
// configPath is the path to the configuration file or directory
// env is the environment file
func buildComponents(configPath, env string) {
	// build Taurus
	var cleanup func()
	T, cleanup, Err = BuildTaurus()
	if Err != nil {
		log.Fatal(Err)
	}
	Cleanup = append(Cleanup, cleanup)

	// 1. build config
	configComponent := config.New(config.WithPrintEnable(true))
	if err := configComponent.Initialize(configPath, env); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	T.Component.Register("config", configComponent)

	// 2. build http
	httpComponent := server.NewServer(server.WithAddr(configComponent.GetString("http.address")+":"+configComponent.GetString("http.port")),
		server.WithReadTimeout(time.Duration(configComponent.GetInt("http.read_timeout"))*time.Second),
		server.WithWriteTimeout(time.Duration(configComponent.GetInt("http.write_timeout"))*time.Second),
		server.WithIdleTimeout(time.Duration(configComponent.GetInt("http.idle_timeout"))*time.Second))
	T.Component.Register("http", httpComponent)
	Cleanup = append(Cleanup, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configComponent.GetInt("http.shutdown_timeout"))*time.Second)
		defer cancel()
		httpComponent.Shutdown(ctx)
	})
}

// GetConfig returns the config component
func GetConfig() *config.Config {
	configComponent, ok := T.Component.Get("config")
	if !ok {
		log.Fatal("config component not found")
	}
	return configComponent.(*config.Config)
}

// GetHttpServer returns the http server component
func GetHttpServer() *server.Server {
	httpComponent, ok := T.Component.Get("http")
	if !ok {
		log.Fatal("http component not found")
	}
	return httpComponent.(*server.Server)
}
