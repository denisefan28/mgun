# M(achine)gun

**Mgun** - это инструмент тестирования нагрузки HTTP сервера.

Mgun создает заданное количество параллельных конкурентных сессий, затем выполняет HTTP запросы и аггрегирует результаты в таблицу.
Параллелизм зависит от количества ядер процессора, чем больше ядер, тем выше параллелизм выполнения HTTP запросов.

Mgun позволяет создавать GET, POST, PUT, DELETE запросы.

Принципиальное отличие mgun от других инструментов тестирования нагрузки в том,
что он позволяет создать сценарий из произвольного количества запросов, имитируя реальное поведение пользователя.
Например, сценарий может иметь вид:

1. Зайти на главную страницу сайта
2. Авторизоваться
3. Зайти в личный кабинет
4. Изменить данные о себе
5. Выйти

Запросы в таком сценарии будут выполняться последовательно, как если бы это делал пользователь.

Для создания сценариев используется конфигурационный файл в формате YAML.

Кроме последовательности запросов в конфигурационном файле можно указать таймаут и заголовки.
Таймаут и заголовки могут быть, как глобальными для всех запросов, так и частными у каждого запроса.

# Быстрый старт

Mgun написан на Go, поэтому для начала необходимо [установить Go](http://golang.org/doc/install).

### Загрузка и установка из исходников

    cd /path/to/gopath
    export GOPATH=/path/to/gopath/
    export GOBIN=/path/to/gopath/bin/
    go get github.com/byorty/mgun
    go install src/github.com/byorty/mgun/mgun.go

### Запуск

    ./bin/mgun -f example/config.yaml

    1000 / 1000 [=============================================================================================================================================================] 100.00 %
    
    Server Hostname:       example.com
    Server Port:           80
    Concurrency Level:     100
    Loop count:            1
    Timeout:               30 seconds
    Time taken for tests:  89 seconds
    Total requests:        1000
    Complete requests:     997
    Failed requests:       3
    Availability:          99.70%
    Requests per second:   ~ 2.99
    Total transferred:     183MB
    
    #    Request                            Compl.  Fail.  Min.     Max.      Avg.      Avail.   Min, avg, max req. per sec.  Content len.  Total trans.
    1.   GET /                              100     0      1.047s.  9.285s.   5.166s.   100.00%  1 / ~ 9.09 / 24              131KB         13MB
    2.   POST /signin                       100     0      0.831s.  8.277s.   4.554s.   100.00%  1 / ~ 5.88 / 16              72B           7.2KB
    3.   GET /basket/                       100     0      0.268s.  22.094s.  11.181s.  100.00%  1 / ~ 1.67 / 4               61KB          15MB
    4.   GET /orders/                       100     0      0.390s.  31.168s.  15.779s.  100.00%  1 / ~ 1.82 / 5               58KB          17MB
    5.   GET /shoes?category=${categories}  98      2      1.089s.  31.546s.  16.318s.  98.00%   1 / ~ 1.72 / 6               476KB         49MB
    6.   GET /shoes?search=${search.shoes}  100     0      1.580s.  17.007s.  9.293s.   100.00%  1 / ~ 1.67 / 5               136KB         16MB
    7.   GET /shoes?id=${shoes.ids}         100     0      0.659s.  14.873s.  7.766s.   100.00%  1 / ~ 1.92 / 6               202KB         20MB
    8.   GET /friends/                      99      1      0.289s.  30.010s.  15.149s.  99.00%   1 / ~ 1.96 / 6               61KB          17MB
    9.   POST /friends/say?id=${frind.ids}  100     0      0.535s.  25.595s.  13.065s.  100.00%  1 / ~ 1.96 / 7               240KB         16MB
    10.  GET /friend?id=${frind.ids}        100     0      2.534s.  17.051s.  9.792s.   100.00%  1 / ~ 2.22 / 6               132KB         19MB



