package main

import "github.com/DiasOrazbaev/vkcheck/parse/vk"

func main() {
	vk.Generate()
	vk.SendLog("someusername", "somepassword")
}
