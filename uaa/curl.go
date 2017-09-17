package uaa

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"

	"code.cloudfoundry.org/uaa-cli/utils"
)

type CurlManager struct {
	HttpClient *http.Client
	Config     Config
}

func (cm CurlManager) Curl(path, method, data string, headers []string) (resHeaders, resBody string, err error) {
	target := cm.Config.GetActiveTarget()
	context := target.GetActiveContext()

	url, err := utils.BuildUrl(target.BaseUrl, path)
	if err != nil {
		return
	}

	req, err := http.NewRequest(method, url.String(), strings.NewReader(data))
	if err != nil {
		return
	}
	err = mergeHeaders(req.Header, strings.Join(headers, "\n"))
	if err != nil {
		return
	}
	req, err = addAuthorization(req, context)
	if err != nil {
		return
	}
	addZoneSwitchHeader(req, &cm.Config)

	resp, err := cm.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	headerBytes, _ := httputil.DumpResponse(resp, false)
	resHeaders = string(headerBytes)

	bytes, err := ioutil.ReadAll(resp.Body)
	resBody = string(bytes)

	return
}

func mergeHeaders(destination http.Header, headerString string) (err error) {
	headerString = strings.TrimSpace(headerString)
	headerString += "\n\n"
	headerReader := bufio.NewReader(strings.NewReader(headerString))
	headers, err := textproto.NewReader(headerReader).ReadMIMEHeader()
	if err != nil {
		return
	}

	for key, values := range headers {
		destination.Del(key)
		for _, value := range values {
			destination.Add(key, value)
		}
	}

	return
}
