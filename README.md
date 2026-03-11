# Урок 1 - Команды и аргументы

Одним из способов передачи входных параметров в нашу программу являются аргументы командной строки.</br>
Для получения доступа к аргументам, с которыми была запущена программа, сущесвтует пакет <b>os</b>, который </br>
содержит в себе string слайс Args[]. Именно этот слайс отвечает за хранение аргументов.</br>
```go
package main
  
import (
    "fmt"
    "os"
)

// Пример запуска программы с передачей аргументов:
// go run main.go first second
func main() {
    // Первым аргументом всегда идет полный путь до исполняемого файла программы
	// В нашем случае - .\main.go
    programName := os.Args[0] // = .\main.go
    firstArg := os.Args[1] // = first
    secondArg := os.Args[2] // = second
	
    fmt.Printf("absolute path of runnable file is %s\n", programName)
    fmt.Printf("first arg is %s\n", firstArg)
    fmt.Printf("second arg is %s\n", secondArg)
}
```

## Флаги

В Golang для более <b>удобной</b> работы с аргументами программы существует пакет <b>flag</b>. </br>
Он дает возможность именовать агрументы (флаги) программы и приводить переданные значения к </br>
конкретным типам, а также устанавливать флагам значение по умолчанию, если есть необходимость. </br>

```go
package main

import (
    "flag"
    "fmt"
)

// Пример запуска программы с передачей аргументов:
// go run main.go -t 2000 -ttr true
func main() {
    timeoutMs := flag.Int("t", 1000, "connection timeout")
    tryToReconnect := flag.Bool("ttr", false, "should try to reconnect?")
    flag.Parse() 
    fmt.Printf("t flag is %d\n", *timeoutMs)
    fmt.Printf("ttr flag is %v\n", *tryToReconnect)
}
```
Для управления работой программных продуктов нередко создают CLI (Command Line Interface), </br>
состоящие из "деревьев" комбинаций команд и флагов. Они позволяют запускать тот или иной функционал </br>
приложения без использования графического интерфейса. Ярким примером такого взаимодействия является </br>
<b>docker cli</b> (между прочим, написанный на Golang!), позволяющий нам запустить экземпляр Базы Данных </br>
в контейнере, написав всего одну командную строку.</br> Например:</br>
<b><em>docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres</em></b> </br>

## Конфигурационные файлы

Для передачи параметров в приложение также используются <b>конфигурационные файлы</b> и <b>переменные окружения</b>. </br>
Как правило, они служат для аргументов программы, которые меняются редко, или не должны меняться вовсе. </br>
К таким аргументам в основном относятся:</br>
- Параметры для URL подключения к Базе Данных;</br>
- Параметры для соединения с каким-либо сервером;</br>
- Параметры для различных сервисов транспорта. К примеру, Apache Kafka, RabbitMQ;</br>
- и т.д.</br>

Пример того, как выглядит файл конфигурации config.yml, хранящий параметры для соединения </br>
с grpc-сервером и для подключения к БД PostgreSQL:

```yml
grpc:
  host: "localhost"
  port: 9000

psql:
  host: "localhost"
  port: "5432"
  user: "postgres"
  pass: "postgres"
  dbname: "dbname"
  sslmode: "disable"
```

Чтобы применить параметры, хранящиеся в файле конфигурации, их как-то надо достать из файла. </br>
Для этого необходимо создать файл config.go, где описан метод Init(), который и будет "дёргать" параметры из конфига.</br>
На помощь к нам могут прийти библиотеки на подобии <b>viper</b>, но на этот раз мы обойдемся без сторонних библиотек.</br> 
Единственное что, придется подключить библиотеку для работы с .yml файлами. </br>
Подключим библиоткеу, выполнив сначала команду:</br>
```text
go mod init task1
```
Команда go mod init используется для инициализации нового модуля Go. Модули в Go позволяют управлять зависимостями в проекте</br>
и гарантировать совместимость версий зависимых библиотек.</br>

При вызове go mod init необходимо передать название модуля в качестве аргумента. Это может быть имя GitHub-репозитория,</br>
URL или произвольная строка. После этого команда создаст файл go.mod, который содержит информацию о модуле, </br>
включая его название, версию и зависимости.</br>
теперь скачаем и пропишем в go.mod зависимость на необходимую библиотеку:
```text
go get -u github.com/go-yaml/yaml
```

Пример написания config.go для конфига, описанного выше:
```go
package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	"os"
)

type Postgres struct {
	User     string `yaml:"user"`
	Port     string `yaml:"port"`
	Password string `yaml:"pass"`
	Host     string `yaml:"host"`
	Dbname   string `yaml:"dbname"`
	Sslmode  string `yaml:"sslmode"`
}

type ServerGRPC struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Postgres   Postgres   `yaml:"psql"`
	ServerGRPC ServerGRPC `yaml:"grpc"`
}

func Init() (*Config, error) {
	filePath := flag.String("c", "config.yml", "Path to configuration file")
	flag.Parse()
	config := &Config{}
	data, err := os.ReadFile(*filePath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}
```

После того, как мы расписали config.go, нам нужно будет проинициализировать конфиг.
Пример инициализации конфига в main:

```go
package main

import (
	"fmt"
	"gopkg.in/yaml.v3/config"
	"log"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("error initializating config")
	}
	fmt.Printf("Server config:\n host: %s\n port: %d", cfg.ServerGRPC.Host, cfg.ServerGRPC.Port)
}
```

### Задание

На вход в программу подается N - порядковый номер числа Фибоначчи. </br>
Порядковый номер может передается как с помощью флагов, так и файла конфигурации. Реализовать выбор между способами ввода.</br>
На выход в консоль программа должна вывести число Фибоначчи, соответствующие переданному порядковому номеру.

Создать unit тест для проверки корректности работы функционала