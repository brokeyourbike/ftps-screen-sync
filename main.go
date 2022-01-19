package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
	"github.com/spf13/viper"
	"golang.design/x/clipboard"

	_ "embed"
)

//go:embed assets/icon.png
var DefaultIcon []byte

//go:embed assets/checkmark.png
var CheckMarkIcon []byte

type Config struct {
	Host       string
	Port       string
	Tls        bool
	Username   string
	Password   string
	SourcePath string
	BaseUrl    string
}

func main() {
	systray.Run(func() {
		systray.SetTemplateIcon(DefaultIcon, DefaultIcon)
		systray.AddSeparator()
		quitButton := systray.AddMenuItem("Quit", "Quit the app")

		go func() {
			for {
				select {
				case <-quitButton.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()

		if err := run(); err != nil {
			log.Fatalf("%s\n", err)
			os.Exit(1)
		}
	}, func() {})
}

func run() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	cfg := Config{
		Host:       viper.GetString("host"),
		Port:       viper.GetString("port"),
		Tls:        viper.GetBool("tls"),
		Username:   viper.GetString("username"),
		Password:   viper.GetString("password"),
		SourcePath: viper.GetString("source_path"),
		BaseUrl:    viper.GetString("base_url"),
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op != fsnotify.Create {
					continue
				}
				if isHidden(path.Base(event.Name)) {
					continue
				}
				if !isPng(path.Base(event.Name)) {
					continue
				}

				fileName := fmt.Sprintf("%s.png", uuid.New().String())
				viewUrl := path.Join(cfg.BaseUrl, fileName)
				clipboard.Write(clipboard.FmtText, []byte(viewUrl))

				go func() {
					systray.SetTemplateIcon(CheckMarkIcon, CheckMarkIcon)
					time.Sleep(time.Second)
					systray.SetTemplateIcon(DefaultIcon, DefaultIcon)
				}()

				go upload(cfg, event.Name, fileName)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(cfg.SourcePath)
	if err != nil {
		return err
	}
	<-done

	return nil
}

func upload(cfg Config, path string, fileName string) error {
	ftpOptions := []ftp.DialOption{
		ftp.DialWithTimeout(5 * time.Second),
	}

	if cfg.Tls {
		ftpOptions = append(ftpOptions, ftp.DialWithExplicitTLS(&tls.Config{
			InsecureSkipVerify: true,
		}))
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

	err = c.Stor(fileName, file)
	if err != nil {
		return err
	}

	if err := c.Quit(); err != nil {
		return err
	}

	return nil
}

func isHidden(path string) bool {
	return path[0] == 46
}

func isPng(path string) bool {
	return filepath.Ext(path) == ".png"
}
