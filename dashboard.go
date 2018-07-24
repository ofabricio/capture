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
    <link rel="icon" href="data:;base64,iVBORw0KGgo=">
    <script src="https://ajax.googleapis.com/ajax/libs/angularjs/1.6.6/angular.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/2.0.4/socket.io.js"></script>
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,800" rel="stylesheet">
    <title>Dashboard</title>
    <style>

    :root {
        --b00: #fafafa;
        --b01: #f0f0f1;
        --b02: #e5e5e6;
        --b03: #a0a1a7;
        --b04: #696c77;
        --b05: #383a42;
        --b06: #202227;
        --b07: #090a0b;
        --b08: #ca1243;
        --b09: #d75f00;
        --b0A: #c18401;
        --b0B: #50a14f;
        --b0C: #0184bc;
        --b0D: #82aaff;
        --b0E: #c792ea;
        --b0F: #986801;
    }

    * { padding: 0; margin: 0; box-sizing: border-box }

    html, body, .dashboard {
        height: 100%;
        font: 1em 'Open Sans', verdana, sans-serif;
        font-weight: 400;
    }

    div { display: flex; position: relative }

    .dashboard { background: var(--b06) }

    .list, .req, .res {
        flex: 0 0 37%;
        overflow: auto;
    }

    .list-inner, .req-inner, .res-inner{
        margin: 1rem;
        overflow-x: hidden;
        overflow-y: auto;
        flex: 1;
    }
    .req-inner, .res-inner {
        background: var(--b05);
        padding: 1rem;
    }
    .req-inner { margin: 1rem 0 }

    .list { flex: 0 0 26% }
    .list-inner { flex-direction: column }
    .list-item {
        flex-shrink: 0;
        font-weight: 400;
        height: 52px;
        padding: 1rem;
        color: var(--b03);
        background: var(--b07);
        cursor: pointer;
        margin-bottom: .5rem;
        align-items: center;
    }
    .list-item:hover { }
    .list-item, .req-inner, .res-inner {
        box-shadow: 0px 2px 5px 0px rgba(0, 0, 0, 0.1);
    }
    .list-item.selected {
        color: var(--b0D);
        border-right: 1rem solid var(--b0D);
        transition: border .1s linear;
    }

    .ok,
    .GET    { color: var(--b0B) }
    .POST   { color: var(--b09) }
    .warn,
    .PUT    { color: var(--b0A) }
    .PATCH  { color: var(--b04) }
    .error,
    .DELETE { color: var(--b08) }

    .method { font-size: 0.7em; margin-right: 1rem; padding: .25rem .5rem }
    .status { font-size: 0.8em; padding-left: 1rem }
    .path   { font-size: 0.8em; flex: 1; text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl }

    .req-inner, .res-inner { flex-direction: column  }

    pre {
        flex: 1;
        color: var(--b03);
        word-break: normal; word-wrap: break-word; white-space: pre-wrap;
        z-index: 1;
    }
    .req-inner:before, .res-inner:before {
        bottom: 1rem;
        font-size: 4em;
        color: var(--b06);
        position: fixed;
        font-weight: 800;
    }
    .req-inner:before {
        content: "REQUEST";
    }
    .res-inner:before {
        content: "RESPONSE";
    }

    .bt-pretty {
        position: absolute;
        right: .5rem;
        top: 0.25rem;
        font-size: .75em;
        color: var(--b0E);
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
                let regex = /.*\n([\{\[].*[\}\]]).*/gs;
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
