## CLI клиент
В каталоге `cmd/cli` находится небольшая утилита для работы с кэшем
через HTTP API сервиса. По умолчанию используется адрес
`http://localhost:8080`, изменить его можно флагом `-addr`. Утилита
требует указания имени кэша через `-cache`.

Сборка выполняется стандартной командой:

```bash
go build ./cmd/cli
```

Примеры использования:

```bash
# сохранить значение
./cli -cache user put -key 42 -value '{"name":"Bob"}'

# получить значение
./cli -cache user get -key 42

# удалить значение
./cli -cache user evict -key 42

# получить несколько значений
./cli -cache user get-all -key 1 -key 2

# сохранить несколько значений
./cli -cache user put-all -entry '1={"name":"Ann"}' -entry '2={"name":"Bob"}'

# удалить несколько
./cli -cache user evict-all -key 1 -key 2
```
