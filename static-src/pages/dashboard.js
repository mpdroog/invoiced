"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
var React = require("react");
var axios_1 = require("axios");
var recharts_1 = require("recharts");
var Dashboard = (function (_super) {
    __extends(Dashboard, _super);
    function Dashboard() {
        var _this = _super.call(this) || this;
        _this.state = {
            "metrics": {}
        };
        return _this;
    }
    Dashboard.prototype.render = function () {
        var items = [];
        var sorted = Object.keys(this.state.metrics).sort();
        var pref = "0.00";
        var years = { "2016": 0, "2017": 0 };
        for (var i = 0; i < sorted.length; i++) {
            var key = sorted[i];
            var revenue = this.state.metrics[key].RevenueEx;
            var delta = (((revenue * 100) - (pref * 100)) / 100).toFixed(0);
            var change = {};
            if (delta > 0) {
                change = { backgroundColor: "green", color: "white" };
            }
            items.push(React.createElement("tr", { key: key },
                React.createElement("td", null, key),
                React.createElement("td", null,
                    "\u20AC ",
                    revenue),
                React.createElement("td", { style: change },
                    "\u20AC ",
                    delta)));
            pref = revenue;
            years[key.substr(0, key.indexOf("-"))] += revenue * 100;
        }
        var stats = [];
        for (var i = 0; i < sorted.length; i++) {
            var key = sorted[i];
            var vals = this.state.metrics[key];
            vals.RevenueEx = parseInt(vals.RevenueEx);
            vals.name = key;
            stats.push(vals);
        }
        var smallHead = {
            fontSize: "12px",
            float: "right",
            border: "1px solid gray",
            padding: "10px",
            marginLeft: "5px"
        };
        return React.createElement("div", null,
            React.createElement("div", { className: "normalheader col-md-6" },
                React.createElement("div", { className: "hpanel" },
                    React.createElement("div", { className: "panel-body" },
                        React.createElement("h2", { className: "font-light m-b-xs" },
                            React.createElement("i", { className: "fa fa-bank" }),
                            "Revenue",
                            React.createElement("span", { style: smallHead },
                                "2017: \u20AC ",
                                years[2017] / 100),
                            React.createElement("span", { style: smallHead },
                                "2016: \u20AC ",
                                years[2016] / 100)),
                        React.createElement("table", { className: "table" },
                            React.createElement("thead", null,
                                React.createElement("tr", null,
                                    React.createElement("th", null, "Date"),
                                    React.createElement("th", null, "Revenue"),
                                    React.createElement("th", null, "\u0394"))),
                            React.createElement("tbody", null, items))))),
            React.createElement("div", { className: "normalheader col-md-6" },
                React.createElement("div", { className: "hpanel" },
                    React.createElement("div", { className: "panel-body" },
                        React.createElement("h2", { className: "font-light m-b-xs pa" },
                            React.createElement("i", { className: "fa fa-area-chart" }),
                            "Graph"),
                        React.createElement(recharts_1.LineChart, { width: 600, height: 200, data: stats },
                            React.createElement(recharts_1.XAxis, { dataKey: "name" }),
                            React.createElement(recharts_1.Line, { type: "monotone", dataKey: "RevenueEx", stroke: '#82ca9d', fill: '#82ca9d' }),
                            React.createElement(recharts_1.Tooltip, null))))));
    };
    Dashboard.prototype.componentDidMount = function () {
        this.ajax();
    };
    Dashboard.prototype.ajax = function () {
        var _this = this;
        axios_1.default.get('/api/v1/metrics', {})
            .then(function (res) {
            _this.setState({ metrics: res.data });
        })
            .catch(function (err) {
            handleErr(err);
        });
    };
    return Dashboard;
}(React.Component));
exports.default = Dashboard;
