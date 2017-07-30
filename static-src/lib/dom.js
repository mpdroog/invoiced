"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var DOM = (function () {
    function DOM() {
    }
    DOM.eventFilter = function (e, nodeName) {
        if (e.target) {
            var node = e.target;
            while (node !== window && node !== null) {
                if (node.nodeName === nodeName) {
                    return node;
                }
                node = node.parentNode;
            }
        }
        return null;
    };
    return DOM;
}());
exports.DOM = DOM;
