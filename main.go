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
				<-quitButton.ClickedCh
				systray.Quit()
			}
		}()

		done, err := run()
		if err != nil {
			log.Fatalf("%s\n", err)
			quitButton.ClickedCh <- struct{}{}
		}

		<-done
		quitButton.ClickedCh <- struct{}{}
	}, func() {})
}

func run() (<-chan bool, error) {
	cfg := newConfig()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	defer watcher.Close()

	done := make(chan bool)
	go watch(cfg, watcher)

	err = watcher.Add(cfg.SourcePath)
	if err != nil {
		return done, err
	}

	return done, nil
}

func watch(cfg Config, watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
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

			go upload(cfg, event.Name)

			go func() {
				systray.SetTemplateIcon(CheckMarkIcon, CheckMarkIcon)
				time.Sleep(time.Second)
				systray.SetTemplateIcon(DefaultIcon, DefaultIcon)
			}()
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Println("error:", err)
		}
	}
}

func upload(cfg Config, sourcePath string) error {
	fileName := fmt.Sprintf("%s.png", uuid.New().String())

	go func() {
		viewUrl := path.Join(cfg.BaseUrl, fileName)
		clipboard.Write(clipboard.FmtText, []byte(viewUrl))
	}()

	conn, err := newConn(cfg)
	if err != nil {
		return err
	}

	defer conn.Quit()

	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %v", sourcePath, err)
	}

	return conn.Stor(fileName, file)
}

func newConfig() Config {
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

	return cfg
}

func newConn(cfg Config) (*ftp.ServerConn, error) {
	ftpOptions := []ftp.DialOption{
		ftp.DialWithTimeout(5 * time.Second),
	}

	if cfg.Tls {
		ftpOptions = append(ftpOptions, ftp.DialWithExplicitTLS(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	c, err := ftp.Dial(addr, ftpOptions...)
	if err != nil {
		return nil, fmt.Errorf("cannot dial FTP server %s: %v", addr, err)
	}

	err = c.Login(cfg.Username, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to FTP: %v", err)
	}

	return c, nil
}

func isHidden(path string) bool {
	return path[0] == 46
}

func isPng(path string) bool {
	return filepath.Ext(path) == ".png"
}
