package main

import "tomatoBlogDB/cmd"

// @title
// @version 1.0
// @description This is a sample server Petstore server.

// @host localhost:8888
// @BasePath /

func main() {
	defer cmd.Clean()
	cmd.Start()

}
