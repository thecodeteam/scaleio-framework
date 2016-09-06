protoc  --proto_path=$GOPATH/src:. --go_out=. ./mesos/v1/mesos.proto
protoc  --proto_path=$GOPATH/src:. --go_out=. ./mesos/sched/scheduler.proto
