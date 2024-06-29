package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type EmailConfig struct {
	From     string `json:"from"`
	Password string `json:"password"`
	To       string `json:"to"`
	SmtpHost string `json:"smtpHost"`
	SmtpPort string `json:"smtpPort"`
}

type Config struct {
	SendEmail     bool        `json:"sendEmail"`
	EmailConfig   EmailConfig `json:"emailConfig"`
	CheckInterval int         `json:"checkInterval"`
}

var config Config

// 读取配置文件
func loadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	return nil
}

// 检查设备是否正在使用 (macOS 和 Linux)
func checkDeviceUsageUnix(device string) bool {
	cmd := exec.Command("lsof")
	output, err := cmd.Output()
	if err != nil {
		log.Println("执行命令时出错:", err)
		return false
	}
	return strings.Contains(string(output), device)
}

// 检查设备是否正在使用 (Windows)
func checkDeviceUsageWindows(device string) bool {
	cmd := exec.Command("powershell", "-Command", "Get-Process | Select-String -Pattern "+device)
	output, err := cmd.Output()
	if err != nil {
		log.Println("执行命令时出错:", err)
		return false
	}
	return strings.Contains(string(output), device)
}

// 发送邮件通知
func sendEmail(subject, body string) {
	msg := "From: " + config.EmailConfig.From + "\n" +
		"To: " + config.EmailConfig.To + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", config.EmailConfig.From, config.EmailConfig.Password, config.EmailConfig.SmtpHost)
	err := smtp.SendMail(config.EmailConfig.SmtpHost+":"+config.EmailConfig.SmtpPort, auth, config.EmailConfig.From, []string{config.EmailConfig.To}, []byte(msg))
	if err != nil {
		log.Println("发送电子邮件时出错:", err)
	} else {
		log.Println("电子邮件发送成功")
	}
}

func monitorDevices(wg *sync.WaitGroup) {
	defer wg.Done()
	logFile, err := os.OpenFile("device_usage.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var cameraInUse, micInUse bool

			if runtime.GOOS == "windows" {
				cameraInUse = checkDeviceUsageWindows("camera")
				micInUse = checkDeviceUsageWindows("microphone")
			} else {
				cameraInUse = checkDeviceUsageUnix("AppleCamera")
				micInUse = checkDeviceUsageUnix("CoreAudio")
			}

			if cameraInUse {
				log.Println("摄像头正在使用中！")
				if config.SendEmail {
					sendEmail("检测到未经授权的访问", "摄像头正在使用中！")
				}
			}

			if micInUse {
				log.Println("麦克风正在使用中！")
				if config.SendEmail {
					sendEmail("检测到未经授权的访问", "麦克风正在使用中！")
				}
			}
		}
	}
}

func main() {
	configFile := flag.String("config", "config.json", "配置文件的路径")
	flag.Parse()

	err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件时出错: %v", err)
	}

	fmt.Println("启动设备监视器...")

	var wg sync.WaitGroup
	wg.Add(1)
	go monitorDevices(&wg)
	wg.Wait()
}
