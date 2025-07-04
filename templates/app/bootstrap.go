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
	cleanup    func()
	err        error
)

func Run() {
	// use errChan to receive http server startup error
	errChan := make(chan error, 1)
	T.Http.Start(errChan)

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
	if err := T.Http.Shutdown(ctx); err != nil {
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
		cleanup()
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

// buildComponents builds all components
// configPath is the path to the configuration file or directory
// env is the environment file
func buildComponents(configPath, env string) {
	// build Taurus
	T, cleanup, err = buildTaurus(&ConfigOptions{
		ConfigPath:  configPath,
		Env:         env,
		PrintEnable: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
