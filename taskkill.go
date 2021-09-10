package main

import (
	"fmt"
	"os/exec"
)

func taskKill(process string) (string, error) {
	cmd := exec.Command("cmd.exe", "/C", ("taskkill /im " + process + " /f"))
	_, err := cmd.CombinedOutput()
	// fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		// fmt.Println("Где мой ", process, " сучара?")
		return "процесс " + process + " не найден", err
	} else {
		fmt.Println("Заебись, закрыли ", process)
	}
	return "процесс " + process + " успешно убит", nil
}
