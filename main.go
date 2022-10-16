/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
// @title NATS-LOADER API
// @version 1.0
// @description NATS system loader API
// @contact.name DAPD WSLAID
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name token
// schemes http
package main

import "loader/cmd"

func main() {
	cmd.Execute()
}
