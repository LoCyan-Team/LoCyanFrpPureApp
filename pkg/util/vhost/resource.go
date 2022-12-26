// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vhost

import (
	"bytes"
	"io"
	"net/http"
	"os"

	frpLog "github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/version"
)

var NotFoundPagePath = ""

const (
	NotFound = `<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>无法找到你的请求的网站！</title>
		<meta name="description" content="无法找到你的请求的网站!,请确定FRP服务已启动!">
		<link rel="icon" type="image/ico" href="https://tx.hk47.cc/favicon.ico">
		<link rel="stylesheet" href="https://tx.hk47.cc/404-bootstrap.css">
		<link rel="stylesheet" href="https://tx.hk47.cc/404-style.css">
	</head>
	<body>
		<div id="main" class="container">
			<div class="row my-card justify-content-center">
				<div class="col-lg-4 photo-bg"></div>
				<div class="col-lg-8 card">
					<h1><b>对不起, 无法找到你的请求的网站！</b></h1><br>
					<p>找不到您请求的网站，请确定FRP服务已经正常启动! 这个页面来自于 <a href="https://www.locyanfrp.cn/">LoCyanFrp</a></p>
					<p>We can't find the website you requested, please check that your FRP service available</p>
					<p>This page comes from <a href="https://www.locyanfrp.cn/">LoCyanFrp</a></p>
					<hr><script src="https://tenapi.cn/yiyan/?format=js"></script><hr><br>
				</div>
			</div>
		</div>
	<script src="https://tx.hk47.cc/sakura.js"></script>
	</body>
	</html>
`
)

func getNotFoundPageContent() []byte {
	var (
		buf []byte
		err error
	)
	if NotFoundPagePath != "" {
		buf, err = os.ReadFile(NotFoundPagePath)
		if err != nil {
			frpLog.Warn("read custom 404 page error: %v", err)
			buf = []byte(NotFound)
		}
	} else {
		buf = []byte(NotFound)
	}
	return buf
}

func notFoundResponse() *http.Response {
	header := make(http.Header)
	header.Set("server", "frp/"+version.Full()+"-locyanfrp")
	header.Set("Content-Type", "text/html")

	content := getNotFoundPageContent()
	res := &http.Response{
		Status:        "Not Found",
		StatusCode:    404,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(content)),
		ContentLength: int64(len(content)),
	}
	return res
}

func noAuthResponse() *http.Response {
	header := make(map[string][]string)
	header["WWW-Authenticate"] = []string{`Basic realm="Restricted"`}
	res := &http.Response{
		Status:     "401 Not authorized",
		StatusCode: 401,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
	}
	return res
}
