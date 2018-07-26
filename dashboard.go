package main

import (
	"net/http"
)

func getDashboardHandler() http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/html")
		response.Write([]byte(dashboardHTML))
	})
}

const dashboardHTML = `
<!DOCTYPE html>
<html ng-app="app">
<head>
    <meta charset="utf-8">
    <link rel="icon" href="data:;base64,iVBORw0KGgo=">
    <script src="https://ajax.googleapis.com/ajax/libs/angularjs/1.6.6/angular.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/2.0.4/socket.io.js"></script>
    <link href="https://fonts.googleapis.com/css?family=Inconsolata:400,700" rel="stylesheet">
    <title>Dashboard</title>
    <style>

    :root {
        --bg: #282c34;
        --list-item-bg: #2c313a;
        --list-item-fg: #abb2bf;
        --list-item-sel-bg: #61afef;
        --req-res-bg: #2c313a;
        --req-res-fg: #abb2bf;
        --links: #55b5c1;
        --method-get: #98c379;
        --method-post: #c678dd;
        --method-put: #d19a66;
        --method-patch: #a7afbc;
        --method-delete: #e06c75;
        --status-ok: #98c379;
        --status-warn: #d19a66;
        --status-error: #e06c75;
    }

    * { padding: 0; margin: 0; box-sizing: border-box }

    html, body, .dashboard {
        height: 100%;
        font-family: 'Inconsolata', monospace;
        font-size: 1em;
        font-weight: 400;
    }

    div { display: flex; position: relative }

    .dashboard { background: var(--bg) }

    .list, .req, .res {
        flex: 0 0 37%;
        overflow: auto;
    }

    .list-inner, .req-inner, .res-inner {
        margin: 1rem;
        overflow-x: hidden;
        overflow-y: auto;
        flex: 1;
    }
    .req-inner, .res-inner {
        flex-direction: column;
        background: var(--req-res-bg);
        color: var(--req-res-fg);
        padding: 1rem;
        font-size: 1.1em;
    }
    .req-inner { margin: 1rem 0 }

    .list { flex: 0 0 26% }
    .list-inner { flex-direction: column }
    .list-item {
        flex-shrink: 0;
        font-size: 1.2em;
        font-weight: 400;
        height: 52px;
        padding: 1rem;
        background: var(--list-item-bg);
        color: var(--list-item-fg);
        cursor: pointer;
        margin-bottom: .5rem;
        align-items: center;
        transition: background .15s linear;
    }
    .list-item:hover { }
    .list-item, .req-inner, .res-inner {
        box-shadow: 0px 2px 5px 0px rgba(0, 0, 0, 0.1);
    }
    .list-item.selected {
        background: hsl(219, 22%, 25%);
    }

    .GET    { color: var(--method-get) }
    .POST   { color: var(--method-post) }
    .PUT    { color: var(--method-put) }
    .PATCH  { color: var(--method-patch) }
    .DELETE { color: var(--method-delete) }
    .ok     { color: var(--status-ok) }
    .warn   { color: var(--status-warn) }
    .error  { color: var(--status-error) }

    .method { font-size: 0.7em; margin-right: 1rem; padding: .25rem .5rem }
    .status { font-size: 0.8em; padding-left: 1rem }
    .path   { font-size: 0.8em; flex: 1; text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl }

    pre {
        flex: 1;
        word-break: normal; word-wrap: break-word; white-space: pre-wrap;
        z-index: 1;
    }
    .req-inner:before, .res-inner:before {
        bottom: 1rem;
        font-size: 5em;
        color: var(--bg);
        position: fixed;
        font-weight: 700;
    }
    .req-inner:before {
        content: "↑REQUEST";
    }
    .res-inner:before {
        content: "↓RESPONSE";
    }

    .bt-pretty {
        position: absolute;
        right: .5rem;
        top: 0.25rem;
        font-size: .75em;
        color: var(--links);
        text-decoration: none;
    }
    </style>
</head>
<body>

<div class="dashboard" ng-controller="controller">

    <div class="list">
        <div class="list-inner">
            <div class="list-item" ng-repeat="item in items | orderBy: '-id' track by item.id" ng-click="show(item)"
                 ng-class="{selected: isItemSelected(item)}">
                <span class="method" ng-class="item.method">{{item.method}}</span>
                <span class="path">&lrm;{{item.path}}&lrm;</span>
                <span class="status" ng-class="statusColor(item)">{{item.status}}</span>
            </div>
        </div>
    </div>

    <div class="req">
        <div class="req-inner">
        <a ng-show="canPrettifyRequestBody" ng-click="prettifyBody('request')" href="#" class="bt-pretty">prettify</a>
        <pre>{{request}}</pre>
        </div>
    </div>

    <div class="res">
        <div class="res-inner">
            <a ng-show="canPrettifyResponseBody" ng-click="prettifyBody('response')" href="#" class="bt-pretty">prettify</a>
            <pre>{{response}}</pre>
        </div>
    </div>

</div>

<script type="text/javascript">
    angular.module('app', [])
        .controller('controller', function($scope, $http) {

            $scope.show = item => {
                $scope.path = item.path;
                $scope.selectedId = item.id;
                $http.get(item.itemUrl).then(r => {
                    $scope.request  = r.data.request;
                    $scope.response = r.data.response;
                    $scope.canPrettifyRequestBody = r.data.request.indexOf('Content-Type: application/json') != -1;
                    $scope.canPrettifyResponseBody = r.data.response.indexOf('Content-Type: application/json') != -1;
                });
            }

            $scope.statusColor = item => {
                let status = (item.status + '')[0] - 2;
                return ['ok', 'warn', 'error', 'error'][status] || '';
            }

            $scope.isItemSelected = item => {
                return $scope.selectedId == item.id;
            }

            $scope.prettifyBody = key => {
                let regex = /\n([\{\[](.*\s+)*[\}\]])/;
                let data = $scope[key];
                let match = regex.exec(data);
                let body = match[1];
                let prettyBody = JSON.stringify(JSON.parse(body), null, '    ');
                $scope[key] = data.replace(body, prettyBody);
            }

            let socket = io();
            socket.on('connect', () => {
                socket.on('captures', captures => {
                    $scope.items = captures;
                    $scope.$apply();
                });
            });
        });
</script>
</body>
</html>`
