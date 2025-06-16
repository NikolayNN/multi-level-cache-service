package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Внешний API содержит минимальные данные для того чтобы найти значение  для GET
type CacheId struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
}

// Внешний API содержит минимальные данные для идентификации и значение используется для PUT
type CacheEntry struct {
	CacheId *CacheId         `json:",inline"`
	Value   *json.RawMessage `json:"v"`
}

// Внешний API содержит минимальные данные для идентифиации, значение и др информацию это результат GET
type CacheEntryHit struct {
	CacheEntry *CacheEntry `json:",inline"`
	Found      bool        `json:"f"`
}

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
	fmt.Fprintln(os.Stderr, "commands: get-all, put-all, evict-all")
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

	ids := make([]CacheId, len(keys))
	for i, k := range keys {
		ids[i] = CacheId{CacheName: cache, Key: k}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": ids})

	u := fmt.Sprintf("%s/api/v1/cache/get_all", base)
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

	reqEntries := make([]CacheEntry, len(entries))
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
		reqEntries[i] = CacheEntry{CacheId: &CacheId{CacheName: cache, Key: parts[0]}, Value: &raw}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": reqEntries})

	u := fmt.Sprintf("%s/api/v1/cache/put_all", base)
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

	ids := make([]CacheId, len(keys))
	for i, k := range keys {
		ids[i] = CacheId{CacheName: cache, Key: k}
	}
	body, _ := json.Marshal(map[string]interface{}{"requests": ids})

	u := fmt.Sprintf("%s/api/v1/cache/evict_all", base)
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
