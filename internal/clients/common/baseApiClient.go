package common

type ApiClient interface {
	Get(key string) (string, bool, error)
	BatchGet(keys []string) (map[string]string, error)
}
