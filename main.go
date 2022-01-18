package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/jlaffaye/ftp"
	"github.com/radovskyb/watcher"
	"golang.design/x/clipboard"
)

type Config struct {
	Host       string `env:"HOST" envDefault:""`
	Port       string `env:"PORT" envDefault:"21"`
	Username   string `env:"USERNAME" envDefault:""`
	Password   string `env:"PASSWORD" envDefault:""`
	SourcePath string `env:"SOURCE_PATH" envDefault:""`
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("cannot parse config: %v", err)
	}

	// Instantiate watcher and only watch for CREATE events
	w := watcher.New()
	w.FilterOps(watcher.Create)

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
				go upload(cfg, event.Path)
				clipboard.Write(clipboard.FmtText, []byte(event.Path))
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch test_folder recursively for changes
	if err := w.AddRecursive(cfg.SourcePath); err != nil {
		log.Fatalln(err)
	}

	// Start the watching process
	log.Printf("Watching: %v\n", cfg.SourcePath)
	if err := w.Start(time.Duration(time.Millisecond * 100)); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func upload(cfg Config, path string) error {
	ftpOptions := []ftp.DialOption{
		ftp.DialWithTimeout(5 * time.Second),
		ftp.DialWithExplicitTLS(&tls.Config{}),
	}

	c, err := ftp.Dial(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), ftpOptions...)
	if err != nil {
		return fmt.Errorf("cannot dial FTP: %v", err)
	}

	err = c.Login(cfg.Username, cfg.Password)
	if err != nil {
		return fmt.Errorf("cannot connect to FTP: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %v", path, err)
	}

	err = c.Stor(filepath.Base(path), file)
	if err != nil {
		return err
	}

	if err := c.Quit(); err != nil {
		return err
	}

	return nil
}
