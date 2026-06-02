# 🔒 Masker - CLI утилита для маскирования цифр

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://go.dev/)
[![Tests](https://github.com/qwaseri832/masker/actions/workflows/test.yml/badge.svg)](https://github.com/qwaseri832/masker/actions)
[![Coverage](https://img.shields.io/badge/coverage-72.5%25-yellowgreen)]()

**Masker** — это консольная утилита на Go, которая читает текстовый файл, заменяет все цифры на символ `*` и сохраняет результат в другой файл.

## ✨ Возможности

- 🚀 **Маскировка цифр** — все цифры (0-9) заменяются на `*`
- 📁 **Работа с файлами** — читает из одного файла, пишет в другой
- ⚙️ **Гибкая настройка** — выбор уровней логирования
- 🔄 **Graceful shutdown** — корректное завершение по Ctrl+C
- 🧵 **Worker Pool** — 10 конкурентных воркеров для обработки строк

## 🚀 Быстрый старт

### Установка

```bash
git clone https://github.com/qwaseri832/masker.git
cd masker
go mod download
