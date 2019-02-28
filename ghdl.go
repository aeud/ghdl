package ghdl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	Scheme = "github://"
)

type GithubGetContentAPIResponse struct {
	Content string `json:"content"`
}

var (
	memory = make(map[string][]byte)
)

func void() {}

func RefToParams(ref string) (string, string, string, error) {
	if !strings.Contains(ref, Scheme) {
		return "", "", "", fmt.Errorf("%s does not contains %s", ref, Scheme)
	}
	ref = strings.Replace(ref, Scheme, "", 1)
	a := strings.Split(ref, "/")
	if len(a) < 3 {
		return "", "", "", fmt.Errorf("not enough values (only %d found)", len(a))
	}
	return a[0], a[1], strings.Join(a[2:], "/"), nil
}

func ParamsToRef(owner, repo, path string) string {
	return fmt.Sprintf(
		"%s%s",
		Scheme,
		strings.Join([]string{owner, repo, path}, "/"),
	)
}

func DownloadFile(owner, repo, path string) ([]byte, error) {

	ref := ParamsToRef(owner, repo, path)

	if content, ok := memory[ref]; ok {
		log.Printf("ghdl: retrieve %s from memory", ref)
		return content, nil
	}
	log.Printf("ghdl: download %s", ref)
	endpoint := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s",
		owner, repo, path,
	)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_PASS"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	v := &GithubGetContentAPIResponse{}
	if resp.StatusCode != 200 {
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", bs)
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(v.Content)
	if err != nil {
		return nil, err
	}

	memory[ref] = data

	return data, nil
}

func DownloadFileFromRef(ref string) ([]byte, error) {
	owner, repo, path, err := RefToParams(ref)
	if err != nil {
		return nil, err
	}
	return DownloadFile(owner, repo, path)
}

func DownloadFileFromRefForce(ref string) []byte {
	bs, err := DownloadFileFromRef(ref)
	if err != nil {
		panic(err)
	}
	return bs
}
