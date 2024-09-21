### Simple and Efficient Kafka Manager

### Features

- Multi Cluster Support
- Get Consumer Group List
- Get Consumer Group Info
- Get Topic List
- Get Topic Info
- Get Top Messages

### Tech Stack

- Go 1.21
- IBM/sarama v1.43.3
- Swag for api docs

### Usage
```shell
> export CONFIG_FILE_PATH=local/config.yml
> go run main.go
> open localhost:8087/swagger 
```

### RoadMap

- Top Messages with Pagination
- Change Offset of Given Consumer Group




