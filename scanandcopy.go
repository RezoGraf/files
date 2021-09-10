package main

import (
	"fmt"
	"os"
)

func copyFile(oldpath, newpath string) (string, error) {
	oldpath = oldpath + fileName
	newpath = newpath + fileName
	err := os.Rename(oldpath, newpath)
	// if err != nil {
	// 	log.Fatal(err)
	// }/
	if err != nil {
		// fmt.Println("Ебло, где мой ", fileName, " файл в папке ", oldpath)
		return "неудачно!", err
	} else {
		fmt.Println("Заебись, закопировали файл ", fileName)
	}
	return "успешно!", nil
}
