## DB_NAME=sbma
## DB_USER=jofosware
## DB_PASS=ejpnAfPJ9BqStu4vvT7mTO3uHCCaGqaG
## CACHE=false
## PRODUCTION=false

## -dbname=${DB_NAME} -dbuser=${DB_USER} -dbpass=${DB_PASS} -cache=${CACHE} -production=${PRODUCTION} &

## build: builds all binaries
build: clean build_front build_back
	@printf "All binaries built!\n"

## clean: cleans all binaries and runs go clean
clean:
	@echo "Cleaning..."
	@- rm -f dist/*
	@go clean
	@echo "Cleaned!"

## build_front: builds the front end
build_front:
	@echo "Building front end..."
	@go build -o dist/sbma ./cmd/web
	@echo "Front end built!"

## build_back: builds the back end
build_back:
	@echo "Building back end..."
	@go build -o dist/sbma_api ./cmd/api
	@echo "Back end built!"

## start: starts front and back end
start: start_front start_back

## db_migration: migrates the various tables
db_migration: 
	@echo "Migrating database once"
	@soda migrate
	@echo "Database migrated!"

## start_front: starts the front end
start_front: build_front
	@echo "Starting the front end..."
	./dist/sbma
	@echo "Front end running!"

## start_back: starts the back end
start_back: build_back
	@echo "Starting the back end..."
	./dist/sbma_api 
	@echo "Back end running!"

## stop: stops the front and back end
stop: stop_front stop_back
	@echo "All applications stopped"

## stop_front: stops the front end
stop_front:
	@echo "Stopping the front end..."
	@-pkill -SIGTERM -f "sbma "
	@echo "Stopped front end"

## stop_back: stops the back end
stop_back:
	@echo "Stopping the back end..."
	@-pkill -SIGTERM -f "sbma_api "
	@echo "Stopped back end"
