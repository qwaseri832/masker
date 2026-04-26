package main

import (
    "fmt"
    "masker/masker"
    "os"
)

func main() {
    // Проверка аргументов
    if len(os.Args) < 2 {
        fmt.Println("Использование: go run . <входной_файл> [выходной_файл]")
        return
    }

    inputPath := os.Args[1]
    outputPath := "output.txt"
    if len(os.Args) >= 3 {
        outputPath = os.Args[2]
    }

    // Создание компонентов
    prod := masker.NewFileProducer(inputPath)
    pres := masker.NewFilePresenter(outputPath)

    // Создание и запуск сервиса
    svc := masker.NewService(prod, pres)
    if err := svc.Run(); err != nil {
        fmt.Fprintln(os.Stderr, "Ошибка:", err)
        os.Exit(1)
    }

    fmt.Println("Готово! Результат в", outputPath)
}