BIN := pak

.PHONY: build run clean

# собрать бинарник через Docker и положить рядом
build:
	docker build -t $(BIN) .
	docker create --name $(BIN)-tmp $(BIN)
	docker cp $(BIN)-tmp:/usr/local/bin/$(BIN) ./$(BIN)
	docker rm $(BIN)-tmp
	@echo "=> ./$(BIN) готов"

# запустить команду внутри контейнера
# пример: make run CMD="search curl"
run:
	docker run --rm $(BIN) $(CMD)

clean:
	rm -f $(BIN)
	docker rmi -f $(BIN) 2>/dev/null || true
