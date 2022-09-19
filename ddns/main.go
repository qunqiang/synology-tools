package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	detectRemoteIpService = []string{
		//"http://icanhazip.com/",
		"http://ip.42.pl/raw",
		"http://ident.me/",
		"http://ip.3322.net/",
		"https://api.ip.sb/ip",
	}
)

const (
	MaxRetries = 3
)

func main() {
	fmt.Println(os.Getenv("ENV"))
	fmt.Println("It works")
	ch := time.Tick(time.Minute)
	for range ch {

		var nowIp string
		for _, host := range detectRemoteIpService {
			if content, err := request(http.MethodGet, host, nil, ""); err != nil {
				fmt.Println(fmt.Errorf("%e", err))
				continue
			} else {
				fmt.Println(content)
				nowIp = content
				break
			}
		}

		// 设置 ddns
		for i := 0; i < MaxRetries; i++ {
			if err := setDDNSRecord(nowIp); err != nil {
				fmt.Println(fmt.Errorf("%e", err))
				continue
			}
			break
		}
	}

}

var (
	domainName = os.Getenv("DOMAIN")
	name       = os.Getenv("NAME")
	headers    = map[string]string{
		"Authorization": "sso-key " + os.Getenv("AK") + ":" + os.Getenv("SK"),
	}
)

func getGodaddyDomainURL(domainName, name string) string {
	return fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A/%s", domainName, name)
}

func setDDNSRecord(IP string) error {
	host := getGodaddyDomainURL(domainName, name)
	fmt.Println(host)
	c, err := request(http.MethodGet, host, headers, "")
	if err != nil {
		return err
	}
	oldIP, err := getIp(c)
	if err != nil {
		return err
	}
	fmt.Println("new ip is : ", IP, " dns record ip is:", oldIP)

	if oldIP != IP {
		// 更换 ip
		changeGodaddyDNSRecord(domainName, name, IP)
	}
	return nil
}

func changeGodaddyDNSRecord(domainName, name, IP string) error {
	host := getGodaddyDomainURL(domainName, name)
	putHeaders := headers
	putHeaders["content-type"] = "application/json"
	body := []map[string]interface{}{
		{
			"data": IP,
			"ttl":  3600,
		},
	}
	response, err := request(http.MethodPut, host, putHeaders, body)
	if err != nil {
		return err
	}
	fmt.Println(response)
	return nil
}

func getIp(content string) (string, error) {
	reg, err := regexp.Compile(`([0-9]{1,3}\.){3}[0-9]{1,3}`)
	if err != nil {
		return "", err
	}
	remoteIp := reg.FindString(content)
	return strings.TrimSuffix(remoteIp, "\n"), nil
}

func request(method, host string, headers map[string]string, body interface{}) (string, error) {
	var content []byte

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest(method, host, nil)
	if err != nil {
		return "", err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	if body != nil {
		if v, ok := headers["content-type"]; ok && v == "application/json" {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return "", err
			}
			fmt.Println(string(jsonBody))
			req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBody))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
