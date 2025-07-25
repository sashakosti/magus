# Magus: Твой Личный RPG-Помощник в Терминале!

Привет, искатель приключений! Magus — это не просто утилита, это твой верный спутник в мире квестов, прокачки и всяких RPG-штучек, прямо в твоем терминале! Написанный на Go, он поможет тебе управлять своими задачами, как будто это настоящие приключения, а ты — их главный герой!

## Что умеет Magus? (Спойлер: много всего крутого!)

*   **Управление Квестами**: Добавляй новые задания, смотри список активных, отмечай выполненные и погружайся в детали каждого квеста. Ни один дракон не скроется!
*   **Прокачка Персонажа**: Отслеживай свой опыт и уровни. Чувствуй, как ты становишься сильнее с каждым выполненным квестом!
*   **Система Перков**: Открывай и применяй крутые перки, которые сделают твоего персонажа (или тебя!) еще более уникальным.
*   **Интерактивный TUI**: Погрузись в атмосферу приключений с нашим интерактивным терминальным интерфейсом. Это почти как игра, только лучше!

## Как начать свое приключение с Magus?

### Что тебе понадобится?

*   Go (версия 1.18 или новее — наш любимый эликсир!)

### Установка (проще, чем победить гоблина!)

1.  Клонируй наше сокровище (репозиторий):
    ```bash
    git clone https://github.com/your-username/magus.git
    cd magus
    ```
2.  Собери своего Магуса (это как создать артефакт!):
    ```bash
    go build -o magus .
    ```

### Использование (твои первые заклинания!)

Запусти `magus` из терминала и пусть магия начнется!

```bash
./magus
```

Или используй конкретные команды, чтобы творить чудеса:

*   `./magus add <описание_квеста>`: Добавить новый квест. Вперед, к приключениям!
*   `./magus list`: Показать все активные квесты. Что у нас сегодня по плану?
*   `./magus complete <id_квеста>`: Отметить квест как выполненный. Поздравляем, герой!
*   `./magus show <id_квеста>`: Показать детали конкретного квеста. Вспомни, что тебя ждет!
*   `./magus roadmap <id_квеста>`: Показать роадмап для цели и всех её подзадач.
*   `./magus why`: (Возможно, чтобы понять, почему ты такой крутой или почему этот квест так важен!)

Загляни в папку `cmd/` для более подробной информации о командах. Там спрятаны все секреты!

## Структура Проекта (наша карта сокровищ)

*   `cmd/`: Здесь живут все команды Cobra CLI. Это как твоя книга заклинаний.
*   `data/`: Тут хранятся все твои сокровища: JSON-данные для перков, игрока и квестов.
*   `player/`: Логика, связанная с игроком, включая опыт и типы. Твой персонаж здесь оживает!
*   `quests/`: Данные и логика, связанные с квестами. Сердце всех приключений.
*   `rpg/`: Основные механики RPG, такие как уровни и перки. Здесь происходит вся магия!
*   `storage/`: Отвечает за сохранение твоих приключений. Ничего не потеряется!
*   `tui/`: Компоненты Терминального Пользовательского Интерфейса. Твой портал в мир Magus.
*   `utils/`: Всякие полезные функции. Наши маленькие помощники.

## Вклад в Проект (присоединяйся к гильдии!)

Мы всегда рады новым героям! Не стесняйся открывать issues или отправлять pull requests. Вместе мы сделаем Magus еще круче!

## Лицензия

[Укажи свою лицензию здесь, например, MIT License. Пусть все знают правила игры!]