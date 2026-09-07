
#!/usr/bin/env python3

import os
import subprocess
import sys
from pathlib import Path

# ---------------------- Содержимое файлов ----------------------

GO_MOD = """module clinic

go 1.21
"""

MAIN_GO = '''package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type PageData struct {
	Title   string
	Content string
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			filepath.Join("templates", "base.html"),
			filepath.Join("templates", "blocks", "header.html"),
			filepath.Join("templates", "blocks", "content.html"),
			filepath.Join("templates", "blocks", "footer.html"),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := PageData{
			Title:   "Главная",
			Content: "Добро пожаловать в клинику!",
		}
		err = tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
'''

GITIGNORE = """# Бинарные файлы
*.exe
*.dll
*.so
*.dylib
clinic

# Зависимости
vendor/

# Логи и временные файлы
*.log
*.tmp

# Конфигурации IDE
.idea/
.vscode/
*.swp
*.swo

# Системные файлы
.DS_Store
Thumbs.db
"""

TEMPLATES = {
    "templates/base.html": '''<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
</head>
<body>
    {{template "header" .}}
    {{template "content" .}}
    {{template "footer" .}}
</body>
</html>''',
    "templates/blocks/header.html": '''{{define "header"}}
<header>
    <h1>Клиника</h1>
    <nav>Меню: Главная | Услуги | Контакты</nav>
</header>
{{end}}''',
    "templates/blocks/content.html": '''{{define "content"}}
<main>
    <h2>{{.Title}}</h2>
    <p>{{.Content}}</p>
</main>
{{end}}''',
    "templates/blocks/footer.html": '''{{define "footer"}}
<footer>
    <p>&copy; 2026 Клиника. Все права защищены.</p>
</footer>
{{end}}''',
}

# --------------------------------------------------------------

def run_cmd(cmd, cwd, check=True):
    """Выполняет команду, возвращает stdout. При ошибке завершает программу."""
    print(f"Выполняется: {cmd}")
    result = subprocess.run(cmd, shell=True, cwd=cwd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Ошибка при выполнении команды:\n{cmd}\n{result.stderr}")
        if check:
            sys.exit(1)
        return None
    if result.stdout:
        print(result.stdout.strip())
    return result.stdout

def create_file(path, content, force=False):
    """Создаёт файл с содержимым, если его нет или если force=True."""
    if path.exists():
        if not force:
            print(f"Файл {path} уже существует. Пропускаем (используйте --force для перезаписи).")
            return
        else:
            print(f"Перезаписываем {path}")
    else:
        print(f"Создаём {path}")
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content)

def main():
    import argparse
    parser = argparse.ArgumentParser(description="Автоматическая настройка бэкенда на Go и первый коммит")
    parser.add_argument("--dir", default=".", help="Рабочая директория (по умолчанию текущая)")
    parser.add_argument("--force", action="store_true", help="Перезаписывать существующие файлы")
    parser.add_argument("--no-build", action="store_true", help="Пропустить компиляцию .exe")
    parser.add_argument("--no-push", action="store_true", help="Пропустить git push (только коммит)")
    args = parser.parse_args()

    base_dir = Path(args.dir).resolve()
    if not base_dir.exists():
        print(f"Директория {base_dir} не существует.")
        sys.exit(1)

    print(f"Работаем в {base_dir}")

    # 1. Создание файлов проекта
    print("\n--- Создание файлов ---")
    create_file(base_dir / "go.mod", GO_MOD, args.force)
    create_file(base_dir / "main.go", MAIN_GO, args.force)
    create_file(base_dir / ".gitignore", GITIGNORE, args.force)
    for rel_path, content in TEMPLATES.items():
        create_file(base_dir / rel_path, content, args.force)

    # 2. Инициализация Go модуля (если ещё нет)
    print("\n--- Инициализация Go модуля ---")
    mod_path = base_dir / "go.mod"
    if mod_path.exists():
        # Проверяем, что модуль уже инициализирован
        with open(mod_path, 'r') as f:
            if "module clinic" in f.read():
                print("Модуль уже инициализирован.")
            else:
                run_cmd("go mod init clinic", cwd=base_dir)
    else:
        run_cmd("go mod init clinic", cwd=base_dir)

    # 3. Компиляция .exe (опционально)
    if not args.no_build:
        print("\n--- Компиляция приложения ---")
        run_cmd("go build -o clinic.exe main.go", cwd=base_dir)

    # 4. Git инициализация и коммит
    print("\n--- Инициализация Git ---")
    git_dir = base_dir / ".git"
    if not git_dir.exists():
        run_cmd("git init", cwd=base_dir)
    else:
        print("Git репозиторий уже инициализирован.")

    print("\n--- Добавление файлов в Git ---")
    run_cmd("git add .", cwd=base_dir)

    print("\n--- Создание коммита ---")
    # Проверяем, есть ли изменения для коммита
    status = subprocess.run("git status --porcelain", shell=True, cwd=base_dir, capture_output=True, text=True)
    if status.stdout.strip():
        run_cmd('git commit -m "Initial commit: backend on Go with template blocks"', cwd=base_dir)
    else:
        print("Нет изменений для коммита.")

    # 5. Добавление удалённого репозитория
    print("\n--- Настройка удалённого репозитория ---")
    # Проверяем, есть ли уже remote origin
    remotes = subprocess.run("git remote -v", shell=True, cwd=base_dir, capture_output=True, text=True).stdout
    if "origin" in remotes:
        print("Remote 'origin' уже существует, пропускаем добавление.")
    else:
        run_cmd("git remote add origin git@github.com:goshva/clinic.git", cwd=base_dir)

    # Устанавливаем ветку main
    run_cmd("git branch -M main", cwd=base_dir)

    # 6. Пуш (если не запрещено)
    if not args.no_push:
        print("\n--- Отправка в удалённый репозиторий ---")
        run_cmd("git push -u origin main", cwd=base_dir)
    else:
        print("\nПуш пропущен (--no-push).")

    print("\n✅ Готово! Бэкенд создан, первый коммит выполнен.")
    print(f"Запустите сервер: {base_dir / 'clinic.exe'} (или go run main.go)")

if __name__ == "__main__":
    main()
