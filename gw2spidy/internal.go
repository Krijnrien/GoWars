package gw2spidy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func (spidy *GW2Spidy) dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, spidy.timeout)
}

func (spidy *GW2Spidy) fetchRawEndpoint(url string) (io io.ReadCloser, err error) {
	var resp *http.Response
	if resp, err = spidy.client.Get(url); err != nil {
		return
	}
	return resp.Body, nil
}

func (spidy *GW2Spidy) fetchEndpoint(ver, datatype string, tag string, id int, spec string, page int, result interface{}) (err error) {
	var endpoint *url.URL
	endpoint, _ = url.Parse("https://www.gw2spidy.com/api/")
	endpoint.Path += ver + "/" + datatype + "/" + tag + "/" + strconv.Itoa(id) + "/" + spec + "/" + strconv.Itoa(page)
	var resp *http.Response
	if resp, err = spidy.client.Get(endpoint.String()); err != nil {
		return err
	}
	var data []byte
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.Unmarshal(data, &result); err != nil {
		var gwerr APIError
		if err = json.Unmarshal(data, &gwerr); err != nil {
			return err
		}
		return fmt.Errorf("Endpoint returned error: %v", gwerr)
	}
	return
}
