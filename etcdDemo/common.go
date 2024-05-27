package etcddemo

var (
	ETCD_CLUSTER = []string{"127.0.0.1:2379"}
)

const (
	SERVICE_ROOT_PATH = "/service/grpc" // etcd key 前缀
	HELLO_SERVICE = "hello_service"
)