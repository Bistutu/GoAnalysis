package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func checkGraphviz() {
	_, err := exec.LookPath("dot")
	if err != nil {
		fmt.Println("正在尝试安装 Graphviz，请耐心等待...")
		cmd := exec.Command("brew", "install", "graphviz")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			log.Fatalf("安装 Graphviz 失败: %v", err)
		}
	}
}

func main() {
	checkGraphviz()

	rand.Seed(time.Now().UnixNano())

	var duration string // duration 表示要分析的秒数

	switch len(os.Args) {
	case 1:
		fmt.Println("请提供目标端口号")
		return
	case 2:
		duration = "3"
	case 3:
		duration = os.Args[2]
	}

	targetPort := os.Args[1]
	randomPort := strconv.Itoa(10000 + rand.Intn(50000))

	customProfileFile := "pprof.samples.cpu.pb.gz" // 自定义文件名

	profileURL := fmt.Sprintf("http://127.0.0.1:%s/debug/pprof/profile?seconds=%s", targetPort, duration)
	resp, err := http.Get(profileURL)
	if err != nil {
		log.Fatalf("获取 pprof 数据失败: %v", err)
	}
	defer resp.Body.Close()

	pprofData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取 pprof 数据失败: %v", err)
	}

	err = os.WriteFile(customProfileFile, pprofData, 0644)
	if err != nil {
		log.Fatalf("保存 pprof 数据失败: %v", err)
	}

	httpPort := ":" + randomPort
	fmt.Printf("在新的随机端口 %s 上启动 HTTP 服务器\n", randomPort)

	go func() {
		cmd := exec.Command("go", "tool", "pprof", "-http="+httpPort, customProfileFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("启动 HTTP 服务器失败: %v", err)
		}
		time.Sleep(1000 * time.Millisecond)
		err = os.Remove(customProfileFile)
		if err != nil {
			fmt.Println(err)
		}
	}()

	// 删除生成的文件

	fmt.Printf("请访问 http://127.0.0.1:%s/ui/flamegraph2 查看火焰图\n", randomPort)

	select {}
}
