package main

import (
	"fmt"
	"log"
	"net/http"
)

// 简单的测试服务器,用于验证网络抓包功能
func startTestServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>测试页面</title>
    <style>
        body { font-family: Arial; padding: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; }
        button { padding: 12px 24px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; font-size: 14px; }
        .btn-primary { background: #007bff; color: white; }
        #result { margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🧪 网络请求测试页面</h1>
        <p>这个页面运行在 http://localhost:9999</p>
        <button class="btn-primary" onclick="testFetch()">测试 Fetch</button>
        <button class="btn-primary" onclick="testXHR()">测试 XHR</button>
        <div id="result"></div>
    </div>
    <script>
        function log(msg) {
            document.getElementById('result').innerHTML += '<div>' + msg + '</div>';
        }
        
        async function testFetch() {
            log('发送 Fetch 请求...');
            try {
                const resp = await fetch('/api/test');
                const data = await resp.json();
                log('✅ Fetch 成功: ' + JSON.stringify(data));
            } catch (e) {
                log('❌ Fetch 失败: ' + e.message);
            }
        }
        
        function testXHR() {
            log('发送 XHR 请求...');
            const xhr = new XMLHttpRequest();
            xhr.open('GET', '/api/data');
            xhr.onload = function() {
                log('✅ XHR 成功: ' + xhr.responseText);
            };
            xhr.onerror = function() {
                log('❌ XHR 失败');
            };
            xhr.send();
        }
        
        log('页面加载完成,可以开始测试');
    </script>
</body>
</html>
`)
	})

	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `{"message":"Hello from API","timestamp":%d}`, r.Context().Value("time"))
	})

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `{"data":"Test data","method":"%s"}`, r.Method)
	})

	log.Println("测试服务器启动在 http://localhost:9999")
	log.Println("在网络抓包工具中输入: http://localhost:9999")

	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal(err)
	}
}

// 在 main 函数中可以选择性启动
// go startTestServer()
