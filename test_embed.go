package main

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:web
var staticFiles embed.FS

func main() {
	// 测试文件是否存在
	testFiles := []string{
		"web/index.html",
		"web/login.html",
		"web/_next/static/chunks/082921070cbd64c6.css",
		"web/_next/static/chunks/8082ab48faca5ea1.js",
	}
	
	for _, file := range testFiles {
		if _, err := fs.Stat(staticFiles, file); err != nil {
			fmt.Printf("❌ %s - NOT FOUND: %v\n", file, err)
		} else {
			fmt.Printf("✅ %s - EXISTS\n", file)
		}
	}
	
	// 尝试创建 sub FS
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		fmt.Printf("❌ Failed to create sub FS: %v\n", err)
		return
	}
	
	fmt.Println("\n--- Testing with sub FS ---")
	testFilesInSub := []string{
		"index.html",
		"_next/static/chunks/082921070cbd64c6.css",
	}
	
	for _, file := range testFilesInSub {
		if _, err := fs.Stat(webFS, file); err != nil {
			fmt.Printf("❌ %s - NOT FOUND: %v\n", file, err)
		} else {
			fmt.Printf("✅ %s - EXISTS\n", file)
		}
	}
}
