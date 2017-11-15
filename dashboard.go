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
    <title>Dashboard</title>
    <style>

    * { padding: 0; margin: 0; box-sizing: border-box }

    html, body, .dashboard {
        height: 100%;
        font: 1em verdana, arial, helvetica, sans-serif;
    }

    div { display: flex; position: relative }

    .dashboard { background: #eceff1 }

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
        background: #fefefe;
        padding: 1rem;
    }
    .req-inner { margin: 1rem 0 }

    .list { flex: 0 0 26% }
    .list-inner { flex-direction: column }
    .list-item {
        flex-shrink: 0;
        padding: 1rem;
        color: #767676;
        background: #fefefe;
        cursor: pointer;
        margin-bottom: 0.5rem;
        align-items: center;
    }
    .list-item:hover { background: #eceff1 }
    .list-item, .req-inner, .res-inner {
        box-shadow: 0px 1px 1px 0px rgba(0,0,0,0.25);
    }
    .selected { background: #ff4081 !important; color: #fff }
    .list-item.selected .method,
    .list-item.selected .status { color: #fff  }

    .ok,
    .GET    { color: #88d43f }
    .POST   { color: #ef9c26 }
    .warn,
    .PUT    { color: #4c87dd }
    .PATCH  { color: #767676 }
    .error,
    .DELETE { color: #e53f42 }

    .method { font-size: 0.7em; margin-right: 1rem; padding: .25rem .5rem }
    .status { font-size: 0.8em; padding-left: 1rem }
    .url    { font-size: 0.8em; flex: 1; text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl }

    .req-inner, .res-inner { flex-direction: column  }

    .url-big { flex-shrink: 0; line-height: 1.5em; word-break: break-all; padding-bottom: 1rem; margin-bottom: 1rem; border-bottom: 1px solid #eee }
    .url-big:empty { border: 0 }

    pre { flex: 1; color: #555; word-break: normal; word-wrap: break-word; white-space: pre-wrap; }

    </style>
</head>
<body>

<div class="dashboard" ng-controller="controller">

    <div class="list">
        <div class="list-inner">
            <div class="list-item" ng-repeat="item in items" ng-click="show(item)"
                 ng-class="{selected: selected == item}">
                <span class="method" ng-class="item.method">{{item.method}}</span>
                <span class="url">&lrm;{{item.url}}&lrm;</span>
                <span class="status" ng-class="statusColor(item)">{{item.status}}</span>
            </div>
        </div>
    </div>

    <div class="req">
        <div class="req-inner">
            <div class="url-big">{{url}}</div>
            <pre>{{request}}</pre>
        </div>
    </div>

    <div class="res">
        <div class="res-inner">
            <pre>{{response}}</pre>
        </div>
    </div>

</div>

<script type="text/javascript">
    angular.module('app', [])
        .controller('controller', function($scope) {

            $scope.show = item => {
                $scope.request  = item.request;
                $scope.response = item.response;
                $scope.url = item.url;
                $scope.selected = item;
            }
            $scope.statusColor = item => {
                let status = (item.response.status + '')[0] - 2;
                return ['ok', 'warn', 'error', 'error'][status] || '';
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
