/*
Copyright © 2026 rtsp-recorder contributors

A CLI tool for recording RTSP video streams to MP4 files.
*/
package main

import (
	"rtsp-recorder/cmd"
	_ "rtsp-recorder/logger"
)

func main() {
	cmd.Execute()
}
