package main

const dashboardHTML = `
<!DOCTYPE html>
<html ng-app="app">
<head>
    <meta charset="utf-8">
    <link rel="icon" href="data:;base64,iVBORw0KGgo=">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/angular.js/1.7.2/angular.min.js"></script>
    <link href="https://fonts.googleapis.com/css?family=Inconsolata:400,700" rel="stylesheet">
    <title>Dashboard</title>
    <style>

    :root {
        --bg: #282c34;
        --list-item-bg: #2c313a;
        --list-item-fg: #abb2bf;
        --list-item-sel-bg: hsl(219, 22%, 25%);
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
        --btn-bg: var(--list-item-bg);
        --btn-hover: var(--list-item-sel-bg);
    }

    * { padding: 0; margin: 0; box-sizing: border-box }

    html, body, .dashboard {
        height: 100%;
        font-family: 'Inconsolata', monospace;
        font-size: 1em;
        font-weight: 400;
        background: var(--bg);
    }

    body { padding: .5rem; }

    div { display: flex; position: relative }

    .list, .req, .res {
        flex: 0 0 37%;
        overflow: auto;
        flex-direction: column;
        padding: .5rem;
    }

    .list, .req { padding-right: .5rem; }
    .req, .res { padding-left: .5rem; }

    .list-inner, .req-inner, .res-inner {
        overflow-x: hidden;
        overflow-y: auto;
        flex: 1;
    }

    .req-inner, .res-inner {
        background: var(--req-res-bg);
    }

    .req, .res {
        color: var(--req-res-fg);
    }

    .list { flex: 0 0 26% }
    .list-inner { flex-direction: column }
    .list-item {
        flex-shrink: 0;
        font-size: 1.2em;
        font-weight: 400;
        height: 50px;
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
        background: var(--list-item-sel-bg);
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
        z-index: 2;
        padding: 1rem;
        width: 100%;
        font-family: inherit;
        font-weight: 400;
        line-height: 1.2em;
    }
    .req:before, .res:before {
        bottom: .5rem;
        left: 1rem;
        font-size: 5em;
        color: var(--bg);
        position: absolute;
        font-weight: 700;
        z-index: 1;
    }
    .req:before {
        content: "↑REQUEST";
    }
    .res:before {
        content: "↓RESPONSE";
    }

    .controls {
        margin-bottom: .5rem;
    }
    .controls > * {
        margin-right: .5rem;
    }
    button {
        background: var(--btn-bg);
        border: 0;
        padding: .5rem 1rem;
        font-size: .75em;
        font-family: inherit;
        color: var(--links);
        cursor: pointer;
        outline: 0;
    }
    button:disabled {
        color: hsl(187, 5%, 50%);
        cursor: default;
    }
    button:hover:enabled {
        background: var(--btn-hover);
    }

    .welcome {
        position: absolute;
        background: rgba(0, 0, 0, .5);
        justify-content: center;
        line-height: 1.5rem;
        z-index: 9;
        color: #fff;
        font-size: 2em;
        top: 50%;
        right: 1rem;
        left: 1rem;
        transform: translate(0%, -50%);
        padding: 3rem;
        box-shadow: 0px 0px 20px 10px rgba(0, 0, 0, 0.1);
    }
    .welcome span {
        font-size: .5em;
        color: #999;
    }
    </style>
</head>
<body>

<div class="dashboard" ng-controller="controller">

    <div class="list">
        <div class="controls">
            <button ng-disabled="items.length == 0" ng-click="clearDashboard()">clear</button>
        </div>
        <div class="list-inner">
            <div class="list-item" ng-repeat="item in items | orderBy: '-id' track by $index" ng-click="show(item)"
                 ng-class="{selected: selectedItem.id == item.id}">
                <span class="method" ng-class="item.method">{{item.method}}</span>
                <span class="path">&lrm;{{item.path}}&lrm;</span>
                <span class="status" ng-class="statusColor(item)">{{item.status == 999 ? 'failed' : item.status}}</span>
            </div>
        </div>
    </div>

    <div class="req">
        <div class="controls">
            <button ng-disabled="!canPrettifyBody('request')" ng-click="prettifyBody('request')">prettify</button>
            <button ng-disabled="selectedItem.id == null" ng-click="copyCurl()">curl</button>
            <button ng-disabled="selectedItem.id == null" ng-click="retry()">retry</button>
        </div>
        <div class="req-inner">
            <pre>{{selectedItem.request}}</pre>
        </div>
    </div>

    <div class="res">
        <div class="controls">
            <button ng-disabled="!canPrettifyBody('response')" ng-click="prettifyBody('response')">prettify</button>
        </div>
        <div class="res-inner">
            <pre ng-class="{error: selectedItem.status == 999}">{{selectedItem.response}}</pre>
        </div>
    </div>

    <div class="welcome" ng-show="items.length == 0">
        <p>Waiting for requests on http://localhost:<<.ProxyPort>>/<br>
        <span>Proxying <<.TargetURL>></span></p>
    </div>

</div>

<script type="text/javascript">
    angular.module('app', [])
        .controller('controller', function($scope, $http) {

            $scope.selectedItem = {};

            $scope.show = item => {
                $scope.selectedItem.id = item.id;
                $scope.selectedItem.status = item.status;
                $http.get(<<.DashboardItemInfoPath>> + item.id).then(r => {
                    $scope.selectedItem.request  = r.data.request;
                    $scope.selectedItem.response = r.data.response;
                    $scope.selectedItem.curl = r.data.curl;
                });
            }

            $scope.statusColor = item => {
                if (item.status < 300) return 'ok';
                if (item.status < 400) return 'warn';
                return 'error';
            }

            $scope.clearDashboard = () => {
                $http.get(<<.DashboardClearPath>>)
                    .then(() => $scope.selectedItem = {});
            }

            $scope.canPrettifyBody = name => {
                if (!$scope.selectedItem[name]) return false;
                return $scope.selectedItem[name].indexOf('Content-Type: application/json') != -1;
            }

            $scope.copyCurl = () => {
                let e = document.createElement('textarea');
                e.value = $scope.selectedItem.curl;
                document.body.appendChild(e);
                e.select();
                document.execCommand('copy');
                document.body.removeChild(e);
            }

            $scope.retry = () => {
                $http.get(<<.DashboardRetryPath>> + $scope.selectedItem.id);
            }

            $scope.prettifyBody = key => {
                let regex = /\n([\{\[](.*\s*)*[\}\]])/;
                let data = $scope.selectedItem[key];
                let match = regex.exec(data);
                let body = match[1];
                let prettyBody = JSON.stringify(JSON.parse(body), null, '    ');
                $scope.selectedItem[key] = data.replace(body, prettyBody);
            }

            const evt = new EventSource(<<.DashboardConnPath>>);
            evt.addEventListener('connected', e => {
                $scope.selectedItem = {};
                $scope.$apply();
            });
            evt.addEventListener('captures', e => {
                $scope.items = JSON.parse(e.data);
                if (!$scope.items.find(i => i.id == $scope.selectedItem.id)) {
                    $scope.selectedItem = {}
                };
                $scope.$apply();
            });
            evt.onerror = () => {
                $scope.items = [];
                $scope.selectedItem = {};
                $scope.$apply();
            };
        });
</script>
</body>
</html>`
