<h1>gRPC-сервис для обмена сообщениями по принципу Publisher-Subscriber с использованием собственной реализации шины событий.</h1>

<h2>Архитектура</h2>
.<br>
│ api/<br>
│   ── pubsub.proto       # Protobuf-схема сервиса<br>
│ cmd/<br>
│   ── client/             # Тестовый клиент(исключительно отладка)<br>
│   ── server/             # Основной сервер<br>
│ internal/<br>
│   ── server/             # gRPC-логика<br>
│ subpub/                  # Реализация PubSub (задание 1)<br>
│ custom_errors/           # ошибки <br>
│ envs/                    # работа с envs <br> 
│ go.mod<br>
│ go.sum<br>
│ .env                   # Пример конфигурации<br>

<h2>Требования</h2>
<ol>
<li>Go 1.24</li>
<li>Утилита protoc</li>
</ol>

<h2>Установка зависимостей</h2>
<pre><code>
    go mod download
</code></pre>

<h2>Генерация gRPC-кода</h2>
<pre><code>
    protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative     api/pubsub.proto
</code></pre>

<h2>Создание в корне .env с наполнением</h2>
<pre><code>
    PORT=:50051
</code></pre>

<h2>Запуск сервера</h2>
<pre><code>
    go run cmd/server/main.go
</code></pre>

<h2>Запуск клиента</h2>
<pre><code>
    go run cmd/client/main.go
</code></pre>

<h2>Тесты с утилитой grpcurl</h2>
<pre><code>
    Подписаться на топик "news"
    grpcurl -plaintext -d '{"key":"news"}' localhost:50051 pubsub.PubSub/Subscribe
    Опубликовать сообщение
    grpcurl -plaintext -d '{"key":"news","data":"Hello"}' localhost:50051 pubsub.PubSub/Publish
</code></pre>

<h2>Сборка</h2>
<pre><code>
    go build pubsub-server cmd/server/main.go
</code></pre>

<h2>Сборка docker</h2>
<pre><code>
    docker build -t pubsub-server .
    docker run --env-file .env pubsub-server
</code></pre>