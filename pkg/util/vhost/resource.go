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
	NotFound = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
    <meta name="viewport" content="width=device-width,user-scalable=no,initial-scale=1.0,maximum-scale=1.0,minimum-scale=1.0">
    <title>无法找到您所请求的网站 | LoCyanFrp</title>
	<link rel="icon" href="https://www.locyanfrp.cn/favicon.ico" />
</head>
<body>
	<div class="container">
	    <h1>无法找到您所请求的网站</h1>
        <p>我们无法找到您所请求的网站，请确认 Frp 客户端已正常启动。</p>
        <p>We can not find your website, please confirm that Frp client started.</p>
        <p class="powered-by">Powered by <a target="_blank" href="https://www.locyanfrp.cn">LoCyanFrp</a></p>
	</div>
</body>
<style>
    * {
        margin: 0;
        padding: 0;
        font-family: 'Microsoft YaHei', Arial, sans-serif;
    }
    .container {
        height: 100vh;
        display: flex;
        justify-content: center;
        flex-direction: column;
        align-items: center;
        margin-inline: 0.75rem;
    }
    .container h1 {
        font-weight: 400;
        margin-bottom: 1rem;
    }
    .container p {
        color: gray;
    }
    .container .powered-by {
        margin-top: 2rem;
    }
    .container .powered-by a {
        color: rgb(21, 129, 218);
        text-decoration: none;
        transition: 0.3s;
    }
    .container .powered-by a:hover {
        color: rgb(24, 144, 243);
    }
</style>
<style>
    @media (prefers-color-scheme: dark) {
        html {
            background-color: rgb(41, 41, 41);
        }
        .container h1 {
            color: white;
        }
    }
</style>
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
	header.Set("server", "frp/"+version.Full())
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
