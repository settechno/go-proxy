## Простой SOCKS5 прокси на golang

### Точки входа:
./cmd/server.go серверная часть

./cmd/create-user.go серверный скрипт добавления пользователя


### Состав сервера:
**server.exe** - само приложение сервера

**create-user.exe** - скрипт добавления пользователей

**config.json** - настройки сервера. Формат:
```json
{
  "port": 1080,
  "use_auth": true,
  "user_file": "users.json"
}
```
**users.json** - список пользователей. Формат:
```json
[
  {
    "username": "alice",
    "password": "alice123"
  },
  {
    "username": "bob",
    "password": "bob123"
  }
]
```

### Добавление пользователей
Для добавления пользователя с ником alice и паролем alice123 вызвать
```cmd
create-user.exe alice alice123
```
Скрипт добавит в users.json информацию о новом пользователе.



### TODO:
- Сделать хранение паролей в виде хешей
- Сделать поддержку HTTP и MTPROTO
- Оптимизировать авторизацию