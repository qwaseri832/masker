package main

import (
    "fmt"
    "masker/masker"
    "os"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

func run() error {
    inputPath, outputPath := parseArgs()

    prod := masker.NewFileProducer(inputPath)
    pres := masker.NewFilePresenter(outputPath)
    maskerImpl := masker.DigitsMasker{} // стратегия маскирования

    svc := masker.NewService(prod, pres, maskerImpl)
    return svc.Run()
}

func parseArgs() (string, string) {
    input := "input.txt"
    output := "output.txt"
    if len(os.Args) >= 2 {
        input = os.Args[1]
    }
    if len(os.Args) >= 3 {
        output = os.Args[2]
    }
    return input, output
}
