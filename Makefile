## build_back: builds the back end
build_back:
	@echo "Building back end..."
	@go build -o dist/shopit_api ./cmd/api
	@echo "Back end built!"

## start: back end
start: start_back start_front

## start_back: starts the back end
start_back: build_back
	@echo "Starting the back end..."
	./dist/shopit_api &
	@echo "Back end running!"

## build_front: builds the front end
build_front:
	@echo "Building front end..."
	@go build -o dist/shopit_web ./cmd/web
	@echo "Front end built!"

## start_front: starts the front end
start_front: build_front
	@echo "Starting the front end..."
	./dist/shopit_web &
	@echo "Front end running!"

## stop: stops the back end and front end
stop: stop_back stop_front
	@echo "All applications stopped"

## stop_back: stops the back end
stop_back:
	@echo "Stopping the back end..."
	@-pkill -SIGTERM -f "shopit_api"
	@echo "Stopped back end"

## stop_front: stops the front end
stop_front:
	@echo "Stopping the front end..."
	@-pkill -SIGTERM -f "shopit_web"
	@echo "Stopped front end"