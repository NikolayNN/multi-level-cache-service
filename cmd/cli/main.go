package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"aur-cache-service/api/dto"
)

// stringSlice collects repeated flag values.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	addr := flag.String("addr", "http://localhost:8080", "server address")
	cache := flag.String("cache", "", "cache name (required)")
	flag.Parse()
	if *cache == "" || flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	base := strings.TrimRight(*addr, "/")

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	switch cmd {
	case "get":
		cmdGet(client, base, *cache, args)
	case "put":
		cmdPut(client, base, *cache, args)
	case "evict":
		cmdEvict(client, base, *cache, args)
	case "get-all":
		cmdGetAll(client, base, *cache, args)
	case "put-all":
		cmdPutAll(client, base, *cache, args)
	case "evict-all":
		cmdEvictAll(client, base, *cache, args)
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cli -cache <name> [global options] <command> [options]")
	fmt.Fprintln(os.Stderr, "commands: get, put, evict, get-all, put-all, evict-all")
}

func cmdGet(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	key := fs.String("key", "", "key")
	fs.Parse(args)
	if *key == "" {
		fs.Usage()
		os.Exit(1)
	}

	u := fmt.Sprintf("%s/api/cache/%s/%s", base, url.PathEscape(cache), url.PathEscape(*key))
	resp, err := client.Get(u)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
	io.Copy(os.Stdout, resp.Body)
}

func cmdPut(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("put", flag.ExitOnError)
	key := fs.String("key", "", "key")
	value := fs.String("value", "", "json value")
	fs.Parse(args)
	if *key == "" || *value == "" {
		fs.Usage()
		os.Exit(1)
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(*value), &raw); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	u := fmt.Sprintf("%s/api/cache/%s/%s", base, url.PathEscape(cache), url.PathEscape(*key))
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader([]byte(*value)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
}

func cmdEvict(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("evict", flag.ExitOnError)
	key := fs.String("key", "", "key")
	fs.Parse(args)
	if *key == "" {
		fs.Usage()
		os.Exit(1)
	}

	u := fmt.Sprintf("%s/api/cache/%s/%s", base, url.PathEscape(cache), url.PathEscape(*key))
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
}

func cmdGetAll(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("get-all", flag.ExitOnError)
	var keys stringSlice
	fs.Var(&keys, "key", "key (repeatable)")
	fs.Parse(args)
	if len(keys) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	ids := make([]dto.CacheId, len(keys))
	for i, k := range keys {
		ids[i] = dto.CacheId{CacheName: cache, Key: k}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": ids})

	u := fmt.Sprintf("%s/api/cache/batch/get", base)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
	io.Copy(os.Stdout, resp.Body)
}

func cmdPutAll(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("put-all", flag.ExitOnError)
	var entries stringSlice
	fs.Var(&entries, "entry", "key=JSON (repeatable)")
	fs.Parse(args)
	if len(entries) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	reqEntries := make([]dto.CacheEntry, len(entries))
	for i, e := range entries {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "entry must be key=value")
			os.Exit(1)
		}
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(parts[1]), &raw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		reqEntries[i] = dto.CacheEntry{CacheId: &dto.CacheId{CacheName: cache, Key: parts[0]}, Value: &raw}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": reqEntries})

	u := fmt.Sprintf("%s/api/cache/batch/put", base)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
}

func cmdEvictAll(client *http.Client, base, cache string, args []string) {
	fs := flag.NewFlagSet("evict-all", flag.ExitOnError)
	var keys stringSlice
	fs.Var(&keys, "key", "key (repeatable)")
	fs.Parse(args)
	if len(keys) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	ids := make([]dto.CacheId, len(keys))
	for i, k := range keys {
		ids[i] = dto.CacheId{CacheName: cache, Key: k}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": ids})

	u := fmt.Sprintf("%s/api/cache/batch/delete", base)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, resp.Status)
		os.Exit(1)
	}
}
